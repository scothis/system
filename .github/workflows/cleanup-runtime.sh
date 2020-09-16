#!/usr/bin/env bash

set -o nounset

source $FATS_DIR/macros/cleanup-user-resources.sh

echo "Cleanup Kafka"
kapp delete -n apps -a internal-only-kafka -y

echo "Cleanup riff Streaming Runtime"
kapp delete -n apps -a riff-streaming-runtime -y

echo "Cleanup KEDA"
kapp delete -n apps -a keda -y

if [ $GATEWAY = "kafka" ]; then
  echo "Cleanup Kafka"
  kapp delete -n apps -a internal-only-kafka -y
fi
if [ $GATEWAY = "pulsar" ]; then
  echo "Cleanup Pulsar"
  kapp delete -n apps -a internal-only-pulsar -y
fi

echo "Cleanup Cert Manager"
kapp delete -n apps -a cert-manager -y
