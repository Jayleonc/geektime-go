package service

import "fmt"

type Demo struct {
}

func NewDemo() *Demo {
	return &Demo{}
}

func (d *Demo) Execute() error {
	fmt.Println("执行Demo Execute")
	return nil
}
