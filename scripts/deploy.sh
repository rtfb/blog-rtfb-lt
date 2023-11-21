#!/bin/bash

set -e
set -x

# remote=rtfb@rtfb.lt
remote=rtfb@kertinis.lt
env="staging"
suffix="-$env"

if [ "$1" == "prod" ]; then
    env="production"
    suffix=""
fi

echo "Deploying $env..."

rtfblog_proj="../rtfblog"

pushd $rtfblog_proj
if [[ -d build ]]; then
    rm -r build
fi
make drun
popd

package=./package
rm -r $package
mkdir -p $package
cp -r $rtfblog_proj/build/* $package
if [[ -f $package/server.conf ]]; then
    rm $package/server.conf
fi
if [[ -f $package/server.log ]]; then
    rm $package/server.log
fi
cp -r $rtfblog_proj/db $package
cd cmd/migrate-db
go build
cd ../..

cd cmd/reset
go build
cd ../..

SUFFIX=$suffix make dbuild
SUFFIX=$suffix make dsave

scp rtfblog$suffix.tar $remote:/home/rtfb/
ssh $remote "docker load -i /home/rtfb/rtfblog$suffix.tar"
ssh $remote "sudo systemctl restart rtfblog.service"

echo "$env deployed."
