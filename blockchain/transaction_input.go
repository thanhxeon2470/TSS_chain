package blockchain

import (
	"bytes"
	"encoding/gob"
	"log"

	"github.com/thanhxeon2470/TSS_chain/wallet"
)

// TXInput represents a transaction input
type TXInput struct {
	Txid      []byte
	Vout      int
	Signature []byte
	PubKey    []byte
}

// UsesKey checks whether the address initiated the transaction
func (in *TXInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := wallet.HashPubKey(in.PubKey)

	return bytes.Compare(lockingHash, pubKeyHash) == 0
}

type TXInputs struct {
	Inputs []TXInput
}

// Serialize serializes TXInputs
func (ins TXInputs) Serialize() []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(ins)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

// Deserialize deserializes TXInputs
func DeserializeInputs(data []byte) TXInputs {
	var inputs TXInputs

	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&inputs)
	if err != nil {
		log.Panic(err)
	}

	return inputs
}
