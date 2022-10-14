package cli

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/thanhxeon2470/TSS_chain/blockchain"
	"github.com/thanhxeon2470/TSS_chain/p2p"
	"github.com/thanhxeon2470/TSS_chain/wallet"

	"github.com/ethereum/go-ethereum/crypto/ecies"
)

// Send proposal to the strorage miner and get encode of iHash
func (cli *CLI) SendProposal(prkFrom, to string, amount int, iHash string) string {
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
	UTXOSet := blockchain.UTXOSet{bc}

	prkECIES := ecies.ImportECDSA(&w.PrivateKey)
	ct, err := ecies.Encrypt(rand.Reader, &prkECIES.PublicKey, []byte(iHash), nil, nil)
	if err != nil {
		return ""
	}
	tx := blockchain.NewUTXOTransaction(w, to, amount, nil, ct, &UTXOSet)

	if tx == nil {
		fmt.Println("Fail to create transaction!")

		return ""
	}
	proposal := Proposal{tx.ID, []byte(to), []byte(iHash), amount}

	nodes := os.Getenv("BOOTSNODES")
	if nodes == "" {
		fmt.Printf("BOOTSNODES env. var is not set!")
		os.Exit(1)
	}
	bootsNodestmp := strings.Split(nodes, "_")
	p2p.InitP2P(0, bootsNodestmp, false)

	SendProposal(proposal)
	SendTx(tx)

	return hex.EncodeToString(ct)
}
