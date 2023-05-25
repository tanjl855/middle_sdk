#!/bin/bash
set -e

go mod tidy

./pkg.sh

srv_name=kuxiao-sdk

cd build/$srv_name

./$srv_name