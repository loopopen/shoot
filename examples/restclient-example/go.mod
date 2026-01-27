module restclientexample

go 1.24.6

tool github.com/lopolopen/shoot/cmd/shoot

require github.com/lopolopen/shoot v0.0.0

require (
	golang.org/x/mod v0.26.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/tools v0.35.0 // indirect
)

replace github.com/lopolopen/shoot => ../..
