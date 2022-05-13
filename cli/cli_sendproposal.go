package cli

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/thanhxeon2470/TSS_chain/blockchain"
	"github.com/thanhxeon2470/TSS_chain/wallet"
)

func (cli *CLI) sendProposal(prkFrom, to string, amount int, allowuser []string, iHash string) bool {
	// if !wallet.ValidateAddress(prkFrom) {
	// 	log.Panic("ERROR: Sender address is not valid")
	// }
	if !wallet.ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	// wallets, err :=
	// if err != nil {
	// 	log.Panic(err)
	// }
	w := wallet.DecodePrivKey([]byte(prkFrom))

	bc := blockchain.NewBlockchainView()
	UTXOSet := blockchain.UTXOSet{bc}
	tx := blockchain.NewUTXOTransaction(w, to, amount, allowuser, iHash, &UTXOSet)
	proposal := proposal{[]byte(to), []byte(iHash), amount}
	sendProposal(os.Getenv("KNOWNNODE"), proposal)
	timeCreateTx := time.Now().Unix()
	fmt.Print("Wait for storage miner accept proposal!...")

	// wait for proposal response
	for (time.Now().Unix() - timeCreateTx) > 30 {
		if proposalCheck == true {
			sendTx(os.Getenv("KNOWNNODE"), tx)
			proposalCheck = false
			break
		}
	}

	fmt.Println("Deal Successfully!")
	return true
}
