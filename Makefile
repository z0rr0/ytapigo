PROGRAM=YtapiGo
BIN=bin/main
VERSION=`bash version.sh`
PKG=github.com/z0rr0/ytapigo/ytapi
MAIN=main.go
TARGET=yg


all: test

deps:
	go get -u

build: deps
	go build -o $(TARGET) -ldflags "$(VERSION)" $(MAIN)

lint: build
	go vet $(PKG)
	golint $(PKG)
	go vet $(PKG)/cloud
	golint $(PKG)/cloud

test: lint
	# go tool cover -html=coverage.out
	# go tool trace ratest.test trace.out
	go test -race -v -cover -coverprofile=coverage.out -trace trace.out $(PKG)/cloud
	go test -race -v -cover -coverprofile=coverage.out -trace trace.out $(PKG)

travis: build
	go vet $(PKG)
	go vet $(PKG)/cloud
	go test -race -v -cover $(PKG)/cloud

arm:
	env GOOS=linux GOARCH=arm go install -ldflags "$(VERSION)" $(MAIN)

linux:
	env GOOS=linux GOARCH=amd64 go install -ldflags "$(VERSION)" $(MAIN)

clean:
	rm $(TARGET)
