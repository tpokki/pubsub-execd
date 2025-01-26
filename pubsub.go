package main

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub"
)

func subscribe(ctx context.Context, config PubSubConfig) (*pubsub.Subscription, error) {
	client, err := pubsub.NewClient(ctx, config.project)
	if err != nil {
		return nil, err
	}

	sub := client.Subscription(config.Subscription)
	if ok, err := sub.Exists(ctx); err != nil {
		return nil, err
	} else if !ok {
		return nil, fmt.Errorf("subscription %s does not exist", config.Subscription)
	}
	return sub, nil
}
