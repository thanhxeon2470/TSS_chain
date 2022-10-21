package main

import (
	"github.com/joho/godotenv"
	"github.com/thanhxeon2470/TSS_chain/cli"
)

func main() {
	godotenv.Load()

	cli := cli.CLI{}
	cli.Run()
}
