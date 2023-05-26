package main

import (
	"fmt"
)

// 原型模式
type PrototypeManager struct {
}

var Manager *PrototypeManager

type Type1 struct {
	Name string
}

func (t *Type1) Clone() *Type1 {
	tc := t
	return tc
}

func main() {
	t1 := &Type1{
		Name: "type1",
	}

	t2 := t1.Clone()

	// if t1 == t2 {
	// 	log.Fatal("error!")
	// }

	fmt.Println(&t1, &t2, t1 == t2)
}
