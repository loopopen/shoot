.PHONY: test release

golden:
	go test ./cmd

goldenup:
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
