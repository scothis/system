#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

readonly version=$(cat VERSION)
readonly git_sha=$(git rev-parse HEAD)
readonly git_timestamp=$(TZ=UTC git show --quiet --date='format-local:%Y%m%d%H%M%S' --format="%cd")
readonly slug=${version}-${git_timestamp}-${git_sha:0:16}

readonly riff_version=0.6.0-snapshot

source ${FATS_DIR}/.configure.sh

export KO_DOCKER_REPO=$(fats_image_repo '#' | cut -d '#' -f 1 | sed 's|/$||g')
kubectl create ns apps

echo "Installing Cert Manager"
kapp deploy -n apps -a cert-manager -f https://storage.googleapis.com/projectriff/release/${riff_version}/cert-manager.yaml -y

source $FATS_DIR/macros/no-resource-requests.sh

echo "Installing KEDA"
kapp deploy -n apps -a keda -f https://storage.googleapis.com/projectriff/release/${riff_version}/keda.yaml -y

echo "Installing riff Streaming Runtime"
if [ $MODE = "push" ]; then
  kapp deploy -n apps -a riff-streaming-runtime -f https://storage.googleapis.com/projectriff/riff-system/snapshots/riff-streaming-${slug}.yaml -y
elif [ $MODE = "pull_request" ]; then
  ko resolve --strict -f config/riff-streaming.yaml | kapp deploy -n apps -a riff-streaming-runtime -f - -y
fi

if [ $GATEWAY = "kafka" ]; then
  echo "Installing Kafka"
  kapp deploy -n apps -a internal-only-kafka -f https://storage.googleapis.com/projectriff/release/${riff_version}/internal-only-kafka.yaml -y
fi
if [ $GATEWAY = "pulsar" ]; then
  echo "Installing Pulsar"
  kapp deploy -n apps -a internal-only-pulsar -f https://storage.googleapis.com/projectriff/release/${riff_version}/internal-only-pulsar.yaml -y
fi
