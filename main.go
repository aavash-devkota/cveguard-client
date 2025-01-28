package main

import (
	"cveguard-client/cmd"
	"github.com/joho/godotenv"
	"log"
	"os"
	"path/filepath"
)

func main() {
	loadDotEnv()
	cmd.Execute()
}

func loadDotEnv() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	godotenv.Load(filepath.Join(homeDir, ".cveguard-client.env"))
}
