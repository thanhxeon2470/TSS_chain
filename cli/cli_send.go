package cli

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/thanhxeon2470/TSS_chain/blockchain"
	"github.com/thanhxeon2470/TSS_chain/wallet"
)

func (cli *CLI) Send(prkFrom, to string, amount int, mineNow bool) bool {
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
		defer bc.DB.Close()

		tx := blockchain.NewUTXOTransaction(w, to, amount, nil, "", &UTXOSet)
		if tx == nil {
			fmt.Println("Fail to create transaction!")

			return false
		}
		cbTx := blockchain.NewCoinbaseTX(string(w.GetAddress()), "")
		txs := []*blockchain.Transaction{cbTx, tx}

		newBlock := bc.MineBlock(txs)
		UTXOSet.Update(newBlock)
	} else {
		bc := blockchain.NewBlockchainView()
		defer bc.DB.Close()
		UTXOSet := blockchain.UTXOSet{bc}
		tx := blockchain.NewUTXOTransaction(w, to, amount, nil, "", &UTXOSet)
		if tx == nil {
			fmt.Println("Fail to create transaction!")

			return false
		}
		sendTx(strings.Split(os.Getenv("KNOWNNODE"), "_")[0], tx)
	}

	fmt.Println("Success!")
	return true
}
