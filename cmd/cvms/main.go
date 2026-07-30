package main

import (
	"clu-vms/internal/impl/core"
	"fmt"
	"os"
)

var Version = "development"

func main() {
	fmt.Println("cvms " + Version)
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		return
	}
	ctx := core.NewLocalContext(wd)
	err = ctx.RunCommand([]string{"ls", "-l"})
	if err != nil {
		fmt.Println(err)
	}
}
