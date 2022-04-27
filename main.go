package main

import (
	"github.com/thanhxeon2470/testchain/cli"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	cli := cli.CLI{}
	cli.Run()
}
