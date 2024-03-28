package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {

	switch os.Args[1] {
	case "run":
		args := RunArgs{}
		flag.StringVar(&args.ConfigPath, "config", "", "configuration file path")
		flag.Parse()

		app = GetApplication(&args)

		app.Run()
	default:
		fmt.Printf("Supported arguments: run")
	}

}
