#!/bin/bash -xe

# Ensure dependencies are downloaded
BUILD_NAME=$1

if [[ ! -z "${BUILD_NAME}" ]]; then
    echo "Setting dependencies for build ${BUILD_NAME}..."
    arm dependencies get armoryspinnaker ${BUILD_NAME} armory-io/dinghy  > ./.dinghy-dependencies
fi

if [[ ! -f "./.dinghy-dependencies" ]]; then
    echo "This is not part of a multipipeline build. Run ./bin/dinghyVersion.sh [build you're targeting]"
    exit 1
fi

echo "The following build is configured: $(cat ./.dinghy-dependencies | jq --raw-output -c '.build') based on upstream build $(cat ./.dinghy-dependencies | jq --raw-output -c '.dependencies.dinghy.jobRef')."
echo "See more details at http://build-orchestrator.armory.io/projects/armoryspinnaker/builds/$(cat ./.dinghy-dependencies | jq --raw-output -c '.build.name')"

if [[ -z "$DINGHY_VERSION" ]]; then
    GIT_HASH=`git rev-parse --short HEAD`
    DINGHY_VERSION=$(cat ./dinghy_version)
    PREFIX=$(cat ./.dinghy-dependencies | jq --raw-output -c '.build.buildPrefix')
    if [[ -z "$DINGHY_VERSION" ]]; then
        DINGHY_VERSION=`date +%Y-%m-%d`
    fi
    DINGHY_VERSION="${DINGHY_VERSION}-${GIT_HASH}-${PREFIX}-0000"
fi
export DOCKER_IMAGE="docker.io/armory/dinghy:${DINGHY_VERSION}"