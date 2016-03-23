FROM golang

RUN apt-get update

RUN go get github.com/dgrijalva/jwt-go
RUN go get golang.org/x/oauth2
RUN go get golang.org/x/oauth2/google
RUN go get github.com/codegangsta/negroni
RUN go get github.com/cyclopsci/apollo
RUN go get github.com/lib/pq

COPY ./ /go/src/github.com/serdmanczyk/gardenspark/
WORKDIR /go/src/github.com/serdmanczyk/gardenspark/
RUN make buildgo

ENTRYPOINT ["/bin/bash"]
