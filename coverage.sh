#!/bin/bash

function coverage {
	echo $1":"
	go test -v -covermode=count -coverprofile=cover.tmp $1
	if [ -f cover.tmp ]
    then
        cat cover.tmp | tail -n +2 >> cover.out
        rm cover.tmp
    fi
}

PackageRoot='github.com/griesbacher/nagflux/'

echo "mode: count" > cover.out
for dir in $(find `ls` -type d);
do
if ls $dir/*_test.go &> /dev/null; then
	coverage $PackageRoot$dir
fi
done

#coverage $PackageRoot