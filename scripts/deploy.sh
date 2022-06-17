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
cp -r $rtfblog_proj/build $package
rm $package/server.conf
rm $package/server.log
cp -r $rtfblog_proj/db $package
cp ./stuff/images/* $package/static/
cp ./testdata/rtfblog-dump.sql $package/rtfblog-dump.sql
cd cmd/migrate-db
go build
cd ../..
cp cmd/migrate-db/migrate-db $package
tar czvf package.tar.gz ./package
rm -rf $package

remote=rtfb@rtfb.lt

scp -q scripts/unpack.sh $remote:/home/rtfb/unpack.sh
scp -q package.tar.gz $remote:/home/rtfb/package.tar.gz
rm ./package.tar.gz
full_path=/home/rtfb/package$suffix
ssh $remote "service rtfblog stop"
ssh $remote "/home/rtfb/unpack.sh package$suffix"
ssh $remote "rm $full_path/db/pg/dbconf.yml"
ssh $remote "ln -s /home/rtfb/rtfblog-dbconf.yml $full_path/db/dbconf.yml"
ssh $remote "$full_path/migrate-db --db=$full_path/db --env=$goose_env"
ssh $remote "service rtfblog start"

echo "$env deployed."
