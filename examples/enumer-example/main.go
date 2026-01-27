package main

import (
	"encoding/json"
	"enumerexample/enums"
	"fmt"

	"github.com/lopolopen/shoot"
)

type X struct {
	Name  string
	Light enums.Light
}

func main() {
	fmt.Printf("LightRed: %s\n", enums.LightRed)

	x := X{
		Name:  "x",
		Light: enums.LightGreen,
	}
	xJson, _ := json.Marshal(x)
	fmt.Println(string(xJson))

	var x2 X
	json.Unmarshal(xJson, &x2)
	fmt.Println(x2)

	l, _ := shoot.ParseEnum[enums.Light]("Red")
	fmt.Println(l)

	if !shoot.IsEnum[enums.Light](5) {
		fmt.Println("not")
	}

	if shoot.TryParseEnum("Green", &l) {
		fmt.Println(l)
	}
}
