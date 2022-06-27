package rpc

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"strings"
	"time"

	"github.com/thanhxeon2470/TSS_chain/blockchain"
	"github.com/thanhxeon2470/TSS_chain/utils"
	"github.com/thanhxeon2470/TSS_chain/wallet"
)

const protocol = "tcp"
const nodeVersion = 1
const commandLength = 12

var nodeAddress string
var miningAddress string
var StorageMiningAddress string
var proposalCheck = false
var randomIdentity = 0
var knownNodes = []string{}
var blocksInTransit = [][]byte{}

var mempool = make(map[string]blockchain.Transaction)
var timeReceivedTx = make(chan int64)
var timeMining int64 = 30 // 30s

type addr struct {
	AddrList []string
}

type getdata struct {
	// AddrFrom string
	Type string
	ID   []byte
}

type findipfs struct {
	IpfsHashENC []byte
}
type Ipfs struct {
	User map[string]bool
}

type getbalance struct {
	Addr string
}
type Balance struct {
	Value int
	FTXs  map[string]blockchain.InfoIPFS
}

type gettxins struct {
	Addr string
}
type Txins struct {
	Inputs []byte
}

func commandToBytes(command string) []byte {
	var bytes [commandLength]byte

	for i, c := range command {
		bytes[i] = byte(c)
	}

	return bytes[:]
}

func bytesToCommand(bytes []byte) string {
	var command []byte

	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}

	return fmt.Sprintf("%s", command)
}

func extractCommand(request []byte) []byte {
	return request[:commandLength]
}

func sendData(addr string, data []byte) {
	time.Sleep(time.Second / 10)
	conn, err := net.Dial(protocol, addr)

	if err != nil {
		fmt.Printf("%s is not available\n", addr)

		return
	}
	defer conn.Close()

	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
}

func SendGetData(address, kind string, id []byte) {
	payload := gobEncode(getdata{kind, id})
	request := append(commandToBytes("getdata"), payload...)

	sendData(address, request)
}

func SendInforIPFS(addr string, user map[string]bool) {
	data := Ipfs{user}
	payload := gobEncode(data)
	request := append(commandToBytes("ipfs"), payload...)

	sendData(addr, request)
}

func SendFindIPFS(addr string, ipfsHashENC []byte) {
	data := findipfs{ipfsHashENC}
	payload := gobEncode(data)
	request := append(commandToBytes("findipfs"), payload...)

	sendData(addr, request)
}

func SendTxIns(addr string, txins *blockchain.TXInputs) {
	data := Txins{txins.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("txins"), payload...)

	sendData(addr, request)
}

func SendGetTxIns(addr string, addrTSS string) {
	data := gettxins{addrTSS}
	payload := gobEncode(data)
	request := append(commandToBytes("gettxins"), payload...)

	sendData(addr, request)
}

func SendGetBlance(addr string, addrTSS string) {
	payload := gobEncode(getbalance{addrTSS})
	request := append(commandToBytes("getbalance"), payload...)

	sendData(addr, request)
}

func SendBalance(addr string, bals int, FTXs map[string]blockchain.InfoIPFS) {

	payload := gobEncode(Balance{bals, FTXs})
	request := append(commandToBytes("balance"), payload...)

	sendData(addr, request)
}

func handleFindIPFS(request []byte, addrFrom string) {
	var buff bytes.Buffer
	var payload findipfs

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	bc := blockchain.NewBlockchainView()
	FTXSet := blockchain.FTXset{bc}
	defer bc.DB.Close()
	listUser := FTXSet.FindIPFS(string(payload.IpfsHashENC))

	SendInforIPFS(addrFrom, listUser)
}

func handleGetTxIns(request []byte, addrFrom string) {
	var buff bytes.Buffer
	var payload gettxins

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	address := payload.Addr
	if !wallet.ValidateAddress(address) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	bc := blockchain.NewBlockchainView()
	defer bc.DB.Close()

	UTXOSet := blockchain.UTXOSet{bc}

	pubKeyHash := utils.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-wallet.AddressChecksumLen]
	UTXOs := UTXOSet.FindUTXO(pubKeyHash)

	balance := 0
	for _, out := range UTXOs {
		balance += out.Value
	}
	acc, validOutputs := UTXOSet.FindSpendableOutputs(pubKeyHash, balance)

	txins := blockchain.TXInputs{nil}
	if acc < balance {
		log.Panic("ERROR: Not enough funds")

	}
	// Build a list of inputs
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			input := blockchain.TXInput{txID, out, nil, nil}
			txins.Inputs = append(txins.Inputs, input)

			fmt.Println(hex.EncodeToString(txID), " ==== ", out)
		}
	}

	SendTxIns(addrFrom, &txins)
}

func handleGetBlance(request []byte, addrFrom string) {
	var buff bytes.Buffer
	var payload getbalance

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	address := payload.Addr
	if !wallet.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := blockchain.NewBlockchainView()
	UTXOSet := blockchain.UTXOSet{bc}
	FTXSet := blockchain.FTXset{bc}
	defer bc.DB.Close()

	balance := 0
	pubKeyHash := utils.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUTXO(pubKeyHash)
	FTXs := FTXSet.FindFTX(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	SendBalance(addrFrom, balance, FTXs)
}

func HandleRPC(conn net.Conn) {
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}
	command := bytesToCommand(request[:commandLength])
	fmt.Printf("Received %s command\n", command)
	addrFrom := fmt.Sprintf("%s:%s", strings.Split(conn.RemoteAddr().String(), ":")[0], "3456")
	switch command {
	case "getbalance":
		handleGetBlance(request, addrFrom)
	case "gettxins":
		handleGetTxIns(request, addrFrom)
	case "findipfs":
		handleFindIPFS(request, addrFrom)
	default:
		fmt.Println("Unknown command!")
	}

	conn.Close()
}

func handleBlance(request []byte) []byte {
	var buff bytes.Buffer

	buff.Write(request[commandLength:])
	return buff.Bytes()

}

func handleTxIns(request []byte) []byte {
	var buff bytes.Buffer

	buff.Write(request[commandLength:])
	return buff.Bytes()

}
func handleIPFS(request []byte) []byte {
	var buff bytes.Buffer

	buff.Write(request[commandLength:])
	return buff.Bytes()

}

func HandleRPCReceive(conn net.Conn) []byte {
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}
	command := bytesToCommand(request[:commandLength])
	fmt.Printf("Received %s command\n", command)
	switch command {
	case "balance":
		return handleBlance(request)
	case "txins":
		return handleTxIns(request)
	case "ipfs":
		return handleIPFS(request)
	default:
		fmt.Println("Unknown command!")
	}

	conn.Close()
	return nil
}

func gobEncode(data interface{}) []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}
