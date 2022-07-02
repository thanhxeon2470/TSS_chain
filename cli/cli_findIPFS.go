package cli

import (
	"encoding/hex"
	"fmt"

	"github.com/thanhxeon2470/TSS_chain/blockchain"
)

// This return list user and author. If the boolen is true, this user is owner
func (cli *CLI) FindIPFS(ipfsHash string) map[string]bool {
	// if !wallet.ValidateAddress(address) {
	// 	log.Panic("ERROR: Address is not valid")
	// }
	bc := blockchain.NewBlockchainView()
	FTXSet := blockchain.FTXset{bc}
	defer bc.DB.Close()

	// balance := 0
	// pubKeyHash := utils.Base58Decode([]byte(address))
	// pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	// UTXOs := UTXOSet.FindUTXO(pubKeyHash)
	// FTXs := FTXSet.FindFTX(pubKeyHash)
	ipfsHashBytes, err := hex.DecodeString(ipfsHash)
	if err != nil {
		return nil
	}
	listUser := FTXSet.FindIPFS(ipfsHashBytes)
	fmt.Printf("IPFS hash: %s\n", ipfsHash)
	if len(listUser) > 0 {
		fmt.Printf("List user:\n")
		i := 1
		for user, author := range listUser {
			if author {
				fmt.Printf("(%d) %s | %s | \n", i, user, "author")
			} else {
				fmt.Printf("(%d) %s | \n", i, user)
			}

			i += 1

		}
	}

	return listUser
}
