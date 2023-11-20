package main

type ListV1[T any] interface {
	Add(index int, val T) error
	Append(val T) error
	Delete(index int) error
}

type Node[T any] struct {
	Data T
}

func (n Node[T]) Add(int, T) {

}

func Use() {
	n := Node[int]{}
	n.Add(1, 123)

}
