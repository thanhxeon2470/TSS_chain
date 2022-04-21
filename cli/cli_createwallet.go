package cli

import "fmt"

func (cli *CLI) createWallet() {
	wallet := NewWallet()
	address := wallet.GetAddress()

	fmt.Printf("Your Bitcoin private key should be kept a secret. Whomever you share the private key with has access to spend all the bitcoins associated with that address.\n Your new address: %s\n Your private key: %s\n", address, EncondePrivKey(wallet.PrivateKey))
}

func (cli *CLI) AddWallet(priKey []byte) {

	wallet := DecodePrivKey(priKey)

	fmt.Printf("Your addres: %s\n", wallet.GetAddress())

}
