package main

import (
	"testchain/cli"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	cli := cli.CLI{}
	cli.Run()
}
