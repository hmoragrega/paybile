FROM golang:1.14

ENV SERVICE migrator
ARG WORKDIR=/go/src/github.com/hmoragrega/$SERVICE

WORKDIR $WORKDIR

COPY . .
RUN go mod download

RUN go build -o $SERVICE ./cmd/migrator

RUN useradd -ms /bin/bash migrator
USER migrator

ENTRYPOINT ["/bin/bash", "-c", "./$SERVICE"]
