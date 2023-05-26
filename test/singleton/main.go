package main

import (
	"fmt"
	"log"
	"sync"
)

// 单例模式
type singleton struct {
	Value int
}

type Singleton interface {
	getValue() int
}

func (s singleton) getValue() int {
	return s.Value
}

var (
	instance *singleton
	once     sync.Once
)

func GetInstance(v int) Singleton {
	once.Do(func() {
		instance = &singleton{Value: v}
	})

	return instance
}

func main() {
	ins1 := GetInstance(32)
	ins2 := GetInstance(2)
	if ins1 != ins2 {
		log.Fatal("error")
	}
	fmt.Println(ins1.getValue(), ins2.getValue(), ins1 == ins2)
}
