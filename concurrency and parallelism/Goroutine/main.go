package main

import (
	"fmt"
	"time"
)

func main(){
	result := Foo(1)
	fmt.Println("The Result:", result)
	//go Foo()
	// the same result
	// go func(){
	// 	Foo()
	// }()
}

func Foo(n int) string{
	time.Sleep(time.Second * 2)
	return fmt.Sprintf("result %d", n)
}