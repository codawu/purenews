package main

import (
	"fmt"
	"os"
	"purenews/prepare"
	"purenews/router"
)

func main() {
	if err := prepare.Serve(new(router.News)); err != nil {
		fmt.Printf("Serve start error %q \n", err)
		os.Exit(1)
	}
}
