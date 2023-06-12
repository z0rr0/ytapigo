PROGRAM=YtapiGo
TS=$(shell date -u +"%F_%T")
TAG=$(shell git tag | sort --version-sort | tail -1)
COMMIT=$(shell git log --oneline | head -1)
VERSION=$(firstword $(COMMIT))
FLAG=-X main.Version=$(TAG) -X main.Revision=git:$(VERSION) -X main.BuildDate=$(TS)
TARGET=yg

all: test

build:
	go build -o $(TARGET) -ldflags "$(FLAG)" .

lint: build
	go vet ./...

test: lint
	go test -race -v -cover ./...

arm:
	env GOOS=linux GOARCH=arm go build -o $(TARGET) -ldflags "$(FLAG)" .

linux:
	env GOOS=linux GOARCH=amd64 go build -o $(TARGET) -ldflags "$(FLAG)" .

clean:
	rm $(TARGET)
