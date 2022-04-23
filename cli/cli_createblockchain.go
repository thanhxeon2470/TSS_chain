package cli

import (
	"blockchain_go/blockchain"
	"blockchain_go/wallet"
	"fmt"
	"log"
)

func (cli *CLI) createBlockchain(address, nodeID string) {
	if !wallet.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := blockchain.CreateBlockchain(address, nodeID)
	defer bc.db.Close()

	UTXOSet := blockchain.UTXOSet{bc}
	FTXSet := blockchain.FTXset{bc}
	UTXOSet.Reindex()
	FTXSet.ReindexFTX()

	fmt.Println("Done!")
}
