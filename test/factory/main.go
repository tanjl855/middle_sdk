package main

import "fmt"

// 工厂模式
type Factory interface {
	Proc()
}

type Tanxiao struct {
	Name string
}

func (t *Tanxiao) Proc() {
	fmt.Println("Tanxiao proc here!")
}

type Dongdong struct {
	Name string
}

func (t *Dongdong) Proc() {
	fmt.Println("Dongdong proc here!")
}

func NewFactory(factoryName string) Factory {
	if factoryName == "" {
		return nil
	}
	switch factoryName {
	case "tanxiao":
		return &Tanxiao{Name: factoryName}
	case "dongdong":
		return &Dongdong{Name: factoryName}
	}
	return nil
}

func main() {
	NewFactory("tanxiao").Proc()
	NewFactory("dongdong").Proc()
}
