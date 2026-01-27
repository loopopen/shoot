module mapperexample

go 1.24.0

tool github.com/lopolopen/shoot/cmd/shoot

require (
	github.com/lopolopen/shoot v0.0.0
	github.com/shopspring/decimal v1.4.0
	gorm.io/gorm v1.31.1
)

require (
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/mod v0.30.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	golang.org/x/tools v0.39.0 // indirect
)

replace github.com/lopolopen/shoot => ../..
