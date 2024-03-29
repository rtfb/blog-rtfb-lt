FROM ubuntu:22.04

RUN mkdir -p /app
RUN mkdir -p /host
ADD cmd/migrate-db/migrate-db /app/migrate-db
ADD cmd/reset/passwd-reset /app/passwd-reset
ADD package /app
RUN ln -s /host/rtfblog-dbconf.yml /app/db/dbconf.yml

WORKDIR /app

ENTRYPOINT ["/app/rtfblog"]
