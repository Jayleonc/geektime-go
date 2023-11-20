package main

import "fmt"

type List interface {
	Add(int)
}

type Product struct {
	Name  string
	Price int
}

func (p *Product) Add(i int) {
	p.Price = p.Price + i
}

func main() {
	p := Product{
		Name:  "Jayleonc",
		Price: 11,
	}

	p.Add(10)
	fmt.Println(p.Price)
}
