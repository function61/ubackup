FROM alpine:3.11.3

WORKDIR /

# don't add docker from apk because it's quite expensive - we only need the client.
# statically compiled downloads available here: https://download.docker.com/linux/static/stable/x86_64/
RUN apk add ca-certificates \
	&& mkdir -p /tmp/docker-install \
	&& cd /tmp/docker-install \
	&& wget -O - https://download.docker.com/linux/static/stable/x86_64/docker-18.09.0.tgz | tar -xzf - \
	&& mv docker/docker /usr/bin/docker \
	&& rm -rf /tmp/docker-install

CMD ["ubackup", "scheduler", "run"]

ADD rel/ubackup_linux-amd64 /usr/local/bin/ubackup
