package main

import (
	"fmt"
	"mapperexample/app/dto"
	"mapperexample/domain/model"
)

func main() {
	// orderDto := dto.NewOrder("1", "99", enums.OrderStatusCompleted)
	// order := orderDto.ToModel()
	// fmt.Println("Domain model.Order: ", order)

	// orderDto2 := new(dto.Order).FromModel(order)
	// fmt.Println("DTO dto.Order: ", orderDto2)

	// userDto := dto.NewUser("1", "Zhen Chen", "zhen@chen.com")
	// user := userDto.ToModel()
	// fmt.Println("Domain model.User: ", user)

	// userDto2 := &dto.User{}
	// userDto2.FromModel(user)
	// fmt.Println("DTO dto.User: ", userDto2)

	// o := dto.Order{}
	// x := o.ToDomain()
	// fmt.Println(x)

	o := model.Order{}
	x := new(dto.Order).FromDomain(&o)
	fmt.Println(x)
}
