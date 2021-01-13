package main

import (
	"fmt"
	"os"
	"time"
)
import "zldface_server/utils"

func ff() {
	f, _ := os.Open("test.go")
	go func() {
		err := utils.SaveFile(f, "test_bak.go")
		fmt.Print(err)
	}()
	defer f.Close()
}
func main() {
	ff()
	time.Sleep(time.Second)
}
