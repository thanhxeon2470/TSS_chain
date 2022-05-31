package cli

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/thanhxeon2470/TSS_chain/blockchain"
	"github.com/thanhxeon2470/TSS_chain/wallet"

	"github.com/ethereum/go-ethereum/crypto/ecies"
)

// Send proposal to the strorage miner and get encode of iHash
func (cli *CLI) SendProposal(prkFrom, to string, amount int, iHash string) []byte {
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

	prkECIES := ecies.ImportECDSA(&w.PrivateKey)
	ct, err := ecies.Encrypt(rand.Reader, &prkECIES.PublicKey, []byte(iHash), nil, nil)
	if err != nil {
		return nil
	}
	tx := blockchain.NewUTXOTransaction(w, to, amount, nil, hex.EncodeToString(ct), &UTXOSet)

	if tx == nil {
		fmt.Println("Fail to create transaction!")

		return nil
	}
	proposal := proposal{tx.ID, []byte(to), []byte(iHash), amount}
	sendProposal(strings.Split(os.Getenv("KNOWNNODE"), "_")[0], proposal)
	sendTx("127.0.0.1:3000", tx)

	return ct
}
