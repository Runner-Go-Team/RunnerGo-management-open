package main

import (
	"fmt"
)

func square() func() int { //返回一个自己定义的函数类型
	var x int = 0
	fmt.Println("外部x：", x)
	return func() int {
		fmt.Println("内部x：", x)

		x++
		fmt.Println(&x)
		return x
	}
}
func main() {
	//f := square()
	////square()
	//fmt.Println(f())
	////square()
	//fmt.Println(f())

	var p *int
	fmt.Printf("%T\n", p)

}
