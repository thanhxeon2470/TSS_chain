package cli

import (
	"fmt"
	"log"

	"github.com/thanhxeon2470/testchain/blockchain"
	"github.com/thanhxeon2470/testchain/wallet"
)

func (cli *CLI) CreateBlockchain(address string) {
	if !wallet.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := blockchain.CreateBlockchain(address)
	defer bc.DB.Close()

	UTXOSet := blockchain.UTXOSet{bc}
	UTXOSet.Reindex()

	fmt.Println("Done!")
}
