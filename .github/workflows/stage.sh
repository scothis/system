#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

readonly version=$(cat VERSION)
readonly git_sha=$(git rev-parse HEAD)
readonly git_timestamp=$(TZ=UTC git show --quiet --date='format-local:%Y%m%d%H%M%S' --format="%cd")
readonly slug=${version}-${git_timestamp}-${git_sha:0:16}

mkdir bin

  echo ""
  echo "# Stage riff System: streaming"
  echo ""
  ko resolve --strict -P -t ${slug} -f config/riff-streaming.yaml > bin/riff-streaming.yaml
  gsutil cp bin/riff-streaming.yaml gs://projectriff/riff-system/snapshots/riff-streaming-${slug}.yaml
