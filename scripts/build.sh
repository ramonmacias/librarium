#!/bin/sh
srcPath="cmd"
pkgFile="main.go"
app="librarium"
src="$srcPath/$app/$pkgFile"

printf "\nStart running: $app\n"
export $(grep -v '^#' config/.env | xargs) && time go run $src
unset $(grep -v '^#' config/.env | sed -E 's/(.*)=.*/\1/' | xargs)
printf "\nStopped running: $app\n\n"
