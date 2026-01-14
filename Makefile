.PHONY: test release

golden:
	cd ./cmd/test && go generate ./...
	go test ./cmd

golden-up:
	go test ./cmd -update

generate:
	go generate ./...

test:
	go test ./...

release: test
	sed -i '' "s/= \"v[^\"]*\"/= \"${tag}\"/" ./internal/shoot/consts.go
	git add -A
	git commit -m"chore: ${tag}"
	git tag ${tag}

push:
	git push && git push --tags