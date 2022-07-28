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
	var payload Findipfs

	err := GobDecode(args.Req, &payload)
	if err != nil {
		return err
	}
	bc := blockchain.NewBlockchainView()
	FTXSet := blockchain.FTXset{bc}
	defer bc.DB.Close()
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

	bc := blockchain.NewBlockchainView()
	defer bc.DB.Close()

	UTXOSet := blockchain.UTXOSet{bc}

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
	res.Res, err = GobEncode(Balance{balance, FTXs})
	return err
}

func HandleRPC() {
	rpcRequest := new(RPC)
	rpc.Register(rpcRequest)
	rpc.HandleHTTP()
	portRPC := fmt.Sprintf(":%s", os.Getenv("PORT_RPC"))
	fmt.Println("Blockchain RPC is listening at port ", portRPC)
	log.Panic(http.ListenAndServe(portRPC, nil))
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

// feedback proposal
type Fbproposal struct {
	TxHash []byte
	Accept bool
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
