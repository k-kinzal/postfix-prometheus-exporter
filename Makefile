USER     := k-kinzal
REPO     := postfix-prometheus-exporter
GIT_TAG  := $(shell git tag --points-at HEAD)
GIT_HASH := $(shell git rev-parse HEAD)
VERSION  := $(shell if [ -n "$(GIT_TAG)" ]; then echo "$(GIT_TAG)"; else echo "$(GIT_HASH)"; fi)

DIST_DIR := $(shell if [ -n "$(GOOS)$(GOARCH)" ]; then echo "./dist/$(GOOS)-$(GOARCH)"; else echo "./dist"; fi)

.PHONY: build
build:
	go build -ldflags "-s -w -X main.version=$(VERSION) -X main.gitCommit=$(GIT_HASH)" -o $(DIST_DIR)/postfix-prometheus-exporter .

.PHONY: cross-build
cross-build:
	@make build GOOS=linux GOARCH=amd64
	@make build GOOS=darwin GOARCH=amd64

.PHONY: package
package: cross-build
	find dist/*/import -type f -exec dirname {} \; | sed 's|^dist/||' | xargs -I{} tar cvzfh dist/{}.tar.gz -C dist/{} import

.PHONY: clean
clean:
	@rm -rf $(DIST_DIR)
