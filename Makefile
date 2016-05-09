default: build

pwd := $(shell pwd)

build:
	docker run \
		-e CGO_ENABLED=0 \
		-e GOOS=linux \
		-v $(GOPATH)/src/:/go/src/ \
		-v $(pwd):/go/src/github.com/serdmanczyk/freyr/ \
		-w /go/src/github.com/serdmanczyk/freyr/ \
		golang   go build -ldflags "-s" -a -installsuffix cgo -o freyr

surtr:
	docker run \
		-e CGO_ENABLED=0 \
		-e GOOS=linux \
		-v $(GOPATH)/src/:/go/src/ \
		-v $(pwd):/go/src/github.com/serdmanczyk/freyr/ \
		-w /go/src/github.com/serdmanczyk/freyr/cmd/surtr \
		golang go build -ldflags "-s" -a -installsuffix cgo -o surtr

rundev:
	docker-compose -f docker-compose.debug.yml -p dev build
	docker-compose -f docker-compose.debug.yml -p dev up --force-recreate

buildstatic:
	docker-compose build

runstatic:
	docker-compose up -d

integration:
	docker-compose -f docker-compose.integration.yml -p integration up --force-recreate

acceptance:
	docker-compose -f docker-compose.acceptance.yml -p acceptance down
	docker-compose -f docker-compose.acceptance.yml -p acceptance build --no-cache
	docker-compose -f docker-compose.acceptance.yml -p acceptance up --force-recreate
