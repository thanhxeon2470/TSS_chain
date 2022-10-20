package cli

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/thanhxeon2470/TSS_chain/blockchain"
	"github.com/thanhxeon2470/TSS_chain/p2p"
	"github.com/thanhxeon2470/TSS_chain/wallet"
)

func (cli *CLI) Send(prkFrom, to string, amount int) bool {
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
	defer bc.DB.Close()
	UTXOSet := blockchain.UTXOSet{Blockchain: bc}
	tx := blockchain.NewUTXOTransaction(w, to, amount, nil, nil, &UTXOSet)
	if tx == nil {
		fmt.Println("Fail to create transaction!")

		return false
	}
	nodes := os.Getenv("BOOTSNODES")
	if nodes == "" {
		fmt.Printf("BOOTSNODES env. var is not set!")
		os.Exit(1)
	}
	bootsNodestmp := strings.Split(nodes, "_")
	p2p.InitP2P(0, bootsNodestmp, false)
	time.Sleep(2 * time.Second)
	SendTx(tx)
	time.Sleep(time.Second)
	fmt.Println("Success!")
	return true
}
