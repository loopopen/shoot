.PHONY: test release test-modules

MODULES := \
	. \
	examples/constructor-example \
	examples/enumer-example \
	examples/enumer-example2 \
	examples/mapper-example \
	examples/mapper-example2 \
	examples/mapper-example3 \
	examples/mapper-example4 \
	examples/restclient-example

tidy-modules:
	@for dir in $(MODULES); do \
		(cd $$dir && GOWORK=off go mod tidy) || exit 1; \
	done

gen-modules:
	@for dir in $(MODULES); do \
		(cd $$dir && go generate ./...) || exit 1; \
	done

golden-up:
	go test ./cmd -update

test: tidy-modules gen-modules
	cd ./internal && go test ./...
	cd ./cmd/test && go generate ./...
	go test ./cmd

tag:
	@grep -o 'v[^"]*' ./internal/shoot/consts.go

release: test
	@sed "s/= \"v[^\"]*\"/= \"${tag}\"/" ./internal/shoot/consts.go > ./internal/shoot/consts.go.tmp
	@mv ./internal/shoot/consts.go.tmp ./internal/shoot/consts.go
	make gen-modules
	git add -A
	git commit -m"chore: ${tag}"
	git tag ${tag}

push:
	git push && git push --tags

tidy: tidy-modules
# 	docker run --rm -v $(PWD):/app -w /app golang:1.24 sh -c "go mod tidy"
