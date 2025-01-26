package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"text/template"

	"log/syslog"

	"cloud.google.com/go/pubsub"
)

var logger *syslog.Writer

func init() {
	flag.String("config", "config.yaml", "config file")
	var err error
	logger, err = syslog.New(syslog.LOG_INFO, "pubsub-executor")
	if err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()

	// Load config
	config := loadConfig(flag.Lookup("config").Value.String())

	ctx, _ := context.WithCancel(context.Background())

	// run triggers
	for _, trigger := range config.Triggers {
		go run(ctx, trigger)
	}

	// wait for termination signal
	select {
	case <-ctx.Done():
		logger.Info("Terminating")
		os.Exit(0)
	}
}

func run(ctx context.Context, trigger TriggerConfig) {
	sub, err := subscribe(ctx, trigger.PubSub)
	if err != nil {
		panic(err)
	}

	// set the max extension to the trigger timeout to avoid message re-delivery
	sub.ReceiveSettings.MaxExtension = trigger.Run.Timeout

	// set the number of goroutines to the concurrency level
	sub.ReceiveSettings.NumGoroutines = trigger.Run.Concurrency
	sub.ReceiveSettings.Synchronous = trigger.Run.Concurrency == 1

	// parser for args template
	tmpl, err := template.New("args").Parse(trigger.Run.Args.Expression)
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

		logger.Info(fmt.Sprintf("Running command: %s %s", trigger.Run.Exec, args.String()))

		// run the command
		cmd := exec.CommandContext(ctx, trigger.Run.Exec, args.String())
		cmd.Stdout = logger
		cmd.Stderr = logger

		err = cmd.Run()
		if err != nil {
			msg.Nack()
			logger.Err(err.Error())
			return
		}

		msg.Ack()
	})
}
