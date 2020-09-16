#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# TODO restore tests once we have a replacement for function builds
exit 0

source ${FATS_DIR}/.configure.sh

# setup namespace
kubectl create namespace ${NAMESPACE}
fats_create_push_credentials ${NAMESPACE}
source ${FATS_DIR}/macros/create-riff-dev-pod.sh

echo "##[group]Create gateway"
if [ $GATEWAY = "inmemory" ]; then
  riff streaming inmemory-gateway create test --namespace $NAMESPACE --tail
fi
if [ $GATEWAY = "kafka" ]; then
  riff streaming kafka-gateway create test --bootstrap-servers kafka.kafka.svc.cluster.local:9092 --namespace $NAMESPACE --tail
fi
if [ $GATEWAY = "pulsar" ]; then
  riff streaming pulsar-gateway create test --service-url pulsar://pulsar.pulsar.svc.cluster.local:6650 --namespace $NAMESPACE --tail
fi
echo "##[endgroup]"

for test in node ; do
  name=system-${RUNTIME}-fn-uppercase-${test}
  image=$(fats_image_repo ${name})

  echo "##[group]Run function ${name}"

  riff function create ${name} --image ${image} --namespace ${NAMESPACE} --tail \
    --git-repo https://github.com/${FATS_REPO} --git-revision ${FATS_REFSPEC} --sub-path functions/uppercase/${test}

  lower_stream=${name}-lower
  upper_stream=${name}-upper

  riff streaming stream create ${lower_stream} --namespace $NAMESPACE --gateway test --content-type 'text/plain' --tail
  riff streaming stream create ${upper_stream} --namespace $NAMESPACE --gateway test --content-type 'text/plain' --tail

  riff streaming processor create $name --function-ref $name --namespace $NAMESPACE --input ${lower_stream} --output ${upper_stream} --tail

  kubectl exec riff-dev -n $NAMESPACE -- subscribe ${upper_stream} --payload-encoding raw | tee result.txt &
  sleep 10
  kubectl exec riff-dev -n $NAMESPACE -- publish ${lower_stream} --payload "system" --content-type "text/plain"

  actual_data=""
  expected_data="SYSTEM"
  cnt=1
  while [ $cnt -lt 60 ]; do
    echo -n "."
    cnt=$((cnt+1))

    actual_data=$(cat result.txt | jq -r .payload)
    if [ "$actual_data" == "$expected_data" ]; then
      break
    fi

    sleep 1
  done
  fats_assert "$expected_data" "$actual_data"

  kubectl exec riff-dev -n $NAMESPACE -- pkill subscribe

  riff streaming stream delete ${lower_stream} --namespace $NAMESPACE
  riff streaming stream delete ${upper_stream} --namespace $NAMESPACE
  riff streaming processor delete $name --namespace $NAMESPACE

  riff function delete ${name} --namespace ${NAMESPACE}
  fats_delete_image ${image}

  echo "##[endgroup]"
done

riff streaming ${GATEWAY}-gateway delete test --namespace $NAMESPACE
