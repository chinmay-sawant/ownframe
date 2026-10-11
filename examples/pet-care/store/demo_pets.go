package store

import "github.com/chinmay-sawant/ownframe/examples/pet-care/model"

var demoPets = []model.Pet{
	{ID: "demo-maple", Name: "Maple", Species: "Dog", Breed: "Golden retriever", BirthDate: "2021-04-12", Notes: "Loves slow morning walks and tennis balls."},
	{ID: "demo-miso", Name: "Miso", Species: "Cat", Breed: "Tabby", BirthDate: "2022-09-03", Notes: "Favorite sunny spot is beside the window."},
	{ID: "demo-pip", Name: "Pip", Species: "Rabbit", Breed: "Mini lop", BirthDate: "2023-02-18", Notes: "Fresh hay and a tidy habitat each day."},
}
