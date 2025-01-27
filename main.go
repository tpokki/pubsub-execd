package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"log/syslog"
	"os"
	"os/exec"
	"text/template"

	"cloud.google.com/go/pubsub"
)

var logger *log.Logger

func init() {
	flag.String("f", "config.yaml", "path to config file")
}

func main() {
	flag.Parse()

	// Load config
	config := loadConfig(flag.Lookup("f").Value.String())

	// setup logger
	switch config.Logging.Output {
	case "syslog":
		writer, err := syslog.New(syslog.LOG_INFO, "pubsub-execd")
		if err != nil {
			panic(err)
		}
		logger = log.New(writer, "", 0)
	case "none":
		logger = log.New(io.Discard, "", 0)
	case "stderr":
		logger = log.New(os.Stderr, "", log.LstdFlags)
	case "stdout":
		fallthrough
	default:
		logger = log.New(os.Stdout, "", log.LstdFlags)
	}

	ctx, _ := context.WithCancel(context.Background())

	// run triggers
	for _, trigger := range config.Triggers {
		sub, err := subscribe(ctx, trigger.PubSub)
		if err != nil {
			panic(err)
		}

		go run(ctx, trigger.Run, sub)
	}

	// wait for termination signal
	select {
	case <-ctx.Done():
		logger.Print("Terminating")
		os.Exit(0)
	}
}

func run(ctx context.Context, run RunConfig, sub *pubsub.Subscription) {
	// set the max extension to the trigger timeout to avoid message re-delivery
	sub.ReceiveSettings.MaxExtension = run.Timeout

	// set the number of goroutines to the concurrency level
	sub.ReceiveSettings.NumGoroutines = run.Concurrency
	sub.ReceiveSettings.Synchronous = run.Concurrency == 1

	// parser for args template
	tmpl, err := template.New("args").Parse(run.Args.Expression)
	if err != nil {
		panic(err)
	}

	// register the handler to receive messages
	sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		// assume the message payload is a json object and parse it
		var payload map[string]interface{}
		err := json.Unmarshal(msg.Data, &payload)
		if err != nil {
			// ignore, payload wasn't a json object
			payload = make(map[string]interface{})
		}

		// create input context for to parse the args template
		input := struct {
			Attributes map[string]string
			Payload    map[string]interface{}
		}{
			Attributes: msg.Attributes,
			Payload:    payload,
		}

		// parse the args template
		args := new(bytes.Buffer)
		tmpl.Execute(args, input)

		logger.Print(fmt.Sprintf("Running command: %s %s", run.Exec, args.String()))

		// run the command
		cmd := exec.CommandContext(ctx, run.Exec, args.String())
		cmd.Stdout = logger.Writer()
		cmd.Stderr = logger.Writer()

		err = cmd.Run()
		if err != nil {
			msg.Nack()
			logger.Print("non-zero exit value for command", cmd, err)
			return
		}

		logger.Printf("Command %s executed successfully", run.Exec)
		msg.Ack()
	})
}
