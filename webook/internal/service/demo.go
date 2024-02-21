package service

type Demo struct {
}

func NewDemo() *Demo {
	return &Demo{}
}

func (d *Demo) Execute() error {

	return nil
}
