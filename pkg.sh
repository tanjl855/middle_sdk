#!/bin/bash
#### Setting Environments ####
echo "Setting Environments"
set -xe
export cpwd=`pwd`
output=$cpwd/build

#### Package ####
echo "Setting Package Environments..."
srv_name=tjl-sdk
srv_out=$output/$srv_name
go_path=`go env GOPATH`
go_os=`go env GOOS`
go_arch=`go env GOARCH`

##build normal
echo "Building $srv_name normal executor..."
mkdir -p $srv_out
go build -v -o $srv_out tjl-sdk
mkdir -p $srv_out/conf
##cp...配置文件


###
cd $output
out_tar_name=$srv_name-$go_os-$go_arch
out_tar=$out_tar_name.tar.gz
rm -f $out_tar
tar -czvf $output/$out_tar $srv_name

cd $cpwd

echo "Pac~age $out_tar_name done..."
