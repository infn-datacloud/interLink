#!/bin/bash

export KUBELET_PORT=10255
export INTERLINKCONFIGPATH=$PWD/kustomizations_tmp/InterLinkConfig.yaml
export CONFIGPATH=$PWD/kustomizations/knoc-cfg-local.json
export NODENAME=test-local-vk 
export VKTOKENFILE=/tmp/token
export VK_CONFIG_PATH=$PWD/kustomizations/knoc-cfg-local.json
export VKUBELET_POD_IP=127.0.0.1
./bin/vk   --nodename test-local-vk \
    --provider knoc \
    --startup-timeout 10s \
    --klog.v "2" \
    --klog.logtostderr --metrics-addr=localhost:10255 --log-level debug