package main

import "fmt"

var T string

func main() {
	fmt.Println(T)
}

func init() {
	T = "hello"
}
