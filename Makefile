GOBIN ?= ${GOPATH}/bin
IMG   ?= quay.io/konveyor/tackle2-addon-analyzer:latest

all: cmd

fmt:
	go fmt ./...

vet:
	go vet ./...

image-docker:
	docker build -t ${IMG} .

image-podman:
	podman build -t ${IMG} .

cmd: fmt vet
	go build -ldflags="-w -s" -o bin/addon github.com/konveyor/tackle2-addon-analyzer/cmd
