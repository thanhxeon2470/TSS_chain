package helper

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type ipfsID struct {
	ID              string   `json:"ID"`
	PublicKey       string   `json:"PublicKey"`
	Addresses       []string `json:"Addresses"`
	AgentVersion    string   `json:"AgentVersion"`
	ProtocolVersion string   `json:"ProtocolVersion"`
	Protocols       []string `json:"Protocols"`
}

func IpfsIsRunning() bool {
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

func IpfsClusterIsRunning() bool {
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
func getFileHash(stout []byte) string {
	stoutstr := string(stout)
	fhphase := strings.Split(stoutstr, "\n")[0]
	fh := strings.Split(fhphase, " ")[1]
	return fh
}

func IpfsAdd(filepath string) (string, error) {
	if isRunning := IpfsIsRunning(); !isRunning {
		return "", fmt.Errorf("IPFS is not running @.@")
	}
	addCmd := exec.Command("ipfs", "add", filepath)

	stdout, err := addCmd.Output()
	if err != nil {
		return "", err
	}
	return getFileHash(stdout), nil
}

func IpfsClusterAdd(filepath string) (string, error) {
	if isRunning := IpfsClusterIsRunning(); !isRunning {
		return "", fmt.Errorf("IPFS is not running @.@")
	}
	addCmd := exec.Command("ipfs-cluster-ctl", "add", filepath)

	stdout, err := addCmd.Output()
	if err != nil {
		return "", err
	}
	return getFileHash(stdout), nil
}

func IpfsGet(ipfsHash string) (bool, error) {
	if isRunning := IpfsIsRunning(); !isRunning {
		return false, fmt.Errorf("IPFS is not running @.@")
	}
	addCmd := exec.Command("ipfs", "get", ipfsHash)

	stdout, err := addCmd.Output()
	if err != nil {
		return false, err
	}
	str := string(stdout)
	return strings.Contains(str, ipfsHash), nil
}
