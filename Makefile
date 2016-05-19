default: freyr

.PHONY: freyr surtr

pwd := $(shell pwd)

freyr:
	docker run \
		-e CGO_ENABLED=0 \
		-e GOOS=linux \
		-v $(GOPATH)/src/:/go/src/ \
		-v $(pwd):/go/src/github.com/serdmanczyk/freyr/ \
		-w /go/src/github.com/serdmanczyk/freyr/ \
		golang go get ./... && go build -ldflags "-s" -a -installsuffix cgo -o freyr

surtr:
	docker run \
		-e CGO_ENABLED=0 \
		-e GOOS=linux \
		-v $(GOPATH)/src/:/go/src/ \
		-v $(pwd):/go/src/github.com/serdmanczyk/freyr/ \
		-w /go/src/github.com/serdmanczyk/freyr/cmd/surtr \
		golang go get ./... && go build -ldflags "-s" -a -installsuffix cgo -o surtr

rundev:
	docker-compose -f docker-compose.debug.yml -p dev build
	docker-compose -f docker-compose.debug.yml -p dev up --force-recreate

buildstatic:
	docker-compose build

runstatic:
	docker-compose up -d --build

integration:
	docker-compose -p integration down
	docker-compose -p integration build
	docker-compose -f docker-compose.integration.yml -p integration up --force-recreate --abort-on-container-exit

acceptance:
	docker-compose -p acceptance down
	docker-compose -p acceptance build
	docker-compose -f docker-compose.acceptance.yml -p acceptance up --force-recreate --abort-on-container-exit

test: freyr surtr integration acceptance
