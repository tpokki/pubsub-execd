# pubsub-execd

Cron like daemon process that triggers executables based on messages received from pubsub topic.

## Configuration

Define the triggers in configuration file.

```yaml
logging:
  output: stdout
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

`logging` spec:
- `output` - define the log output, accepted values are `none`, `syslog`, `stdout` and `stderr`. Default is `stdout`.

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

## Example

1. Setup topic `execd-demo-topic` with subscription `exec-demo-sub` in your project.
2. Start the daemon:
```sh
pubsub-execd -f sample.yaml
```

3. Send message to topic:
```sh
gcloud pubsub topics publish \
    --project=sample-project-123 execd-demo-topic \
    --message='{"foo": {"bar": "and-this" }, "foobar": "not printed"}' \
    --attribute=foobar=print-this,not=printed
```

4. Expected output from daemon process:
```
2025/01/27 09:08:26 Running command: /bin/echo print-this and-this
print-this and-this
2025/01/27 09:08:26 Command /bin/echo executed successfully
```
