package main

import (
	"log"

	"go-multipart-zip/unzip"

	"github.com/spf13/viper"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	viper.AutomaticEnv()
	viper.SetConfigName("application")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
}

func main() {
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("No configuration file found: %v\n", err)
	}

	if err := unzip.Cmd.Execute(); err != nil {
		log.Fatalf("Error executing command: %v", err)
	}
}
