package main

import (
	"github.com/thanhxeon2470/TSS_chain/cli"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	// fmt.Println("BAY DAU A")
	// bc := blockchain.NewBlockchainView()
	// fmt.Println("BAY DAU A")
	// best , err := bc.GetLastBlock()
	// if err != nil {
	// 	fmt.Println("NHU CC", err)
	// }
	// fmt.Printf("AVASOFIJOAJW: +v" , best)
	cli := cli.CLI{}
	cli.Run()
}
