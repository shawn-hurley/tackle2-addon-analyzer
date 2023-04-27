GOBIN ?= ${GOPATH}/bin
IMG   ?= quay.io/konveyor/tackle2-addon-windup:latest

all: cmd

fmt:
	go fmt ./...

vet:
	go vet ./...

docker-build:
	docker build -t ${IMG} .

podman-build:
	podman build -t ${IMG} .

cmd: fmt vet
	go build -ldflags="-w -s" -o bin/addon github.com/konveyor/tackle2-addon-windup/cmd

.PHONY: start-minikube
START_MINIKUBE_SH = ./bin/start-minikube.sh
start-minikube:
ifeq (,$(wildcard $(START_MINIKUBE_SH)))
	@{ \
	set -e ;\
	mkdir -p $(dir $(START_MINIKUBE_SH)) ;\
	curl -sSLo $(START_MINIKUBE_SH) https://raw.githubusercontent.com/konveyor/tackle2-operator/main/hack/start-minikube.sh ;\
	chmod +x $(START_MINIKUBE_SH) ;\
	}
endif
	$(START_MINIKUBE_SH);

.PHONY: install-tackle
INSTALL_TACKLE_SH = ./bin/install-tackle.sh
install-tackle:
ifeq (,$(wildcard $(INSTALL_TACKLE_SH)))
	@{ \
	set -e ;\
	mkdir -p $(dir $(INSTALL_TACKLE_SH)) ;\
	curl -sSLo $(INSTALL_TACKLE_SH) https://raw.githubusercontent.com/konveyor/tackle2-operator/main/hack/install-tackle.sh ;\
	chmod +x $(INSTALL_TACKLE_SH) ;\
	}
endif
	export TACKLE_ADDON_WINDUP_IMAGE=$(IMG); \
	export TACKLE_IMAGE_PULL_POLICY='IfNotPresent'; \
	$(INSTALL_TACKLE_SH);

.PHONY: test-e2e
test-e2e:
	bash hack/test-e2e.sh
