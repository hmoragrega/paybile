FROM golang:1.14 AS builder

ENV SERVICE api
ARG WORKDIR=/go/src/github.com/hmoragrega/$SERVICE

WORKDIR $WORKDIR

COPY . .
RUN go mod download

COPY . .
RUN go build -o $SERVICE ./cmd/service
RUN go test -v -race ./...

EXPOSE 8080
CMD ./$SERVICE

# prod
FROM ubuntu:bionic AS prod

ENV SERVICE api
ARG WORKDIR=/go/src/github.com/hmoragrega/$SERVICE

WORKDIR $WORKDIR

COPY --from=builder $WORKDIR/$SERVICE ./$SERVICE

RUN useradd -ms /bin/bash api
USER api

EXPOSE 8080
ENTRYPOINT ["/bin/bash", "-c", "./$SERVICE"]
