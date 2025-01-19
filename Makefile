SHELL := /bin/bash

ifeq ($(PORT),)
PORT := 8080
endif

all: binary

.PHONY: binary
binary: ## build executable for Linux
	go build -o ./build/pi-collector ./cmd/pi-collector

.PHONY: docker-binary
docker-binary: ## (docker) build executable
	 docker buildx bake binary

.PHONY: tunnel
tunnel: ## SSH to $PI_COL_HOST (sources .env) and create tunnel to telnet API
	source .env; ssh -tt -L 4711:localhost:4711 "root@$${PI_COL_HOST}" </dev/null &
	read

.PHONY: run
run: docker-binary tunnel ## build, setup tunnel and run
	./build/pi-collector

.PHONY: build-dev-image
build-dev-image:
	docker build -t pi-telnet-collector --target ssh-tunnel-collector .

.PHONY: run-docker
run-docker: build-dev-image
	docker run \
		--rm \
		-p 8080:$(PORT) \
		-e SSH_AUTH_SOCK=/ssh-agent.sock \
		--mount type=bind,src="$${SSH_AUTH_SOCK}",target=/ssh-agent.sock \
		--env-file=.env \
		--name pi-collector \
    	-it \
		pi-telnet-collector -p $(PORT)

.PHONY: metrics
metrics: ## curl metrics endpoint for the collector
	curl http://localhost:8080/metrics | less

.PHONY: help
help: ## print this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9_-]+:.*?## / {gsub("\\\\n",sprintf("\n%22c",""), $$2);printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
