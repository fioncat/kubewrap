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

.PHONY: fmt
fmt:
	@find . -name \*.go -exec goimports -w {} \;

.PHONY: check
check:
	@CGO_ENABLED=0 go vet ./...

.PHONY: typos
typos:
	@typos ./
