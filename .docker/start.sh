#!/usr/bin/env sh
set -e

cd /root/

aria2c --conf-path=/conf/aria2.conf
./arias -config /conf/arias.toml