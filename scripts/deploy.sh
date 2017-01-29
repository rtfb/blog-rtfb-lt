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
go build ./cmd/migrate-db/
cp ./migrate-db $package
tar czvf package.tar.gz ./package
rm -rf $package

scp -q scripts/unpack.sh rtfb@rtfb.lt:/home/rtfb/unpack.sh
scp -q package.tar.gz rtfb@rtfb.lt:/home/rtfb/package.tar.gz
rm ./package.tar.gz
full_path=/home/rtfb/package$suffix
ssh rtfb@rtfb.lt "service rtfblog stop"
ssh rtfb@rtfb.lt "/home/rtfb/unpack.sh package$suffix"
ssh rtfb@rtfb.lt "rm $full_path/db/pg/dbconf.yml"
ssh rtfb@rtfb.lt "ln -s /home/rtfb/rtfblog-dbconf.yml $full_path/db/dbconf.yml"
ssh rtfb@rtfb.lt "$full_path/migrate-db --db=$full_path/db --env=$goose_env"
ssh rtfb@rtfb.lt "service rtfblog start"

echo "$env deployed."
