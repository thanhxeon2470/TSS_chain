package cli

import (
	"fmt"
	"log"
	"os"

	"github.com/thanhxeon2470/TSS_chain/blockchain"
	"github.com/thanhxeon2470/TSS_chain/wallet"
)

func (cli *CLI) Send(prkFrom, to string, amount int, allowuser []string, iHash string, mineNow bool) bool {
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

	if mineNow {
		bc := blockchain.NewBlockchain()
		UTXOSet := blockchain.UTXOSet{bc}
		FTX := blockchain.FTXset{bc}
		defer bc.DB.Close()

		tx := blockchain.NewUTXOTransaction(w, to, amount, allowuser, iHash, &UTXOSet)
		cbTx := blockchain.NewCoinbaseTX(string(w.GetAddress()), "")
		txs := []*blockchain.Transaction{cbTx, tx}

		newBlock := bc.MineBlock(txs)
		UTXOSet.Update(newBlock)
		FTX.UpdateFTX(newBlock)
	} else {
		bc := blockchain.NewBlockchainView()
		UTXOSet := blockchain.UTXOSet{bc}
		tx := blockchain.NewUTXOTransaction(w, to, amount, allowuser, iHash, &UTXOSet)
		sendTx(os.Getenv("KNOWNNODE"), tx)
	}

	fmt.Println("Success!")
	return true
}
