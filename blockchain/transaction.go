package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"math/big"
	"strings"
	"time"

	"github.com/thanhxeon2470/TSS_chain/wallet"

	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

const subsidy = 10

// Transaction represents a TSS coin transaction
type Transaction struct {
	ID   []byte
	Ipfs []TXIpfs
	Vin  []TXInput
	Vout []TXOutput
}

// IsCoinbase checks whether the transaction is coinbase
func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

// Serialize returns a serialized Transaction
func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}

// Hash returns the hash of the Transaction
func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())

	return hash[:]
}

// Sign signs each input of a Transaction
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey) {
	if tx.IsCoinbase() {
		return
	}

	txCopy := tx.TrimmedCopy()

	for inID, _ := range txCopy.Vin {
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = nil

		dataToSign := fmt.Sprintf("%x\n", txCopy)
		hashToSign := sha256.Sum256([]byte(dataToSign))
		r, s, err := ecdsa.Sign(rand.Reader, &privKey, hashToSign[:])
		if err != nil {
			log.Panic(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)

		tx.Vin[inID].Signature = signature
	}
}

// String returns a human-readable representation of a transaction
func (tx Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.ID))

	for i, input := range tx.Vin {

		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:      %x", input.Txid))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Vout))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PubKey))
	}

	for i, output := range tx.Vout {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("       Script: %x", output.PubKeyHash))
	}
	if len(tx.Ipfs) > 0 {
		lines = append(lines, fmt.Sprintf("     IPFS: %s | EXP: %s", tx.Ipfs[0].IpfsHashENC, time.Unix(tx.Ipfs[0].Exp, 0)))
		for i, allowuser := range tx.Ipfs[0].PubKeyHash {
			lines = append(lines, fmt.Sprintf("       User %d:  %x", i, allowuser))
		}
	}

	return strings.Join(lines, "\n")
}

// TrimmedCopy creates a trimmed copy of Transaction to be used in signing
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput
	var ipfsList []TXIpfs

	for _, vin := range tx.Vin {
		inputs = append(inputs, TXInput{vin.Txid, vin.Vout, nil, nil})
	}

	for _, vout := range tx.Vout {
		outputs = append(outputs, TXOutput{vout.Value, vout.PubKeyHash})
	}

	// for _, ipfs := range tx.Ipfs {
	// 	ipfsList = append(ipfsList, TXIpfs{ipfs.PubKeyOwner, ipfs.SignatureFile, ipfs.IpfsHash, ipfs.PubKeyHash, ipfs.Exp})
	// }
	txCopy := Transaction{tx.ID, ipfsList, inputs, outputs}

	return txCopy
}

// Verify verifies signatures of Transaction inputs
func (tx *Transaction) Verify(bc *Blockchain) bool {
	if tx.IsCoinbase() {
		return true
	}

	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()
	UTXOSet := UTXOSet{bc}
	totalSpent := 0
	for inID, vin := range tx.Vin {
		// Check Transaction exist in UTXO

		totalSpent += UTXOSet.IsTransactionExistInUTXO(vin.Txid, wallet.HashPubKey(vin.PubKey), vin.Vout)

		// Verify signature
		// prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = nil
		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.PubKey)
		x.SetBytes(vin.PubKey[:(keyLen / 2)])
		y.SetBytes(vin.PubKey[(keyLen / 2):])

		dataToVerify := fmt.Sprintf("%x\n", txCopy)
		hashToVerify := sha256.Sum256([]byte(dataToVerify))
		rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
		if ecdsa.Verify(&rawPubKey, hashToVerify[:], &r, &s) == false {
			return false
		}
	}

	// Check total Input == total Output
	for _, vout := range tx.Vout {
		totalSpent -= vout.Value
	}

	return totalSpent >= 0
}

// NewCoinbaseTX creates a new coinbase transaction
func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		randData := make([]byte, 20)
		_, err := rand.Read(randData)
		if err != nil {
			log.Panic(err)
		}

		data = fmt.Sprintf("%x", randData)
	}

	txin := TXInput{[]byte{}, -1, nil, []byte(data)}
	txout := NewTXOutput(subsidy, to)
	tx := Transaction{nil, nil, []TXInput{txin}, []TXOutput{*txout}}
	tx.ID = tx.Hash()

	return &tx
}

// Coinbase Transaction for blockchain genesis, This is network liquidity
func NewCoinbaseTXGenesis(to, data string) *Transaction {
	if data == "" {
		randData := make([]byte, 20)
		_, err := rand.Read(randData)
		if err != nil {
			log.Panic(err)
		}

		data = fmt.Sprintf("%x", randData)
	}

	txin := TXInput{[]byte{}, -1, nil, []byte(data)}
	txout := NewTXOutput(10000, to)
	tx := Transaction{nil, nil, []TXInput{txin}, []TXOutput{*txout}}
	tx.ID = tx.Hash()

	return &tx
}

// NewUTXOTransaction creates a new transaction
func NewUTXOTransaction(w *wallet.Wallet, to string, amount int, pubKeyHashAllow, ipfsHash []byte, UTXOSet *UTXOSet) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput
	var ipfsList []TXIpfs

	pubKeyHash := wallet.HashPubKey(w.PublicKey)
	acc, validOutputs := UTXOSet.FindSpendableOutputs(pubKeyHash, amount)

	if acc < amount {
		log.Panic("ERROR: Not enough funds")
		return nil
	}

	// Build a list of inputs
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			input := TXInput{txID, out, nil, w.PublicKey}
			inputs = append(inputs, input)
		}
	}

	// Build a list of outputs
	from := fmt.Sprintf("%s", w.GetAddress())
	outputs = append(outputs, *NewTXOutput(amount, to))
	if acc > amount {
		outputs = append(outputs, *NewTXOutput(acc-amount, from)) // a change tieenf thoois
	}

	// Build a list of ipfs

	if len(ipfsHash) > 0 {
		if pubKeyHashAllow == nil {
			pubKeyHashAllow = wallet.HashPubKey(w.PublicKey)
		}
		ipfsList = append(ipfsList, *NewTXIpfs(w.PublicKey, nil, ipfsHash, pubKeyHashAllow))
		// ipfsList[0].SignIPFS(w.PrivateKey)
	}

	tx := Transaction{nil, ipfsList, inputs, outputs}
	tx.ID = tx.Hash()
	UTXOSet.Blockchain.SignTransaction(&tx, w.PrivateKey)

	return &tx
}

// DeserializeTransaction deserializes a transaction
func DeserializeTransaction(data []byte) Transaction {
	var transaction Transaction

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&transaction)
	if err != nil {
		log.Panic(err)
	}

	return transaction
}
