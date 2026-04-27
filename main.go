package main

import (
	"go-multipart-zip/util"
	"log"

	"github.com/spf13/cobra"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	rootCmd := &cobra.Command{}
	rootCmd.AddCommand(util.UnzipCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error executing command: %v", err)
	}
}
