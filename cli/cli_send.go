package cli

import (
	"fmt"
	"log"
	"os"

	"github.com/thanhxeon2470/testchain/blockchain"
	"github.com/thanhxeon2470/testchain/wallet"
)

func (cli *CLI) Send(prkFrom, to string, amount int, mineNow bool) {
	// if !wallet.ValidateAddress(prkFrom) {
	// 	log.Panic("ERROR: Sender address is not valid")
	// }
	if !wallet.ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	bc := blockchain.NewBlockchain()
	UTXOSet := blockchain.UTXOSet{bc}
	defer bc.DB.Close()

	// wallets, err :=
	// if err != nil {
	// 	log.Panic(err)
	// }
	w := wallet.DecodePrivKey([]byte(prkFrom))
	tx := blockchain.NewUTXOTransaction(w, to, amount, &UTXOSet)

	if mineNow {
		cbTx := blockchain.NewCoinbaseTX(string(w.GetAddress()), "")
		txs := []*blockchain.Transaction{cbTx, tx}

		newBlock := bc.MineBlock(txs)
		UTXOSet.Update(newBlock)
	} else {
		sendTx(os.Getenv("KNOWNNODE"), tx)
	}

	fmt.Println("Success!")
}
