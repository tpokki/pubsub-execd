# yellow check mark
YC=\033[0;33m✔︎\033[0m
#green check mark
GC=\033[0;32m✔︎\033[0m
SRC=$(wildcard *.go)

APP=pubsub-execd
PLATFORMS=darwin-arm64 linux-amd64

all: $(APP)

$(APP): $(SRC) bin $(PLATFORMS); @ echo "$(GC) build done" ;

bin:
	@ mkdir -p bin

darwin-arm64:; @ echo "$(YC) building for darwin..." ;
	@ GOOS=darwin  GOARCH=arm64 go build -o bin/$(APP)-$@

linux-amd64:; @ echo "$(YC) building for linux..." ;
	@ GOOS=linux   GOARCH=amd64 go build -o bin/$(APP)-$@k

clean: ; @ echo "$(YC) cleaning..." ;
	@ rm -fr bin/
