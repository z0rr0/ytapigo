PROGRAM=YtapiGo
BIN=bin/main
VERSION=`bash version.sh`
PKG=github.com/z0rr0/ytapigo
MAIN=github.com/z0rr0/main
SOURCEDIR=src/$(PKG)


all: test

install:
	go install -ldflags "$(VERSION)" $(MAIN)

lint: install
	go vet $(PKG)
	golint $(PKG)
	go vet $(MAIN)
	golint $(MAIN)

test: lint
	# go tool cover -html=coverage.out
	# go tool trace ratest.test trace.out
	go test -race -v -cover -coverprofile=coverage.out -trace trace.out $(PKG)

arm:
	env GOOS=linux GOARCH=arm go install -ldflags "$(VERSION)" $(MAIN)

linux:
	env GOOS=linux GOARCH=amd64 go install -ldflags "$(VERSION)" $(MAIN)

clean:
	rm -rf $(GOPATH)/$(BIN) $(GOPATH)/$(SOURCEDIR)/*.out $(GOPATH)/src/$(MAIN)/*.out
