#!/bin/bash

PackageRoot='github.com/griesbacher/nagflux/'

echo "mode: count" > cover.out
for dir in $(find `ls` -type d);
do
if ls $dir/*_test.go &> /dev/null; then
	echo $dir
	go test -v -covermode=count -coverprofile=cover.tmp $PackageRoot$dir
	if [ -f cover.tmp ]
    then
        cat cover.tmp | tail -n +2 >> cover.out
        rm cover.tmp
    fi
fi
done

