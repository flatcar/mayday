# MOD can either be "readonly" or "vendor".
# The default is "vendor" which uses checked out modules for building.
# Use make MOD=readonly to build with sources from the Go module cache instead.
MOD ?= vendor

.PHONY: build
build:
	GO111MODULE=on go build -o bin/mayday -mod=$(MOD)

.PHONY: test
test:
	GO111MODULE=on go test -mod=$(MOD)

.PHONY: vendor
vendor:
	GO111MODULE=on go mod vendor
