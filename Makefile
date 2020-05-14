.PHONY: default test
all: default test

APPNAME=gosniffer
VERSION=v1.0.0

gosec:
	go get github.com/securego/gosec/cmd/gosec
sec:
	@gosec ./...
	@echo "[OK] Go security check was completed!"

proxy:
	export GOPROXY=https://goproxy.cn

default: proxy
	go fmt ./...&&revive .&&goimports -w .&&golangci-lint run --enable-all&&go install -ldflags="-s -w" ./...

install: proxy
	go install -ldflags="-s -w" ./...

test: proxy
	go test ./...

linux: proxy
	GOOS=linux GOARCH=amd64 go install -ldflags="-s -w" ./...
	upx ~/go/bin/linux_amd64/$(APPNAME)

# https://hub.docker.com/_/golang
# docker run --rm -v "$PWD":/usr/src/myapp -v "$HOME/dockergo":/go -w /usr/src/myapp golang make docker
# docker run --rm -it -v "$PWD":/usr/src/myapp -w /usr/src/myapp golang bash
# 静态连接 glibc
docker:
	docker run --rm -v "$$PWD":/usr/src/myapp -v "$$HOME/dockergo":/go -w /usr/src/myapp golang make static
	ls -lh ~/dockergo/bin/$(APPNAME)
	upx ~/dockergo/bin/$(APPNAME)
	ls -lh ~/dockergo/bin/$(APPNAME)
	mv ~/dockergo/bin/$(APPNAME)  ~/dockergo/bin/$(APPNAME)-$(VERSION)-amd64-glibc2.28
	gzip -f ~/dockergo/bin/$(APPNAME)-$(VERSION)-amd64-glibc2.28
	ls -lh ~/dockergo/bin/$(APPNAME)*

static: proxy
	go install -v -x -a -ldflags '-extldflags "-static" -s -w' ./...
