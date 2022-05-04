package cli

import (
	"fmt"
	"log"

	"github.com/thanhxeon2470/TSS_chain/wallet"
)

func (cli *CLI) StartNode(minerAddress string) {
	fmt.Printf("Starting node\n")
	if len(minerAddress) > 0 {
		if wallet.ValidateAddress(minerAddress) {
			fmt.Println("Mining is on. Address to receive rewards: ", minerAddress)
		} else {
			log.Panic("Wrong miner address!")
		}
	}
	StartServer(minerAddress)
}
