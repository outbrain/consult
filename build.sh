#!/bin/bash

function bump_version() {
	echo "$1" | awk -F '.' '{$3=$3+1; print $1 "." $2 "." $3 "-dev"}'
}

last_tag="$(git describe --tags --abbrev=0)"

if git describe --tags --exact-match &>/dev/null; then
	version="${last_tag}"
else
	version="$(bump_version ${last_tag})"
fi

echo "Building version ${version}"

if [[ "$1" == "dev" ]]; then
	go build -ldflags "-X main.version=${version}"
else
	gox -osarch='linux/amd64 darwin/amd64' -ldflags "-X main.version=${version}" -output='build/{{.Dir}}-{{.OS}}-{{.Arch}}'
fi