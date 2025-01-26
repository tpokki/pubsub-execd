package main

import (
	"bytes"
	"context"
	"log"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/pubsub/pstest"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestRun(t *testing.T) {
	capture := new(bytes.Buffer)
	logger = log.New(capture, "unit-test", log.LstdFlags)

	// start pubsub emulator
	ctx, closeCtx := context.WithCancel(context.Background())
	srv := pstest.NewServer()
	defer srv.Close()

	// Connect to the server without using TLS.
	conn, err := grpc.NewClient(srv.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// Use the connection when creating a pubsub client.
	client, err := pubsub.NewClient(ctx, "project", option.WithGRPCConn(conn))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	// Create a topic
	topic, err := client.CreateTopic(ctx, "topic")
	if err != nil {
		t.Fatal(err)
	}
	defer topic.Stop()

	// Create a subscription
	sub, err := client.CreateSubscription(ctx, "sub", pubsub.SubscriptionConfig{
		Topic: topic,
	})
	if err != nil {
		t.Fatal(err)
	}

	// start the test
	rc := RunConfig{
		Timeout:     10,
		Concurrency: 1,
		Exec:        "echo",
		Args: ArgsConfig{
			Expression: "a1={{.Attributes.a1}} a2={{ .Attributes.a2 }} s1={{ .Payload.t1.s1 }} t2={{ .Payload.t2 }} t3={{ .Payload.t3 }}",
		},
	}

	go run(ctx, rc, sub)

	// publish some test messages
	topic.Publish(ctx, &pubsub.Message{
		Data: []byte(`{ "t1": { "s1": 42 }, "t2": true, "t3": "hello" }`),
		Attributes: map[string]string{
			"a1": "foofoo",
			"a2": "foobar",
		},
	})

	time.Sleep(5 * time.Second)
	closeCtx()

	if !strings.Contains(capture.String(), "Command echo executed successfully") {
		t.Error("unexpected output", capture.String())
	}
}
