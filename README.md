# pubsub-execd

Cron like daemon process that triggers executables based on messages received from pubsub topic.

## Configuration

Define the triggers in configuration file.

```yaml
triggers:
  - name: echo stuff
    run:
      timeout: 1h
      concurrency: 1
      exec: /bin/echo
      args:
        expression: {{ .Attributes.foobar }} {{ .Payload.foo.bar }}
    pubsub:
      project: sample123
      subscription: execd
  - name: date with format
    run:
      timeout: 20s
      concurrency: 1
      exec: date
      args:
        expression: +{{ .Attributes.format }}
    pubsub:
      project: sample123
      subscription: execd
```

`run` spec:
- `timeout` - if the message is not ack in this time, it will be redelivered to subscription
- `concurrency` - for receiving and processing messages for same trigger in parallel
- `exec` - executable
- `args.expression` - go-template to parse the received message. The attributes are stored as `.Attributes` and if the received data is in `json` format, it is stored in `.Payload`.

`pubsub` spec:
- `project` - project id from where subscription should be found
- `subscription` - subscription name

*n.b* `pubsub-execd` does not create subscriptions, and they must exist before using them in configuration.

## Running

```sh
pubsub-execd -f /path/to/config.yaml
```
