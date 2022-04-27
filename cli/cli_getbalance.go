package cli

import (
	"blockchain_go/blockchain"
	"blockchain_go/utils"
	"blockchain_go/wallet"
	"fmt"
	"log"
)

func (cli *CLI) GetBalance(address string) {
	if !wallet.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := blockchain.NewBlockchain()
	UTXOSet := blockchain.UTXOSet{bc}
	FTXSet := blockchain.FTXset{bc}
	defer bc.DB.Close()

	balance := 0
	pubKeyHash := utils.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUTXO(pubKeyHash)
	FTXs := FTXSet.FindFTX(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
	if len(FTXs) > 0 {
		fmt.Printf("List of file hash\n")
	}
	for i, link := range FTXs {
		fmt.Printf("(%d) %s\n", i, link)
	}
}
