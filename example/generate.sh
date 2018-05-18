#!/usr/bin/env bash

mb=100
td=testdata

mkdir out
mkdir testdata

let "c=$mb * 1024"

dd if=/dev/zero of=${td}/file.txt count=${c} bs=1024

openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ${td}/mykey.key -out ${td}/mycert.crt -subj \
    "/C=GB/ST=London/L=London/O=Global Security/OU=IT Department/CN=example.com"