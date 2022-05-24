package cli

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"

	"github.com/thanhxeon2470/TSS_chain/blockchain"
	"github.com/thanhxeon2470/TSS_chain/helper"
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

type addr struct {
	AddrList []string
}

type block struct {
	// AddrFrom string
	Block []byte
}

// type getblocks struct {
// 	AddrFrom string
// }

type getdata struct {
	// AddrFrom string
	Type string
	ID   []byte
}

type inv struct {
	// AddrFrom string
	Type  string
	Items [][]byte
}

type tx struct {
	// AddrFrom    string
	Transaction []byte
}

type verzion struct {
	Version    int
	BestHeight int
	// AddrFrom   string
}

type proposal struct {
	TxHash               []byte
	StorageMiningAddress []byte
	FileHash             []byte
	Amount               int
}

// feedback proposal
type fbproposal struct {
	TxHash []byte
	Accept bool
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

// func requestBlocks(bc *blockchain.Blockchain) {
// 	for _, node := range knownNodes {
// 		sendVersion(node, bc)
// 	}
// }

// func sendAddr(address string) {
// 	nodes := addr{knownNodes}
// 	nodes.AddrList = append(nodes.AddrList)
// 	payload := gobEncode(nodes)
// 	request := append(commandToBytes("addr"), payload...)

// 	sendData(address, request)
// }

func sendBlock(addr string, b *blockchain.Block) {
	data := block{b.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("block"), payload...)
	sendData(addr, request)
}

func sendData(addr string, data []byte) {
	conn, err := net.Dial(protocol, addr)
	if err != nil {
		fmt.Printf("%s is not available\n", addr)
		var updatedNodes []string

		for _, node := range knownNodes {
			if node != addr {
				updatedNodes = append(updatedNodes, node)
			}
		}

		knownNodes = updatedNodes

		return
	}
	defer conn.Close()

	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
}

func sendInv(address, kind string, items [][]byte) {
	inventory := inv{kind, items}
	payload := gobEncode(inventory)
	request := append(commandToBytes("inv"), payload...)

	sendData(address, request)
}

func sendGetBlocks(address string) {
	// payload := gobEncode(getblocks{nodeAddress})
	request := commandToBytes("getblocks")

	sendData(address, request)
}

func sendGetData(address, kind string, id []byte) {
	payload := gobEncode(getdata{kind, id})
	request := append(commandToBytes("getdata"), payload...)

	sendData(address, request)
}

func sendTx(addr string, tnx *blockchain.Transaction) {
	data := tx{tnx.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("tx"), payload...)

	sendData(addr, request)
}

func sendVersion(addr string, bc *blockchain.Blockchain) {
	bestHeight := bc.GetBestHeight()
	payload := gobEncode(verzion{nodeVersion, bestHeight})
	request := append(commandToBytes("version"), payload...)

	sendData(addr, request)
}

func sendProposal(addr string, pps proposal) {
	payload := gobEncode(pps)
	request := append(commandToBytes("proposal"), payload...)
	sendData(addr, request)
}

func sendFBProposal(addr string, txid []byte, feedback bool) {
	payload := gobEncode(fbproposal{txid, feedback})
	request := append(commandToBytes("feedback"), payload...)
	sendData(addr, request)
}

func handleProposal(request []byte, addrFrom, addrLocal string) {
	var buff bytes.Buffer
	var payload proposal
	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	// add file to ipfs when receive tx
	if len(StorageMiningAddress) > 0 {
		if bytes.Compare(payload.StorageMiningAddress, []byte(StorageMiningAddress)) == 0 {
			// Get file on ipfs
			fh := string(payload.FileHash)
			getCMD := exec.Command("ipfs", "get", fh)
			stdout, err := getCMD.Output()
			if err != nil {
				return
			}
			str := string(stdout)
			if strings.Contains(str, fh) {
				// And update this file to ipfs cluster
				addCMD := exec.Command("ipfs-cluster-ctl", "add", fh)
				stdout, err := addCMD.Output()
				if err != nil {
					return
				}
				str := string(stdout)
				if !strings.Contains(str, "added") {
					fmt.Print("Cant add file to ipfs")
				}
			} else {
				fmt.Print("Cant get file from ipfs")
			}

			sendFBProposal(addrFrom, payload.TxHash, true)
			return
		}
	}
	if addrLocal == knownNodes[0] {
		for _, node := range knownNodes {
			if node != addrLocal && node != addrFrom {
				sendProposal(node, payload)
			}
		}
	}
}

func handleFeedback(request []byte, addrFrom, addrLocal string) {
	var buff bytes.Buffer
	var payload fbproposal
	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	// proposalCheck = payload.Accept
	// randomIdentity = payload.RandomIdentity
	if addrLocal == knownNodes[0] {
		for _, node := range knownNodes {
			if node != addrLocal && node != addrFrom {
				sendFBProposal(node, payload.TxHash, payload.Accept)
			}
		}
	} else if len(mempool) > 0 {
		// When received feedback proposal =>>> check this and send transaction
		for id := range mempool {
			if bytes.Compare([]byte(id), payload.TxHash) == 0 {
				tx := mempool[id]
				sendTx(knownNodes[0], &tx)
			}
		}
	}
}

// func handleAddr(request []byte, bc *blockchain.Blockchain) {
// 	var buff bytes.Buffer
// 	var payload addr

// 	buff.Write(request[commandLength:])
// 	dec := gob.NewDecoder(&buff)
// 	err := dec.Decode(&payload)
// 	if err != nil {
// 		log.Panic(err)
// 	}

// 	knownNodes = append(knownNodes, payload.AddrList...)
// 	fmt.Printf("There are %d known nodes now!\n", len(knownNodes))
// }

func handleBlock(request []byte, bc *blockchain.Blockchain, addrFrom, localAddr string) {
	var buff bytes.Buffer
	var payload block

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blockData := payload.Block
	block := blockchain.DeserializeBlock(blockData)

	fmt.Println("Recevied a new block!")
	bc.AddBlock(block)

	fmt.Printf("Added block %x\n", block.Hash)

	if localAddr == knownNodes[0] {
		for _, node := range knownNodes {
			if node != localAddr && node != addrFrom {
				sendBlock(node, block)
				fmt.Printf("This block is broadcasted to %s\n", node)
			}
		}
	}

	if len(blocksInTransit) > 0 {
		blockHash := blocksInTransit[0]
		sendGetData(addrFrom, "block", blockHash)

		blocksInTransit = blocksInTransit[1:]
	} else {
		UTXOSet := blockchain.UTXOSet{bc}
		FTXSet := blockchain.FTXset{bc}
		UTXOSet.Reindex()
		FTXSet.ReindexFTX()

	}
}

func handleInv(request []byte, bc *blockchain.Blockchain, addrFrom string) {
	var buff bytes.Buffer
	var payload inv

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("Recevied inventory with %d %s\n", len(payload.Items), payload.Type)

	if payload.Type == "block" {
		blocksInTransit = payload.Items

		blockHash := payload.Items[0]
		sendGetData(addrFrom, "block", blockHash)

		//doan nay vo dung
		newInTransit := [][]byte{}
		for _, b := range blocksInTransit {
			if bytes.Compare(b, blockHash) != 0 {
				newInTransit = append(newInTransit, b)
			}
		}
		blocksInTransit = newInTransit
	}

	if payload.Type == "tx" {
		txID := payload.Items[0]

		if mempool[hex.EncodeToString(txID)].ID == nil {
			sendGetData(addrFrom, "tx", txID)
		}
	}
}

func handleGetBlocks(request []byte, bc *blockchain.Blockchain, addrFrom string) {
	blocks := bc.GetBlockHashes()
	sendInv(addrFrom, "block", blocks)
}

func handleGetData(request []byte, bc *blockchain.Blockchain, addrFrom string) {
	var buff bytes.Buffer
	var payload getdata

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	if payload.Type == "block" {
		block, err := bc.GetBlock([]byte(payload.ID))
		if err != nil {
			return
		}

		sendBlock(addrFrom, &block)
	}

	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID)
		tx := mempool[txID]

		sendTx(addrFrom, &tx)
		// delete(mempool, txID)
	}
}

func handleTx(request []byte, bc *blockchain.Blockchain, addrFrom string, addrLocal string) {
	var buff bytes.Buffer
	var payload tx

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	txData := payload.Transaction
	tx := blockchain.DeserializeTransaction(txData)
	mempool[hex.EncodeToString(tx.ID)] = tx

	if addrLocal == knownNodes[0] {
		for _, node := range knownNodes {
			if node != addrLocal && node != addrFrom {
				sendInv(node, "tx", [][]byte{tx.ID})
			}
		}
	} else {
		if len(mempool) >= 2 && len(miningAddress) > 0 {
		MineTransactions:
			var txs []*blockchain.Transaction

			for id := range mempool {
				tx := mempool[id]
				if bc.VerifyTransaction(&tx) {
					txs = append(txs, &tx)
				}
			}

			if len(txs) == 0 {
				fmt.Println("All transactions are invalid! Waiting for new ones...")
				return
			}

			cbTx := blockchain.NewCoinbaseTX(miningAddress, "")
			txs = append(txs, cbTx)

			newBlock := bc.MineBlock(txs)
			UTXOSet := blockchain.UTXOSet{bc}
			FTXSet := blockchain.FTXset{bc}
			UTXOSet.Reindex()
			FTXSet.ReindexFTX()

			fmt.Println("New block is mined!")

			for _, tx := range txs {
				txID := hex.EncodeToString(tx.ID)
				delete(mempool, txID)
			}

			for _, node := range knownNodes {
				sendInv(node, "block", [][]byte{newBlock.Hash})
			}

			if len(mempool) > 0 {
				goto MineTransactions
			}
		}
	}
}

func handleVersion(request []byte, bc *blockchain.Blockchain, addrFrom string) {
	var buff bytes.Buffer
	var payload verzion

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	myBestHeight := bc.GetBestHeight()
	foreignerBestHeight := payload.BestHeight

	fmt.Println("myBestHeight ", myBestHeight)
	if myBestHeight < foreignerBestHeight {
		sendGetBlocks(addrFrom)
	} else if myBestHeight > foreignerBestHeight {
		sendVersion(addrFrom, bc)
	}
	// sendAddr(addrFrom)
	if !nodeIsKnown(addrFrom) {
		knownNodes = append(knownNodes, addrFrom)
		fmt.Printf("There are %d known nodes now!\n", len(knownNodes))
	}
}

func handleConnection(conn net.Conn, bc *blockchain.Blockchain) {
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}
	command := bytesToCommand(request[:commandLength])
	fmt.Printf("Received %s command\n", command)
	addrFrom := fmt.Sprintf("%s:%s", strings.Split(conn.RemoteAddr().String(), ":")[0], os.Getenv("PORT"))
	addrLocal := fmt.Sprintf("%s:%s", strings.Split(conn.LocalAddr().String(), ":")[0], os.Getenv("PORT"))
	switch command {
	// case "addr":
	// 	handleAddr(request, bc)
	case "block":
		handleBlock(request, bc, addrFrom, addrLocal)
	case "inv":
		handleInv(request, bc, addrFrom)
	case "getblocks":
		handleGetBlocks(request, bc, addrFrom)
	case "getdata":
		handleGetData(request, bc, addrFrom)
	case "tx":
		handleTx(request, bc, addrFrom, addrLocal)
	case "version":
		handleVersion(request, bc, addrFrom)
	case "proposal":
		handleProposal(request, addrFrom, addrLocal)
	case "feedback":
		handleFeedback(request, addrFrom, addrLocal)
	default:
		fmt.Println("Unknown command!")
	}

	conn.Close()
	rm := helper.RemoveExpireIPFS(bc)
	if len(rm) > 0 {
		fmt.Println("Remove flie(s)!")
		for i, str := range rm {
			fmt.Printf("(%d) %s\n", i, str)

		}

	}
}

// StartServer starts a node
func StartServer(minerAddress string) {
	knownNodes = append(knownNodes, os.Getenv("KNOWNNODE"))
	miningAddress = minerAddress
	port := fmt.Sprintf(":%s", os.Getenv("PORT"))
	ln, err := net.Listen(protocol, port)
	if err != nil {
		log.Panic(err)
	}
	defer ln.Close()

	bc := blockchain.NewBlockchain()
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		os.Stderr.WriteString("Oops: " + err.Error() + "\n")
		os.Exit(1)
	}

	dif := 0
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				nodeAddress = fmt.Sprintf("%s:%s", ipnet.IP.String(), os.Getenv("PORT"))
				dif += 1

				if nodeAddress == knownNodes[0] {
					break
				}
				dif -= 1
			}
		}
	}

	if dif == 0 {
		sendVersion(knownNodes[0], bc)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Panic(err)
		}
		go handleConnection(conn, bc)
	}
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

func nodeIsKnown(addr string) bool {
	for _, node := range knownNodes {
		if node == addr {
			return true
		}
	}

	return false
}
