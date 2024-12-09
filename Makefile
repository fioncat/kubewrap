COMMIT_ID=$(shell git rev-parse HEAD)
COMMIT_ID_SHORT=$(shell git rev-parse --short HEAD)

TAG=$(shell git describe --tags --abbrev=0 2>/dev/null)

DATE=$(shell date '+%FT%TZ')

# If current commit is tagged, use tag as version, else, use dev-${COMMIT_ID} as version
VERSION=$(shell git tag --points-at ${COMMIT_ID})
VERSION:=$(if $(VERSION),$(VERSION),dev-${COMMIT_ID_SHORT})

.PHONY: build
build:
	@bash build.sh

.PHONY: install
install:
	@bash build.sh "install"

.PHONY: cross
cross:
	@bash build.sh "linux" "amd64"
	@tar -czf bin/kubewrap-linux-amd64.tar.gz -C bin kubewrap
	@bash build.sh "linux" "arm64"
	@tar -czf bin/kubewrap-linux-arm64.tar.gz -C bin kubewrap
	@bash build.sh "darwin" "arm64"
	@tar -czf bin/kubewrap-darwin-arm64.tar.gz -C bin kubewrap
	@bash build.sh "darwin" "amd64"
	@tar -czf bin/kubewrap-darwin-amd64.tar.gz -C bin kubewrap

.PHONY: test
test:
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go test ./...

.PHONY: fmt
fmt:
	@find . -name \*.go -exec goimports -w {} \;

.PHONY: check
check:
	@CGO_ENABLED=0 go vet ./...

.PHONY: typos
typos:
	@typos ./
