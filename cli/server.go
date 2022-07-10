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
	"strings"
	"time"

	"github.com/thanhxeon2470/TSS_chain/blockchain"
	"github.com/thanhxeon2470/TSS_chain/helper"
	"github.com/thanhxeon2470/TSS_chain/rpc"
)

const protocol = "tcp"
const nodeVersion = 1
const commandLength = 12

var nodeAddress string
var miningAddress string
var StorageMiningAddress string
var knownNodes = []string{}
var blocksInTransit = [][]byte{}

var mempool = make(map[string]blockchain.Transaction)
var timeReceivedTx = make(chan int64)
var timeMining int64 = 5 // 30s

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

type txin struct {
	// AddrFrom    string
	Inputs []byte
}

type verzion struct {
	Version    int
	BestHeight int
	// AddrFrom   string
}

type Proposal struct {
	TxHash               []byte
	StorageMiningAddress []byte
	FileHash             []byte
	Amount               int
}

// feedback proposal
type Fbproposal struct {
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

func BytesToCommand(bytes []byte) string {
	var command []byte

	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}

	return fmt.Sprintf("%s", command)
}

// func extractCommand(request []byte) []byte {
// 	return request[:commandLength]
// }

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

func SendBlock(addr string, b *blockchain.Block) {
	data := block{b.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("block"), payload...)
	SendData(addr, request)
}

func SendData(addr string, data []byte) {
	conn, err := net.Dial(protocol, addr)
	if err != nil {
		fmt.Printf("%s is not available\n", addr)
		var updatedNodes []string

		for _, node := range knownNodes[1:] {
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

func SendInv(address, kind string, items [][]byte) {
	inventory := inv{kind, items}
	payload := gobEncode(inventory)
	request := append(commandToBytes("inv"), payload...)

	SendData(address, request)
}

func SendGetBlocks(address string) {
	// payload := gobEncode(getblocks{nodeAddress})
	request := commandToBytes("getblocks")

	SendData(address, request)
}

func SendGetData(address, kind string, id []byte) {
	payload := gobEncode(getdata{kind, id})
	request := append(commandToBytes("getdata"), payload...)

	SendData(address, request)
}

func SendTx(addr string, tnx *blockchain.Transaction) {
	data := tx{tnx.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("tx"), payload...)

	SendData(addr, request)
}

func SendTxIns(addr string, txins *blockchain.TXInputs) {
	data := txin{txins.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("tx"), payload...)

	SendData(addr, request)
}

func SendVersion(addr string, bc *blockchain.Blockchain) {
	bestHeight := bc.GetBestHeight()
	payload := gobEncode(verzion{nodeVersion, bestHeight})
	request := append(commandToBytes("version"), payload...)

	SendData(addr, request)
}

func SendProposal(addr string, pps Proposal) {
	payload := gobEncode(pps)
	request := append(commandToBytes("proposal"), payload...)
	SendData(addr, request)
}

func SendFBProposal(addr string, txid []byte, feedback bool) {
	payload := gobEncode(Fbproposal{txid, feedback})
	request := append(commandToBytes("feedback"), payload...)
	SendData(addr, request)
}

func handleProposal(request []byte, addrFrom, addrLocal string) {
	var buff bytes.Buffer
	var payload Proposal
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
			isSuccess, err := helper.IpfsGet(fh)
			if err != nil {
				return
			}
			if isSuccess {
				// And update this file to ipfs cluster
				_, err := helper.IpfsClusterAdd(fh)
				if err != nil {
					return
				}
				err = os.Remove(fh)
				if err != nil {
					return
				}
			} else {
				fmt.Print("Cant get file from ipfs")
			}

			SendFBProposal(addrFrom, payload.TxHash, true)
			return
		}
	}
	if addrLocal == knownNodes[0] {
		if !nodeIsKnown(addrFrom) {
			knownNodes = append(knownNodes, addrFrom)
			fmt.Printf("There are %d known nodes now!\n", len(knownNodes))
		}
		for _, node := range knownNodes {
			if node != addrLocal && node != addrFrom {
				SendProposal(node, payload)
			}
		}
	}
}

func handleFeedback(request []byte, addrFrom, addrLocal string) {
	var buff bytes.Buffer
	var payload Fbproposal
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
				SendFBProposal(node, payload.TxHash, payload.Accept)
			}
		}
	}
	if len(mempool) > 0 {
		// When received feedback proposal =>>> check this and send transaction
		for id := range mempool {
			if id == hex.EncodeToString(payload.TxHash) {
				tx := mempool[id]
				SendTx(knownNodes[0], &tx)
				delete(mempool, id)

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
				SendBlock(node, block)
				fmt.Printf("This block is broadcasted to %s\n", node)
			}
		}
	}

	if len(blocksInTransit) > 0 {
		blockHash := blocksInTransit[0]
		SendGetData(addrFrom, "block", blockHash)

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
		SendGetData(addrFrom, "block", blockHash)

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
			SendGetData(addrFrom, "tx", txID)
		}
	}
}

func handleGetBlocks(request []byte, bc *blockchain.Blockchain, addrFrom string) {
	blocks := bc.GetBlockHashes()
	SendInv(addrFrom, "block", blocks)
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

		SendBlock(addrFrom, &block)
	}

	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID)
		tx := mempool[txID]

		SendTx(addrFrom, &tx)
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
				SendInv(node, "tx", [][]byte{tx.ID})
			}
		}
	}
	// root lamf wallet app chua chuyen file di duocj
	fmt.Println("Time receive tx...", time.Now().Unix())

	timeReceivedTx <- time.Now().Unix()

}

// After 30s, if less than 3 txs block will be mined
func MiningBlock(bc *blockchain.Blockchain, timeStart chan int64) {
	for {

		t := <-timeStart
		fmt.Println("Wait for mine...", t)
		for {
			timeNow := time.Now().Unix()
			if len(miningAddress) > 0 && len(mempool) >= 1 && (len(mempool) >= 3 || timeNow-t > timeMining) {
				fmt.Println("Mined...", timeNow)
			MineTransactions:
				var txs []*blockchain.Transaction

				for _, tx := range mempool {
					if bc.VerifyTransaction(&tx) {
						txs = append(txs, &tx)
					}
					txID := hex.EncodeToString(tx.ID)
					delete(mempool, txID)
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
					SendInv(node, "block", [][]byte{newBlock.Hash})
				}

				if len(mempool) > 0 {
					goto MineTransactions
				}
				break
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
		SendGetBlocks(addrFrom)
	} else if myBestHeight > foreignerBestHeight {
		SendVersion(addrFrom, bc)
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
	command := BytesToCommand(request[:commandLength])
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
	knownNodes = strings.Split(os.Getenv("KNOWNNODE"), "_")
	miningAddress = minerAddress
	port := fmt.Sprintf(":%s", os.Getenv("PORT"))
	ln, err := net.Listen(protocol, port)
	if err != nil {
		log.Panic(err)
	}
	defer ln.Close()
	fmt.Println("Blockchain is listening at port ", port)

	portRPC := fmt.Sprintf(":%s", os.Getenv("PORT_RPC"))
	lnRpc, err := net.Listen(protocol, portRPC)
	if err != nil {
		log.Panic(err)
	}
	defer lnRpc.Close()
	fmt.Println("Blockchain RPC is listening at port ", portRPC)

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
		SendVersion(knownNodes[0], bc)
	}
	if len(minerAddress) > 0 {
		// timeStartnode <- time.Now().Unix()
		go MiningBlock(bc, timeReceivedTx)
	}

	go handleRPC(lnRpc)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Panic(err)
		}
		go handleConnection(conn, bc)

	}
}

func handleRPC(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Panic(err)
		}
		go rpc.HandleRPC(conn)
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
