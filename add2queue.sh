#!/bin/sh

for file in "$@"; do 
    echo $file
    f=`echo $file | sed 's/%/%25/g' | sed 's/ /%20/g'`
    curl "http://localhost:58080/queue/add?file="$f
done
