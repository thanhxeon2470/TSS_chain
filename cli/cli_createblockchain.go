package cli

import (
	"fmt"
	"log"

	"github.com/thanhxeon2470/TSS_chain/blockchain"
	"github.com/thanhxeon2470/TSS_chain/wallet"
)

func (cli *CLI) CreateBlockchain(address string) {
	if !wallet.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := blockchain.CreateBlockchain(address)
	defer bc.DB.Close()

	UTXOSet := blockchain.UTXOSet{Blockchain: bc}
	FTXSet := blockchain.FTXset{Blockchain: bc}
	UTXOSet.Reindex()
	FTXSet.ReindexFTX()

	fmt.Println("Done!")
}
