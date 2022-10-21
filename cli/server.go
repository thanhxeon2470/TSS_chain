package cli

import (
	"bytes"

	"context"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"os/signal"

	"strconv"
	"strings"
	"time"

	"github.com/thanhxeon2470/TSS_chain/blockchain"
	"github.com/thanhxeon2470/TSS_chain/helper"
	"github.com/thanhxeon2470/TSS_chain/p2p"
	"github.com/thanhxeon2470/TSS_chain/rpc"
)

const protocol = "tcp"
const nodeVersion = 1
const commandLength = 12

var nodeAddress string

var miningAddress string
var StorageMiningAddress string
var bootsNodes = []string{}
var blocksInTransit = [][]byte{}

var mempool = make(map[string]blockchain.Transaction)
var proposalPool = make(map[string]bool)

// var timeReceivedTx = make(chan int64)
var timeMining int64 = 5 // 30s

type block struct {
	// AddrFrom string
	Block []byte
}

type getblocks struct {
	// AddrFrom string
}

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
	// AddrFrom   string
	Version    int
	BestHeight int
}

type Proposal struct {
	// AddrFrom             string
	TxHash               []byte
	StorageMiningAddress []byte
	FileHash             []byte
	Amount               int
}

// feedback proposal
type Fbproposal struct {
	// AddrFrom string
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

	return string(command)
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

func SendBlock(b *blockchain.Block) {
	data := block{b.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("block"), payload...)
	SendData(request)
}

func SendData(data []byte) {
	p2p.Send2Peers(data)
}

func SendInv(kind string, items [][]byte) {
	inventory := inv{kind, items}
	payload := gobEncode(inventory)
	request := append(commandToBytes("inv"), payload...)

	SendData(request)
}

func SendGetBlocks() {
	payload := gobEncode(getblocks{})
	request := append(commandToBytes("getblocks"), payload...)

	SendData(request)
}

func SendGetData(kind string, id []byte) {
	payload := gobEncode(getdata{kind, id})
	request := append(commandToBytes("getdata"), payload...)

	SendData(request)
}

func SendTx(tnx *blockchain.Transaction) {
	data := tx{tnx.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("tx"), payload...)

	SendData(request)
}

// func SendTxIns(addr string, txins *blockchain.TXInputs) {
// 	data := txin{txins.Serialize()}
// 	payload := gobEncode(data)
// 	request := append(commandToBytes("tx"), payload...)

// 	SendData(addr, request)
// }

func SendVersion(bc *blockchain.Blockchain) {
	bestHeight := bc.GetBestHeight()
	payload := gobEncode(verzion{nodeVersion, bestHeight})
	request := append(commandToBytes("version"), payload...)

	SendData(request)
}

func SendProposal(pps Proposal) {
	payload := gobEncode(pps)
	request := append(commandToBytes("proposal"), payload...)
	SendData(request)
}

func SendFBProposal(txid []byte, feedback bool) {
	payload := gobEncode(Fbproposal{txid, feedback})
	request := append(commandToBytes("feedback"), payload...)
	SendData(request)
}

func handleProposal(request []byte) {
	var buff bytes.Buffer
	var payload Proposal
	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	txid := hex.EncodeToString(payload.TxHash)
	if proposalPool[txid] {
		fmt.Println("This proposal is existed!")
		return
	}
	// add file to ipfs when receive tx
	if len(StorageMiningAddress) > 0 {
		if bytes.Equal(payload.StorageMiningAddress, []byte(StorageMiningAddress)) {
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

			SendFBProposal(payload.TxHash, true)
			return
		}
	}
	// proposalPool[txid] = true
	// // if addrLocal == knownNodes[0] {
	// if !nodeIsKnown(payload.AddrFrom) {
	// 	bootsNodes = append(bootsNodes, payload.AddrFrom)
	// 	fmt.Printf("There are %d known nodes now!\n", len(bootsNodes))
	// }
	// for _, node := range bootsNodes {
	// 	if node != nodeIP && node != payload.AddrFrom {
	// 		SendProposal(node, payload)
	// 	}
	// }
	// }
}

func handleFeedback(request []byte) {
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
	txid := hex.EncodeToString(payload.TxHash)
	if !proposalPool[txid] {
		log.Println("Not exist this proposal")
		return
	}

	delete(proposalPool, txid)

	// // if addrLocal == knownNodes[0] {
	// for _, node := range bootsNodes {
	// 	if node != nodeIP && node != payload.AddrFrom {
	// 		SendFBProposal(node, payload.TxHash, payload.Accept)
	// 	}
	// }
	// // }
	// if len(mempool) > 0 {
	// 	// When received feedback proposal =>>> check this and send transaction
	// 	for id := range mempool {
	// 		if id == hex.EncodeToString(payload.TxHash) {
	// 			tx := mempool[id]
	// 			SendTx(bootsNodes[0], &tx)
	// 			delete(mempool, id)

	// 		}
	// 	}
	// }
	// if !nodeIsKnown(payload.AddrFrom) {
	// 	bootsNodes = append(bootsNodes, payload.AddrFrom)
	// 	fmt.Printf("There are %d known nodes now!\n", len(bootsNodes))
	// }
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

func handleBlock(request []byte, bc *blockchain.Blockchain) {
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
	if bc.IsBlockExist(block.Hash) {
		fmt.Println("Recevied a block! But it's existed")

		return
	}
	fmt.Println("Recevied a new block!")
	bc.AddBlock(block)
	UTXOSet := blockchain.UTXOSet{Blockchain: bc}
	FTXSet := blockchain.FTXset{Blockchain: bc}
	UTXOSet.Reindex()
	FTXSet.ReindexFTX()
	fmt.Printf("Added block %x\n", block.Hash)

	first := len(blocksInTransit)
	if first > 0 {
		blockHash := blocksInTransit[first-1]
		SendGetData("block", blockHash)

		blocksInTransit = blocksInTransit[:first-1]

	}

	// // if localAddr == knownNodes[0] {
	// for _, node := range bootsNodes {
	// 	if node != nodeIP && node != payload.AddrFrom {
	// SendBlock("node", block)
	// 		fmt.Printf("This block is broadcasted to %s\n", node)
	// 	}
	// }
	// // }
	// if !nodeIsKnown(payload.AddrFrom) {
	// 	bootsNodes = append(bootsNodes, payload.AddrFrom)
	// 	fmt.Printf("There are %d known nodes now!\n", len(bootsNodes))
	// }
}

func handleInv(request []byte, bc *blockchain.Blockchain) {
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
		blocksInTransit = [][]byte{}
		for _, hash := range payload.Items {
			if !bc.IsBlockExist(hash) {
				blocksInTransit = append(blocksInTransit, hash)
			}
		}

		if len(blocksInTransit) > 0 {
			first := len(blocksInTransit) - 1
			blockHash := blocksInTransit[first]
			SendGetData("block", blockHash)

			blocksInTransit = blocksInTransit[:first]
		}
	}

	if payload.Type == "tx" {
		txID := payload.Items[0]

		if mempool[hex.EncodeToString(txID)].ID == nil {
			SendGetData("tx", txID)
		}
	}
	// if !nodeIsKnown(payload.AddrFrom) {
	// 	bootsNodes = append(bootsNodes, payload.AddrFrom)
	// 	fmt.Printf("There are %d known nodes now!\n", len(bootsNodes))
	// }
}

func handleGetBlocks(request []byte, bc *blockchain.Blockchain) {
	var buff bytes.Buffer
	var payload getblocks

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	blocks := bc.GetBlockHashes()
	SendInv("block", blocks)
	// if !nodeIsKnown(payload.AddrFrom) {
	// 	bootsNodes = append(bootsNodes, payload.AddrFrom)
	// 	fmt.Printf("There are %d known nodes now!\n", len(bootsNodes))
	// }
}

func handleGetData(request []byte, bc *blockchain.Blockchain) {
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

		SendBlock(&block)
	}

	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID)
		tx := mempool[txID]

		SendTx(&tx)
		// delete(mempool, txID)
	}
	// if !nodeIsKnown(payload.AddrFrom) {
	// 	bootsNodes = append(bootsNodes, payload.AddrFrom)
	// 	fmt.Printf("There are %d known nodes now!\n", len(bootsNodes))
	// }
}

func handleTx(request []byte, bc *blockchain.Blockchain) {
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

	// if addrLocal == knownNodes[0] {
	// for _, node := range bootsNodes {
	// 	if node != nodeIP && node != payload.AddrFrom {
	// 		fmt.Printf("This transaction will be broadcasted to %s\n", node)
	// 		SendInv(node, "tx", [][]byte{tx.ID})
	// 	}
	// }
	// }
	// root lamf wallet app chua chuyen file di duocj
	fmt.Println("Time receive tx...", time.Now().Unix())

	// timeReceivedTx <- time.Now().Unix()
	// if !nodeIsKnown(payload.AddrFrom) {
	// 	bootsNodes = append(bootsNodes, payload.AddrFrom)
	// 	fmt.Printf("There are %d known nodes now!\n", len(bootsNodes))
	// }
}

// After 30s, if less than 3 txs block will be mined
func MiningBlock(bc *blockchain.Blockchain) {
	t := time.NewTicker(time.Duration(timeMining) * time.Second)
	for {
		// t := <-timeStart
		fmt.Println("Wait for mine...")
		<-t.C
		for {

			timeNow := time.Now().Unix()
			// if len(miningAddress) > 0 && len(mempool) >= 1 && (len(mempool) >= 3 || timeNow-t > timeMining) {
			if len(miningAddress) > 0 {
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
					break
				}

				cbTx := blockchain.NewCoinbaseTX(miningAddress, "")
				txs = append(txs, cbTx)

				newBlock := bc.MineBlock(txs)
				if newBlock == nil {
					break
				}
				UTXOSet := blockchain.UTXOSet{Blockchain: bc}
				FTXSet := blockchain.FTXset{Blockchain: bc}
				UTXOSet.Reindex()
				FTXSet.ReindexFTX()

				fmt.Println("New block is mined!")

				for _, tx := range txs {
					txID := hex.EncodeToString(tx.ID)
					delete(mempool, txID)
				}

				// for _, node := range bootsNodes {
				SendInv("block", [][]byte{newBlock.Hash})
				// }

				if len(mempool) > 0 {
					goto MineTransactions
				}
				break
			}
		}
	}

}

func handleVersion(request []byte, bc *blockchain.Blockchain) {
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
		SendGetBlocks()
	} else if myBestHeight > foreignerBestHeight {
		SendVersion(bc)
	}
	// sendAddr(addrFrom)
	// if !nodeIsKnown(payload.AddrFrom) {
	// 	bootsNodes = append(bootsNodes, payload.AddrFrom)
	// 	fmt.Printf("There are %d known nodes now!\n", len(bootsNodes))
	// }
}

func handleConnection(request []byte, bc *blockchain.Blockchain) {
	command := BytesToCommand(request[:commandLength])
	fmt.Printf("Received %s command\n", command)

	switch command {
	// case "addr":
	// 	handleAddr(request, bc)
	case "block":
		handleBlock(request, bc)
	case "inv":
		handleInv(request, bc)
	case "getblocks":
		handleGetBlocks(request, bc)
	case "getdata":
		handleGetData(request, bc)
	case "tx":
		handleTx(request, bc)
	case "version":
		handleVersion(request, bc)
	case "proposal":
		handleProposal(request)
	case "feedback":
		handleFeedback(request)
	default:
		fmt.Println("Unknown command!")
	}

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
	nodes := os.Getenv("BOOTSNODES")
	if nodes == "" {
		fmt.Printf("BOOTSNODES env. var is not set!")
		os.Exit(1)
	}
	bootsNodes = strings.Split(nodes, "_")

	portStr := os.Getenv("P2P_PORT")
	port := 0
	if portStr == "" {
		fmt.Printf("P2P_PORT env. var is not set! (default random)")
		os.Exit(1)
	} else {
		p, err := strconv.Atoi(portStr)
		if err != nil {
			panic(err)
		}
		port = p
	}

	// nodeIP = thisNode
	miningAddress = minerAddress
	// port := strings.Split(thisNode, ":")[1]
	// port = fmt.Sprintf(":%s", port)
	// ln, err := net.Listen(protocol, port)
	// if err != nil {
	// 	log.Panic(err)
	// }
	// defer ln.Close()
	// fmt.Println("Blockchain is listening at port ", port)
	p2p.InitP2P(port, bootsNodes, true)
	bc := blockchain.NewBlockchain()

	s := rpc.InitJSONRPCServer(":8332")
	// bcv:= blockchain.NewBlockchainView()

	ctxbc := context.WithValue(context.Background(), rpc.Bckey, bc)
	ctx, cancel := signal.NotifyContext(ctxbc, os.Interrupt, os.Kill)
	// ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()
	fmt.Println("RPC server(standard) running on Port: 8332")
	if err := s.Run(ctx); err != nil {
		log.Fatal(err)
	}

	// for _, node := range bootsNodes {
	SendVersion(bc)
	// }
	if len(minerAddress) > 0 {
		// timeStartnode <- time.Now().Unix()
		go MiningBlock(bc)
	}

	go rpc.HandleRPC(bc)

	for {
		data2handle := <-p2p.Data2Handle
		go handleConnection(data2handle, bc)

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
