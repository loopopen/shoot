.PHONY: test release

golden:
	cd ./cmd/test && go generate ./...
	go test ./cmd

golden-up:
	go test ./cmd -update

test:
	cd ./internal && go test ./...

release: test golden
	sed -i '' "s/= \"v[^\"]*\"/= \"${tag}\"/" ./internal/shoot/consts.go
	git add -A
	git commit -m"chore: ${tag}"
	git tag ${tag}

push:
	git push && git push --tags

# tidy:
# 	podman run --rm -v $(PWD):/app -w /app golang:1.24 sh -c "go mod tidy"

gen-all-x:
	cd ./examples/constructor-example && go generate ./...
	cd ./examples/enumer-example && go generate ./...
	cd ./examples/restclient-example && go generate ./...
	cd ./examples/mapper-example && go generate ./...
	cd ./examples/mapper-example2 && go generate ./...
	cd ./examples/mapper-example3 && go generate ./...
	