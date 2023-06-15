#!/bin/bash
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

pushd $SCRIPT_DIR > /dev/null

helm uninstall arkime -n arkime
helm package chart
helm install arkime arkime3-viewer-2.0.0.tgz -n arkime

popd > /dev/null