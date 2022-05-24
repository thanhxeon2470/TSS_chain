package cli

import (
	"fmt"
	"log"
	"time"

	"github.com/thanhxeon2470/TSS_chain/blockchain"
	"github.com/thanhxeon2470/TSS_chain/utils"
	"github.com/thanhxeon2470/TSS_chain/wallet"
)

func (cli *CLI) GetBalance(address string) (string, int, map[string]blockchain.InfoIPFS) {
	if !wallet.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := blockchain.NewBlockchainView()
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
	i := 0
	for link, in4 := range FTXs {
		i += 1
		if in4.Author {
			fmt.Printf("(%d) %s | %s | (author)\n", i, time.Unix(in4.Exp, 0), link)
		} else {
			fmt.Printf("(%d) %s | %s | \n", i, time.Unix(in4.Exp, 0), link)
		}
	}
	return address, balance, FTXs
}
