package cli

import (
	"log"
	"os"

	"github.com/thanhxeon2470/TSS_chain/blockchain"
	"github.com/thanhxeon2470/TSS_chain/wallet"
)

func (cli *CLI) SendProposal(prkFrom, to string, amount int, allowuser []string, iHash string) bool {
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
	proposal := proposal{tx.ID, []byte(to), []byte(iHash), amount}
	sendProposal(os.Getenv("KNOWNNODE"), proposal)
	sendTx("127.0.0.1:3000", tx)

	return true
}
