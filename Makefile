all: compile img run

compile:
	docker run --rm \
		-v $(shell pwd -P):/go/src/github.com/docker/integreat \
		-v $(shell pwd -P)/build:/go/bin \
		golang:1.7.1-alpine \
		go install github.com/docker/integreat/cmd/integreat

img:
	docker build -t integreat:latest ./build

run:
	docker run -ti --rm -v $(shell pwd -P):/src integreat:latest integreat /src/config/example.yml

clean:
	rm build/integreat


.PHONY: up logs
