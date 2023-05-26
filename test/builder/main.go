package main

import "fmt"

// 建造者模式
type Builder interface {
	Part1()
	Part2()
	Part3()
}

// 管理类
type Director struct {
	BuilderDirector Builder
}

// 构造函数
func NewDirector(builder Builder) *Director {
	return &Director{
		BuilderDirector: builder,
	}
}

func (d *Director) Construct() {
	d.BuilderDirector.Part1()
	d.BuilderDirector.Part2()
	d.BuilderDirector.Part3()
}

type MyBuilder struct{}

func (m *MyBuilder) Part1() {
	fmt.Println("part1")
}

func (m *MyBuilder) Part2() {
	fmt.Println("part2")
}

func (m *MyBuilder) Part3() {
	fmt.Println("part3")
}

func main() {
	builder := &MyBuilder{}
	director := NewDirector(builder)
	director.Construct()
}
