package main

import (
	"context"
	"os"

	"github.com/nicoxiang/geektime-downloader/internal/uiapp"
)

func init() {
	// Get around rsa1024min panic issue
	_ = os.Setenv("GODEBUG", os.Getenv("GODEBUG")+",rsa1024min=0")
}

func main() {
	uiapp.Run(context.Background())
}
