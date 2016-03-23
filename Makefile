default: buildbinary copybinary buildstatic

buildgo:
	CGO_ENABLED=0 GOOS=linux go build -ldflags "-s" -a -installsuffix cgo -o gardenspark ./

buildbinary:
	docker build -t serdmanczyk/gobuild -f ./build.Dockerfile .

copybinary:
	docker run -d serdmanczyk/gobuild /bin/true
	docker cp $(shell docker ps -q -n=1):/go/src/github.com/serdmanczyk/gardenspark/gardenspark .
	chmod 755 ./gardenspark

buildstatic:
	docker build --rm -t serdmanczyk/gardenspark -f static.Dockerfile .

integrationtest:
	docker-compose -f docker-compose.test.yml -p ci build
	docker-compose -f docker-compose.test.yml -p ci up