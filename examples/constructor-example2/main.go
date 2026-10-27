package main

import (
	"fmt"

	"constructorexample2/model"
)

// Test case only, not the recommended way to run shoot.
// go generate normally runs inside the package. Here the command in
// model/user.go switches to this directory and passes [dir]=./model,
// so the process working directory and the package directory differ.
// Run from this directory: go generate ./model

func main() {
	user := model.NewUser("1", "Tom")
	order := model.NewOrder("o-1", 20)
	fmt.Println(user.Name(), order.Id())
}

//panic: go:generate go tool shoot new -getset -type=* ./model
