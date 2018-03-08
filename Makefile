RELEASE=$(shell git describe --always --long --dirty)
PWD := $(shell pwd)
DOCKER_REPO := jaxxstorm/hookpick

build:
	@go build -o hookpick main.go

version:
	@echo $(RELEASE)

build-osx:
	env GOOS=darwin GOARCH=amd64 go build -o hookpick-darwin-amd64 main.go

build-linux:
	env GOOS=linux GOARCH=amd64 go build -o hookpick-linux-amd64 main.go

image:
	docker build -t $(DOCKER_REPO):$(RELEASE) \
		--build-arg GOOS=linux \
		-f Dockerfile .

run:
	docker run --rm -v $(PWD):/root/ $(DOCKER_REPO):$(RELEASE) version

release:
	docker tag $(DOCKER_REPO):$(RELEASE) $(DOCKER_REPO):latest
	docker push $(DOCKER_REPO):$(RELEASE)
	docker push $(DOCKER_REPO):latest
