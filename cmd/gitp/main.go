package main

import (
	"os"

	"github.com/prajwal/gitp/internal/app"
)

func main() {
	os.Exit(app.Run(os.Args[1:]))
}
