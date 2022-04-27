package main

import (
	"blockchain_go/cli"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	cli := cli.CLI{}
	cli.Run()
}
