package cli

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/thanhxeon2470/TSS_chain/wallet"
)

type ipfsID struct {
	ID              string   `json:"ID"`
	PublicKey       string   `json:"PublicKey"`
	Addresses       []string `json:"Addresses"`
	AgentVersion    string   `json:"AgentVersion"`
	ProtocolVersion string   `json:"ProtocolVersion"`
	Protocols       []string `json:"Protocols"`
}

func ipfsIsRunning() bool {
	idCmd := exec.Command("ipfs", "id")
	stdout, err := idCmd.Output()
	if err != nil {
		return false
	}
	idIn4 := ipfsID{}
	err = json.Unmarshal(stdout, &idIn4)
	if err != nil {
		fmt.Println("unmarshle k dc ")
		return false
	}

	if idIn4.Addresses == nil {
		fmt.Println("Ipfs is stopped!")
		return false
	}
	fmt.Println("Ipfs is running!")
	return true

}

func ipfsClusterIsRunning() bool {
	idCmd := exec.Command("ipfs-cluster-ctl", "id")
	stdout, err := idCmd.Output()
	if err != nil {
		return false
	}
	str := string(stdout)
	if strings.Contains(str, "Addresses") {
		fmt.Println("Ipfs cluster ctl is running!")
		return true
	}

	fmt.Println("Ipfs cluster ctl is stopped!")
	return false
}

func (cli *CLI) StartNode(minerAddress string) {
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
		if !(ipfsIsRunning() && ipfsClusterIsRunning()) {
			os.Stderr.WriteString("Oops!!")
			os.Exit(1)
			return
		}
	}
	StartServer(minerAddress)
}
