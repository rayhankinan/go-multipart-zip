package main

import (
	"go-multipart-zip/util"
	"log"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	if err := util.UnzipCmd.Execute(); err != nil {
		log.Fatalf("Error executing command: %v", err)
	}
}
