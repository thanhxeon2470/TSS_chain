package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"math/big"
	"time"

	"github.com/thanhxeon2470/TSS_chain/wallet"
)

// Transaction IPFS
type TXIpfs struct {
	PubKeyOwner   []byte
	SignatureFile []byte

	IpfsHashENC []byte
	PubKeyHash  []byte

	Exp int64
}

// Lock signs the ipfs hash
// func (t *TXIpfs) Lock(addresses [][]byte) {
// 	for _, addr := range addresses {
// 		pubKeyHash := utils.Base58Decode(addr)
// 		pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
// 		t.PubKeyHash = append(t.PubKeyHash, pubKeyHash)
// 	}
// }

// IsLockedWithKey checks if the ipfs hash can be used by the owner of the pubkey
func (t *TXIpfs) IsLockedWithKey(pubKeyHash []byte) bool {
	if bytes.Compare(t.PubKeyHash, pubKeyHash) == 0 {
		return true
	}
	return false
}

// IsOwner check the the author
func (t *TXIpfs) IsOwner(pubKeyHash []byte) bool {
	pubKeyOwnerHash := wallet.HashPubKey(t.PubKeyOwner)

	if bytes.Equal(pubKeyOwnerHash, pubKeyHash) {
		return true
	}
	return false

}

// NewTXIpfs create a new TXIpfs
func NewTXIpfs(pubKeyOwner []byte, signatureFile, ipfsHash []byte, pubKeyHashAllows []byte) *TXIpfs {
	txi := &TXIpfs{pubKeyOwner, signatureFile, ipfsHash, pubKeyHashAllows, time.Now().Unix() + 31536000} // exp 1 year
	// var addressesByte [][]byte
	// for _, address := range addresses {
	// 	if wallet.ValidateAddress(address) {
	// 		addressesByte = append(addressesByte, []byte(address))
	// 	}
	// }
	// txi.Lock(addressesByte)

	return txi
}

func (t *TXIpfs) SignIPFS(privKey ecdsa.PrivateKey) {
	dataToSign := t.IpfsHashENC
	dataToSign = append(dataToSign, t.PubKeyHash...)

	hashToSign := sha256.Sum256(dataToSign)
	r, s, err := ecdsa.Sign(rand.Reader, &privKey, hashToSign[:])
	if err != nil {
		log.Panic(err)
	}
	signature := append(r.Bytes(), s.Bytes()...)
	t.SignatureFile = signature
}

func (t *TXIpfs) verifyIPFS() bool {
	dataToVerify := t.IpfsHashENC
	dataToVerify = append(dataToVerify, t.PubKeyHash...)

	hashToVerify := sha256.Sum256(dataToVerify)

	curve := elliptic.P256()

	r := big.Int{}
	s := big.Int{}
	sigLen := len(t.SignatureFile)
	r.SetBytes(t.SignatureFile[:(sigLen / 2)])
	s.SetBytes(t.SignatureFile[(sigLen / 2):])

	x := big.Int{}
	y := big.Int{}
	keyLen := len(t.PubKeyOwner)
	x.SetBytes(t.PubKeyOwner[:(keyLen / 2)])
	y.SetBytes(t.PubKeyOwner[(keyLen / 2):])

	rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
	if ecdsa.Verify(&rawPubKey, hashToVerify[:], &r, &s) == false {
		return false
	}
	return true
}

type TXIpfsList struct {
	TXIpfsList []TXIpfs
}

// Serialize serializes link the file of IPFS
func (t TXIpfsList) SerializeIPFS() []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(&t)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

// DeserializeOutputs deserializes link the file of IPFS
func DeserializeIPFS(data []byte) TXIpfsList {
	var res TXIpfsList

	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&res)
	if err != nil {
		log.Panic(err)
	}

	return res
}
