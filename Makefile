PROGRAM=YtAPIGo
TS=$(shell date -u -Iseconds)
TAG=$(shell git tag | sort --version-sort | tail -1)
COMMIT=$(shell git log --oneline | head -1)
VERSION=$(firstword $(COMMIT))
LDFLAGS=-X main.Version=$(TAG) -X main.Revision=git:$(VERSION) -X main.BuildDate=$(TS)
TARGET=yg

all: test

build:
	go build -o $(TARGET) -ldflags "$(LDFLAGS)" .

lint: build
	go vet ./...

test: lint
	go test -race -cover ./...

bench: lint
	go test -bench=. -benchmem ./...

fuzz: lint
	go test -fuzz=Fuzz -fuzztime=30s github.com/z0rr0/ytapigo/handle

arm:
	env GOOS=linux GOARCH=arm go build -o $(TARGET) -ldflags "$(LDFLAGS)" .

linux:
	env GOOS=linux GOARCH=amd64 go build -o $(TARGET) -ldflags "$(LDFLAGS)" .

clean:
	rm $(TARGET)
