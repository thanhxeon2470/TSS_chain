package cli

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto/ecies"
	"github.com/thanhxeon2470/TSS_chain/blockchain"
	"github.com/thanhxeon2470/TSS_chain/wallet"
)

func (cli *CLI) IPFSget(prk, ipfsHashENC string) string {
	w := wallet.DecodePrivKey([]byte(prk))

	bc := blockchain.NewBlockchainView()
	defer bc.DB.Close()
	FTXSet := blockchain.FTXset{Blockchain: bc}
	ipfsHashBytes, err := hex.DecodeString(ipfsHashENC)
	if err != nil {
		return ""
	}
	listUser := FTXSet.FindIPFS(ipfsHashBytes)
	addr := w.GetAddress()
	if len(listUser) > 0 {
		for user := range listUser {
			if bytes.Equal(addr, []byte(user)) {
				priKey := ecies.ImportECDSA(&w.PrivateKey)
				ifh, err := hex.DecodeString(ipfsHashENC)
				if err != nil {
					fmt.Println("Private key invalid or wrong IPFS hash Encript")
					return ""
				}
				iHash, err := priKey.Decrypt(ifh, nil, nil)
				if err != nil {
					fmt.Println("Private key invalid or wrong IPFS hash Encript")
					return ""
				}
				fmt.Println("IPFS hash: ", string(iHash))
				return string(iHash)
			}
		}
	}
	fmt.Println("Private key invalid or wrong IPFS hash Encript")
	return ""
}
