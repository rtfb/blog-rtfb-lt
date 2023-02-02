FROM ubuntu:22.04

RUN apt-get update && apt-get install -y \
    curl \
    git \
    sudo \
    vim \
    wget

RUN mkdir -p /app
RUN mkdir -p /host
ADD cmd/migrate-db/migrate-db /app/migrate-db
ADD cmd/reset/passwd-reset /app/passwd-reset
ADD package /app/package
RUN ln -s /host/rtfblog-dbconf.yml /app/package/db/dbconf.yml

WORKDIR /app

ENTRYPOINT ["/app/package/rtfblog"]
