#!/usr/bin/env bash
set -euo pipefail
export LC_ALL=C

if [[ $# != 1 ]]; then
  echo "Usage: release-version.sh VERSION" >&2
  exit 2
fi

version="${1#v}"
number='(0|[1-9][0-9]*)'
identifier='(0|[1-9][0-9]*|[0-9]*[A-Za-z-][0-9A-Za-z-]*)'
build='[0-9A-Za-z-]+'
pattern="^${number}[.]${number}[.]${number}(-${identifier}([.]${identifier})*)?([+]${build}([.]${build})*)?$"
if [[ ! "$version" =~ $pattern ]]; then
  echo "Invalid release version: $1" >&2
  exit 1
fi

prerelease=false
if [[ "${version%%+*}" == *-* ]]; then
  prerelease=true
fi
printf 'version=%s\ntag=v%s\nprerelease=%s\n' "$version" "$version" "$prerelease"
