#!/bin/sh

for file in $*
do
    echo $file
    curl "http://localhost:8080/queue/add?file="$file
done
