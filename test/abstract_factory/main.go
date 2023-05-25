package main

import "fmt"

// 抽象工厂模式
type Lunch interface {
	Cook()
}

type Rise struct{}

func (r *Rise) Cook() {
	fmt.Println("it is rise.")
}

type Tomato struct{}

func (t *Tomato) Cook() {
	fmt.Println("it is tomato.")
}

type LunchFactory interface {
	CreateFood() Lunch
	CreateVegetable() Lunch
}

type SimpleLunchFactory struct {
}

func NewSimpleLunchFactory() LunchFactory {
	return &SimpleLunchFactory{}
}

func (s *SimpleLunchFactory) CreateFood() Lunch {
	return &Rise{}
}

func (s *SimpleLunchFactory) CreateVegetable() Lunch {
	return &Tomato{}
}

func main() {
	NewSimpleLunchFactory().CreateFood().Cook()
	NewSimpleLunchFactory().CreateVegetable().Cook()
}
