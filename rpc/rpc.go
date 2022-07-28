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

func (r *RPC) FindIPFS(args *Args, res *Result) error {
	var buff bytes.Buffer
	var payload findipfs

	buff.Write(args.Req)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		return err
	}

	bc := blockchain.NewBlockchainView()
	FTXSet := blockchain.FTXset{bc}
	defer bc.DB.Close()
	listUser := FTXSet.FindIPFS(payload.IpfsHashENC)
	data := Ipfs{listUser}
	res.Res = gobEncode(data)
	return nil
}

func (r *RPC) GetTxIns(args *Args, res *Result) error {
	var buff bytes.Buffer
	var payload gettxins

	buff.Write(args.Req)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	address := payload.Addr
	if !wallet.ValidateAddress(address) {
		return fmt.Errorf("ERROR: Recipient address is not valid")
	}

	bc := blockchain.NewBlockchainView()
	defer bc.DB.Close()

	UTXOSet := blockchain.UTXOSet{bc}

	pubKeyHash := utils.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-wallet.AddressChecksumLen]

	validOutputs := UTXOSet.FindAllSpendableOutputs(pubKeyHash)

	data := Txins{validOutputs}
	res.Res = gobEncode(data)
	return nil
}

func (r *RPC) GetBlance(args *Args, res *Result) error {
	var buff bytes.Buffer
	var payload getbalance

	buff.Write(args.Req)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	address := payload.Addr
	if !wallet.ValidateAddress(address) {
		return fmt.Errorf("ERROR: Recipient address is not valid")
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
	res.Res = gobEncode(Balance{balance, FTXs})
	return nil
}

func HandleRPC() {
	rpcRequest := new(RPC)
	rpc.Register(rpcRequest)
	rpc.HandleHTTP()
	portRPC := fmt.Sprintf(":%s", os.Getenv("PORT_RPC"))
	fmt.Println("Blockchain RPC is listening at port ", portRPC)
	log.Panic(http.ListenAndServe(portRPC, nil))
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
	ValidOutputs map[string][][2]int
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

func gobEncode(data interface{}) []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}
