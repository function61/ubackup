FROM alpine:latest

WORKDIR /

RUN apk add --no-cache ca-certificates docker-cli

CMD ["ubackup", "scheduler", "run"]

ADD rel/ubackup_linux-amd64 /usr/local/bin/ubackup
