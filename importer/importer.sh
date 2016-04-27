#!/bin/bash
touch $1
> $1
robots=(
    "github.com/trinchan/slackbot/robots/raffl"
)

echo "package importer

import (" >> $1

for robot in "${robots[@]}"
do
    echo "    _ \"$robot\" // automatically generated import to register bot, do not change" >> $1
done
echo ")" >> $1

gofmt -w -s $1
