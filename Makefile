TAG ?=
SKIP_MODULES ?=

.PHONY: tidy tidy-modules gen-modules lint lint-modules test-modules test golden-up tag check-tag release push

# Directories that contain a go.mod. go.work is local and is not consulted.
ALL_MODULES := $(patsubst %/go.mod,%,$(shell find . -name go.mod -not -path '*/vendor/*' | sort))
MODULES := $(filter-out $(SKIP_MODULES),$(ALL_MODULES))

define run-in-modules
	@set -e; \
	for dir in $(MODULES); do \
		printf '%s\n' "==> $$dir"; \
		(cd "$$dir" && $(1)); \
	done
endef

tidy-modules:
	$(call run-in-modules,GOWORK=off go mod tidy)

gen-modules:
	$(call run-in-modules,GOWORK=off go generate ./...)

lint-modules:
	$(call run-in-modules,GOWORK=off go vet ./...)

test-modules:
	$(call run-in-modules,GOWORK=off go test ./...)

golden-up:
	go test ./cmd -update

test: tidy-modules gen-modules test-modules

lint: lint-modules

tag:
	@grep -o 'v[^"]*' ./internal/shoot/consts.go

check-tag:
	@test -n "$(TAG)" || { echo "TAG is required (for example: make release TAG=v1.2.3)" >&2; exit 2; }

release: check-tag
	$(MAKE) test
	@sed "s/= \"v[^\"]*\"/= \"$(TAG)\"/" ./internal/shoot/consts.go > ./internal/shoot/consts.go.tmp
	@mv ./internal/shoot/consts.go.tmp ./internal/shoot/consts.go
	$(MAKE) gen-modules
	git add -A
	git commit -m "chore: $(TAG)"
	git tag $(TAG)

push:
	git push && git push --tags

tidy: tidy-modules
