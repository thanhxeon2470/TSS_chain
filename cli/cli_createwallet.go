package cli

import (
	"fmt"

	"github.com/thanhxeon2470/TSS_chain/utils"
	"github.com/thanhxeon2470/TSS_chain/wallet"
)

func (cli *CLI) CreateWallet() ([]byte, []byte, []byte) {
	w, _ := wallet.NewWallet()
	address := w.GetAddress()

	prk := wallet.EncodePrivKey(w.PrivateKey)
	pub := utils.Base58Encode(w.PublicKey)
	fmt.Printf("Your TSS private key should be kept a secret. Whomever you share the private key with has access to spend all the bitcoins associated with that address.\n Your new address: %s\n Your private key: %s\n", address, prk)
	return address, pub, prk
}

func (cli *CLI) AddWallet(priKey []byte) ([]byte, []byte) {
	wallet := wallet.DecodePrivKey(priKey)
	addr := wallet.GetAddress()
	pub := utils.Base58Encode(wallet.PublicKey)
	fmt.Printf("Your public key: %s\n", pub)
	fmt.Printf("And addres: %s\n", addr)
	return pub, addr
}
