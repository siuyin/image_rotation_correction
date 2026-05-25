#!/bin/bash
set -e
mkdir -p /dist/lib64
cp /app/build/roll_corr /dist/roll_corr
cp /lib64/ld-linux-x86-64.so.2 /dist/lib64/
for lib in $(ldd /app/build/roll_corr | grep "=> /" | awk '{print $3}'); do
    dir=$(dirname "$lib")
    mkdir -p "/dist$dir"
    cp "$lib" "/dist$lib"
done
