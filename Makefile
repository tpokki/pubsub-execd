# yellow check mark
YC=\033[0;33m✔︎\033[0m
#green check mark
GC=\033[0;32m✔︎\033[0m
SRC=$(wildcard *.go)

APP=pubsub-execd
PLATFORMS=darwin-arm64 linux-amd64

all: test $(APP)

test:; @ echo "$(YC) running tests..." ;
	@ go test -v ./...

$(APP): $(SRC) $(PLATFORMS); @ echo "$(GC) build done" ;

bin:
	@ mkdir -p bin

darwin-arm64: bin/$(APP)-darwin-arm64 ; @ echo "$(YC) building for darwin..." ;

bin/$(APP)-darwin-arm64: bin $(SRC)
	@ GOOS=darwin  GOARCH=arm64 go build -o $@

linux-amd64: bin/$(APP)-linux-amd64 ; @ echo "$(YC) building for linux..." ;

bin/$(APP)-linux-amd64: bin $(SRC)
	@ GOOS=linux   GOARCH=amd64 go build -o $@

clean: ; @ echo "$(YC) cleaning..." ;
	@ rm -fr bin/
