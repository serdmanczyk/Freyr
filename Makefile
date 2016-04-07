default: build

pwd := $(shell pwd)

build:
	docker run \
		-e CGO_ENABLED=0 \
		-e GOOS=linux \
		-v $(GOPATH)/src/:/go/src/ \
		-v $(pwd):/go/src/github.com/serdmanczyk/gardenspark/ \
		-w /go/src/github.com/serdmanczyk/gardenspark/ \
		golang go build -ldflags "-s" -a -installsuffix cgo -o gardenspark

test:
	docker-compose -f docker-compose.test.yml -p ci up

rundev:
	docker-compose -f docker-compose.debug.yml -p dev up

runstatic:
	docker-compose build
	docker-compose up
