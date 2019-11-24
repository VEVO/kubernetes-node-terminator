# setting some defaults if those variables are empty
OWNER=vevo
APP_NAME=terminator
IMAGE_NAME=$(OWNER)/$(APP_NAME)
GO_REVISION?=$(shell git rev-parse HEAD)
GO_TO_REVISION?=$(GO_REVISION)
GO_FROM_REVISION?=$(shell git rev-parse refs/remotes/origin/master)
GIT_TAG=$(IMAGE_NAME):$(GO_REVISION)
BUILD_VERSION?=$(shell date +%Y%m%d%H%M%S)-dev
BUILD_TAG=$(IMAGE_NAME):$(BUILD_VERSION)
LATEST_TAG=$(IMAGE_NAME):latest

PHONY: go-build

docker-lint:
	docker run -it --rm -v "${PWD}/Dockerfile":/Dockerfile:ro redcoolbeans/dockerlint

docker-login:
	@docker login -u "$(DOCKER_USER)" -p "$(DOCKER_PASS)"

go-dep:
	go get -d -t -v ./...


go-fmt:
	gofmt -s -w .

go-lint: go-fmt
	go get -u golang.org/x/lint/golint
	golint -set_exit_status ./...
	go vet -v ./...

go-test:
	go test -v ./...

go-build: go-dep go-lint go-test
	@go build -v -a -ldflags "-X main.version=$(BUILD_VERSION)"

build: go-build

release:
	git tag -s $(BUILD_VERSION) -m "Release $(BUILD_VERSION)"
	goreleaser release --rm-dist

# vim: ft=make
