#!/bin/bash

GIT_DESC=$(git describe --tags 2> /dev/null)
GIT_TAG=$(git describe --tags --abbrev=0 2> /dev/null)
GIT_COMMIT=$(git rev-parse HEAD)
GIT_COMMIT_SHORT=$(git rev-parse --short HEAD)

if [[ "$GIT_DESC" == "$GIT_TAG" ]]; then
	BUILD_TYPE="stable"
	BUILD_VERSION="$GIT_TAG"
else
	BUILD_TYPE="dev"
	BUILD_VERSION="${GIT_TAG}-dev_${GIT_COMMIT_SHORT}"
fi

if [[ -z "$BUILD_VERSION" ]]; then
	BUILD_TYPE="dev"
	BUILD_VERSION="dev_${GIT_COMMIT_SHORT}"
fi

if git status --porcelain | grep -E '(M|A|D|R|\?)' > /dev/null; then
	BUILD_TYPE="dev-uncommitted"
	BUILD_VERSION="${BUILD_VERSION}-uncommitted"
fi

cat << EOF
Build Args:
GIT_DESC=${GIT_DESC}
GIT_TAG=${GIT_TAG}
GIT_COMMIT=${GIT_COMMIT}
GIT_COMMIT_SHORT=${GIT_COMMIT_SHORT}
BUILD_TYPE=${BUILD_TYPE}
BUILD_VERSION=${BUILD_VERSION}
EOF

BUILD_FLAGS="-X main.Version=${BUILD_VERSION} -X main.BuildType=${BUILD_TYPE} -X main.BuildCommit=${GIT_COMMIT} -X main.BuildTime=$(date +%F-%Z/%T)"

echo ""
echo "Build kubewrap..."
CGO_ENABLED=0 GOOS="$1" GOARCH="$2" go build -ldflags "${BUILD_FLAGS}" -o ./bin/kubewrap
if [[ $? -ne 0 ]]; then
	echo "Build kubewrap failed"
	exit 1
fi
