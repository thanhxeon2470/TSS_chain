package cli

import (
	"fmt"
	"log"
	"testchain/blockchain"
	"testchain/wallet"
)

func (cli *CLI) CreateBlockchain(address string) {
	if !wallet.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := blockchain.CreateBlockchain(address)
	defer bc.DB.Close()

	UTXOSet := blockchain.UTXOSet{bc}
	FTXSet := blockchain.FTXset{bc}
	UTXOSet.Reindex()
	FTXSet.ReindexFTX()

	fmt.Println("Done!")
}
