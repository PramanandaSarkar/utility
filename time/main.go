package main

import (
	"fmt"
	"time"
)

func main() {
	time := time.Now().Format("02-Jan-2002 15:04:05")
	fmt.Println(time)
}