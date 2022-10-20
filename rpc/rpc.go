package rpc

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"net/rpc"
	"os"

	"github.com/thanhxeon2470/TSS_chain/blockchain"
	"github.com/thanhxeon2470/TSS_chain/p2p"
	"github.com/thanhxeon2470/TSS_chain/utils"
	"github.com/thanhxeon2470/TSS_chain/wallet"
)

type Args struct {
	Req []byte
}

type Result struct {
	Res []byte
}

type RPC struct{}

var bc *blockchain.Blockchain

func (r *RPC) FindIPFS(args *Args, res *Result) error {
	var payload Findipfs

	err := GobDecode(args.Req, &payload)
	if err != nil {
		return err
	}
	// bc := blockchain.NewBlockchainView()
	FTXSet := blockchain.FTXset{Blockchain: bc}
	listUser := FTXSet.FindIPFS(payload.IpfsHashENC)
	data := Ipfs{listUser}
	res.Res, err = GobEncode(data)
	return err
}

func (r *RPC) GetTxIns(args *Args, res *Result) error {
	var payload Gettxins

	err := GobDecode(args.Req, &payload)
	if err != nil {
		return err
	}

	address := payload.Addr
	if !wallet.ValidateAddress(address) {
		return fmt.Errorf("ERROR: Recipient address is not valid")
	}

	// bc := blockchain.NewBlockchainView()
	UTXOSet := blockchain.UTXOSet{Blockchain: bc}

	pubKeyHash := utils.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-wallet.AddressChecksumLen]

	validOutputs := UTXOSet.FindAllSpendableOutputs(pubKeyHash)

	data := Txins{validOutputs}
	res.Res, err = GobEncode(data)
	return err
}

func (r *RPC) GetBlance(args *Args, res *Result) error {
	var payload Getbalance

	err := GobDecode(args.Req, &payload)
	if err != nil {
		return err
	}

	address := payload.Addr
	if !wallet.ValidateAddress(address) {
		return fmt.Errorf("ERROR: Recipient address is not valid")
	}
	// bc := blockchain.NewBlockchainView()
	UTXOSet := blockchain.UTXOSet{Blockchain: bc}
	FTXSet := blockchain.FTXset{Blockchain: bc}

	balance := 0
	pubKeyHash := utils.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUTXO(pubKeyHash)
	FTXs := FTXSet.FindFTX(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}
	res.Res, err = GobEncode(Balance{balance, FTXs})
	return err
}

func (r *RPC) SendProposal(args *Args, res *Result) error {
	var payload Proposal

	err := GobDecode(args.Req, &payload)
	if err != nil {
		return err
	}

	pps := ProposalBlockchain{
		"thisNode",
		payload.TxHash,
		payload.StorageMiningAddress,
		payload.FileHash,
		payload.Amount}

	payloadSend, err := GobEncode(pps)
	if err != nil {
		return err
	}
	request := append(commandToBytes("proposal"), payloadSend...)
	SendData(request)

	return nil
}

func (r *RPC) SendTx(args *Args, res *Result) error {
	var payload blockchain.Transaction

	err := GobDecode(args.Req, &payload)
	if err != nil {
		return err
	}

	data := tx{"thisNode", payload.Serialize()}

	payloadSend, err := GobEncode(data)
	if err != nil {
		return err
	}
	request := append(commandToBytes("tx"), payloadSend...)

	SendData(request)
	return nil
}

func HandleRPC(blockchain *blockchain.Blockchain) {
	bc = blockchain
	rpcRequest := new(RPC)
	rpc.Register(rpcRequest)
	rpc.HandleHTTP()
	portRpc := os.Getenv("PORT_RPC")
	if portRpc == "" {
		fmt.Printf("PORT_RPC env. var is not set!")
		os.Exit(1)
	}
	portRPC := fmt.Sprintf(":%s", portRpc)
	fmt.Println("Blockchain RPC is listening at port ", portRPC)
	log.Panic(http.ListenAndServe(portRPC, nil))
}

type tx struct {
	AddrFrom    string
	Transaction []byte
}

type Findipfs struct {
	IpfsHashENC []byte
}
type Ipfs struct {
	User map[string]bool
}

type Getbalance struct {
	Addr string
}
type Balance struct {
	Value int
	FTXs  map[string]blockchain.InfoIPFS
}

type Gettxins struct {
	Addr string
}
type Txins struct {
	ValidOutputs map[string][][2]int
}

type Proposal struct {
	TxHash               []byte
	StorageMiningAddress []byte
	FileHash             []byte
	Amount               int
}
type ProposalBlockchain struct {
	AddrFrom             string
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

const commandLength = 12

func commandToBytes(command string) []byte {
	var bytes [commandLength]byte

	for i, c := range command {
		bytes[i] = byte(c)
	}

	return bytes[:]
}

func SendData(data []byte) {
	p2p.Send2Peers(data)
}

func GobEncode(data interface{}) ([]byte, error) {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func GobDecode(data []byte, payload interface{}) error {
	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(payload)
	if err != nil {
		return err
	}
	return nil
}
