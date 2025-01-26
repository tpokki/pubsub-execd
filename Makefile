all: pubsub-execd

clean:
	rm -f pubsub-execd

pubsub-execd: $(shell find . -name '*.go')
	go build -o pubsub-execd