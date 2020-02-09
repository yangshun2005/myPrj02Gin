package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var cliName = flag.String("name", "nick", "Input Your Name")

	//fmt.Fprintf(os.Stderr, *cliName)
	flag.PrintDefaults()
	fmt.Printf(*cliName + "\n")

	arg := os.Args[1:]

	if arg[1] == *cliName {
		fmt.Println("ok")
	} else {
		fmt.Println(*cliName)
	}
	flag.Parse()
	fmt.Println("输出：", arg)
}
