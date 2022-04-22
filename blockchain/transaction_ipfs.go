package blockchain

import (
	"blockchain_go/utils"
	"bytes"
	"encoding/gob"
	"log"
)

// Transaction IPFS
type TXIpfs struct {
	IpfsHash   []byte
	PubKeyHash []byte
}

// Lock signs the ipfs hash
func (t *TXIpfs) Lock(address []byte) {
	pubKeyHash := utils.Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	t.PubKeyHash = pubKeyHash
}

// IsLockedWithKey checks if the ipfs hash can be used by the owner of the pubkey
func (t *TXIpfs) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(t.PubKeyHash, pubKeyHash) == 0
}

// NewTXIpfs create a new TXIpfs
func NewTXIpfs(ipfsHash string, address string) *TXIpfs {
	txi := &TXIpfs{[]byte(ipfsHash), nil}
	txi.Lock([]byte(address))

	return txi
}

// Serialize serializes link the file of IPFS
func (t TXIpfs) SerializeIPFS() []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(&t)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

// DeserializeOutputs deserializes link the file of IPFS
func DeserializeIPFS(data []byte) TXIpfs {
	var outputs TXIpfs

	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&outputs)
	if err != nil {
		log.Panic(err)
	}

	return outputs
}
