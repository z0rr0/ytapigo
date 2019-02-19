PROGRAM=YtapiGo
BIN=bin/main
VERSION=`bash version.sh`
PKG=github.com/z0rr0/ytapigo
MAIN=main.go
SOURCEDIR=src/$(PKG)
TARGET=yg


all: test


build:
	go build -o $(TARGET) -ldflags "$(VERSION)" $(MAIN)

lint: build
	go vet $(PKG)
	golint $(PKG)

test: lint
	# go tool cover -html=coverage.out
	# go tool trace ratest.test trace.out
	go test -race -v -cover -coverprofile=coverage.out -trace trace.out $(PKG)

arm:
	env GOOS=linux GOARCH=arm go install -ldflags "$(VERSION)" $(MAIN)

linux:
	env GOOS=linux GOARCH=amd64 go install -ldflags "$(VERSION)" $(MAIN)

clean:
	rm $(TARGET)
