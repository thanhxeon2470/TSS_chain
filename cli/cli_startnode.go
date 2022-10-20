package cli

import (
	"fmt"
	"log"
	"os"

	"github.com/thanhxeon2470/TSS_chain/helper"
	"github.com/thanhxeon2470/TSS_chain/wallet"
)

// type ipfsID struct {
// 	ID              string   `json:"ID"`
// 	PublicKey       string   `json:"PublicKey"`
// 	Addresses       []string `json:"Addresses"`
// 	AgentVersion    string   `json:"AgentVersion"`
// 	ProtocolVersion string   `json:"ProtocolVersion"`
// 	Protocols       []string `json:"Protocols"`
// }

func (cli *CLI) StartNode(thisNode, minerAddress string) {
	fmt.Printf("Starting node\n")
	if len(minerAddress) > 0 {
		if wallet.ValidateAddress(minerAddress) {
			fmt.Println("Mining is on. Address to receive rewards: ", minerAddress)
		} else {
			log.Panic("Wrong miner address!")
		}
	}
	if len(StorageMiningAddress) > 0 {
		if wallet.ValidateAddress(StorageMiningAddress) {
			fmt.Println("Storage Mining is on. Address to receive rewards: ", StorageMiningAddress)
		} else {
			log.Panic("Wrong storage miner address!")
		}
		if !(helper.IpfsIsRunning() && helper.IpfsClusterIsRunning()) {
			os.Stderr.WriteString("Oops!!")
			os.Exit(1)
			return
		}
	}
	StartServer(minerAddress)
}
