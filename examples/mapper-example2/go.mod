module mapperexample2

go 1.24.6

tool github.com/lopolopen/shoot/cmd/shoot

require github.com/shopspring/decimal v1.4.0

require (
	github.com/lopolopen/shoot v0.0.0 // indirect
	golang.org/x/mod v0.32.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/tools v0.41.0 // indirect
)

replace github.com/lopolopen/shoot => ../..
