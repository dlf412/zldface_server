package main

import (
	"flag"
	"fmt"
	"os"
	"zldface_server/cache"
)

func main() {
	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		fmt.Printf("    %s [load|clear|reload] \n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}
	operation := flag.Args()[0]
	if operation == "load" {
		cache.LoadAllFeatures()
	} else if operation == "clear" {
		cache.ClearAllFeatures()
	} else if operation == "reload" {
		cache.ClearAllFeatures()
		cache.LoadAllFeatures()
	} else {
		flag.Usage()
		os.Exit(1)
	}
}
