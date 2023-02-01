#!/bin/bash

env="staging"
suffix="-$env"

if [ "$1" == "prod" ]; then
    env="production"
    suffix=""
fi

goose_env=$env
echo "Deploying $env..."

rtfblog_proj="../rtfblog"

killall rtfblog
pushd $rtfblog_proj
make all
popd

package=./package
mkdir -p $package
cp -r $rtfblog_proj/build/* $package
rm $package/server.conf
rm $package/server.log
cp -r $rtfblog_proj/db $package
cp ./stuff/images/* $package/static/
cp ./testdata/rtfblog-dump.sql $package/rtfblog-dump.sql
cd cmd/migrate-db
go build
cd ../..

cd cmd/reset
go build
cd ../..

# remote=rtfb@rtfb.lt
remote=rtfb@kertinis.lt

SUFFIX=$suffix make dbuild
SUFFIX=$suffix make dsave

scp -q rtfblog$suffix.tar $remote:/home/rtfb/
ssh $remote "service rtfblog$suffix stop"
ssh $remote "docker load -i /home/rtfb/rtfblog$suffix.tar"
ssh $remote "service rtfblog$suffix start"

echo "$env deployed."
