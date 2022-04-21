package transactions

import (
	"bytes"
	"encoding/gob"
	"log"
)

// Transaction IPFS
type TXIpfs struct {
	IpfsHash   []byte
	PubKeyHash []byte
}

// IsLockedWithKey checks if the output can be used by the owner of the pubkey
func (out *TXIpfs) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
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
func DeserializeOutputsIPFS(data []byte) TXIpfs {
	var outputs TXIpfs

	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&outputs)
	if err != nil {
		log.Panic(err)
	}

	return outputs
}
