package main

import (
	"context"
	"os"

	"github.com/sadiksaifi/aiusage/internal/app"
)

func main() {
	if err := app.Run(context.Background()); err != nil {
		os.Exit(1)
	}
}
