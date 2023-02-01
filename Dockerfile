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

WORKDIR /app

ENTRYPOINT ["/app/package/rtfblog"]
