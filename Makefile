default: freyr

.PHONY: freyr surtr

pwd := $(shell pwd)
contdir := /go/src/github.com/serdmanczyk/freyr/

static_build_prefix := GO15VENDOREXPERIMENT=1 docker run \
		-e CGO_ENABLED=0 \
		-e GOOS=linux \
		-e "GOPATH=$(contdir)vendor:/go" \
		-v $(pwd):$(contdir)

static_build_args := golang go build -ldflags "-s" -a -installsuffix cgo -o

ymlbase := -f docker-compose.yml
ymldebug := -f docker-compose.debug.yml
ymlint := -f docker-compose.integration.yml
ymlacc := -f docker-compose.acceptance.yml
ymlprod := -f docker-compose.prod.yml

filesdebug := $(ymlbase) $(ymldebug) 
filesint := $(ymlbase) $(ymldebug) $(ymlint)
filesacc := $(ymlbase) $(ymldebug) $(ymlacc)
filesprod := $(ymlbase) $(ymlprod)

int_params := go test $$(go list ./... | grep -v vendor) -tags=integration

projdebug := -p debug
projint := -p integration
projacc := -p acceptance

freyr:
	$(static_build_prefix) -w $(contdir) $(static_build_args) freyr

surtr:
	$(static_build_prefix) -w $(contdir)cmd/surtr $(static_build_args) surtr

web:
	cd static && webpack

docker:
	docker-compose $(filesprod) build

debug:
	docker-compose $(filesdebug) $(projdebug) up freyr

integration:
	docker-compose $(filesint) $(projint) run freyr $(int_params)

acceptance:
	docker-compose $(filesacc) $(projacc) run surtr

down: testdown
	docker-compose $(filesdebug) $(projdebug) down

prod:
	docker-compose $(filesprod) up -d

proddown:
	docker-compose $(filesprod) down

testdown:
	docker-compose $(filesint) $(projint) down
	docker-compose $(filesacc) $(projacc) down
	docker ps -a | grep acceptance | awk '{print $$1}' | xargs docker rm

# all: freyr surtr web docker
all: freyr surtr docker
test: integration acceptance testdown
