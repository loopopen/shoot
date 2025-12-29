package main

import (
	"constructorsample/model"
	"encoding/json"
	"fmt"

	"github.com/lopolopen/shoot"
)

func main() {
	user := model.NewUser("11", "Tom", 0, 21, "123456")
	fmt.Printf("%+v\n", user)

	user.SetAge(31)
	user.SetTel("999999")

	userJson, err := json.Marshal(user)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(userJson))

	var user2 model.User
	err = json.Unmarshal(userJson, &user2)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", &user2)

	book := model.NewBook("Love")
	book.SetWriters([]string{"Bill"})
	book.Remarks = "comming soon!"
	book.SetOwner(user)

	bookJson, err := json.Marshal(book)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(bookJson))

	conf := shoot.NewWith(
		model.NameOfConf("name"),
		model.HostOfConf([]string{"host"}),
	)
	fmt.Println(conf)
}
