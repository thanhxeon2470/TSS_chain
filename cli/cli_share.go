package cli

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/thanhxeon2470/TSS_chain/blockchain"
	"github.com/thanhxeon2470/TSS_chain/helper"
	"github.com/thanhxeon2470/TSS_chain/utils"
	"github.com/thanhxeon2470/TSS_chain/wallet"

	"github.com/ethereum/go-ethereum/crypto/ecies"
)

func (cli *CLI) Share(prkFrom, to string, amount int, pubkeyallowuser string, iHashEncode string) string {
	// if !wallet.ValidateAddress(prkFrom) {
	// 	log.Panic("ERROR: Sender address is not valid")
	// }
	if !wallet.ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	// wallets, err :=
	// if err != nil {
	// 	log.Panic(err)
	// }
	w := wallet.DecodePrivKey([]byte(prkFrom))

	bc := blockchain.NewBlockchainView()
	defer bc.DB.Close()

	UTXOSet := blockchain.UTXOSet{bc}

	curve := elliptic.P256()
	pubKey := utils.Base58Decode([]byte(pubkeyallowuser))

	// Decode to get hash file
	priKey := ecies.ImportECDSA(&w.PrivateKey)
	ifh, err := hex.DecodeString(iHashEncode)
	if err != nil {
		return ""
	}
	iHash, err := priKey.Decrypt(ifh, nil, nil)
	if err != nil {
		return ""
	}
	var newIHash string = ""
	isSuccess, err := helper.IpfsGet(string(iHash))
	if err != nil {
		return ""
	}
	if isSuccess {
		source, err := os.Open(string(iHash))
		if err != nil {
			return ""
		}
		defer source.Close()

		destination, err := os.Create(string(iHash) + "copy")
		if err != nil {
			return ""
		}
		buf := make([]byte, 1024)
		for {
			n, err := source.Read(buf)
			if err != nil && err != io.EOF {
				return ""
			}
			if n == 0 {
				if _, err := destination.Write(pubKey); err != nil {
					return ""
				}
				break
			}

			if _, err := destination.Write(buf[:n]); err != nil {
				return ""
			}
		}
		newIHash, err = helper.IpfsAdd(string(iHash) + "copy")
		if err != nil {
			return ""
		}

		err = os.Remove(string(iHash))
		if err != nil {
			return ""
		}
		err = os.Remove(string(iHash) + "copy")
		if err != nil {
			return ""
		}
	} else {
		fmt.Print("Cant get file from ipfs")
		return ""
	}

	// Encode to new allow user
	if newIHash == "" {
		return ""
	} else {
		x := big.Int{}
		y := big.Int{}
		keyLen := len(pubKey)
		x.SetBytes(pubKey[:(keyLen / 2)])
		y.SetBytes(pubKey[(keyLen / 2):])
		rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}

		pubECIES := ecies.ImportECDSAPublic(&rawPubKey)

		ct, err := ecies.Encrypt(rand.Reader, pubECIES, []byte(newIHash), nil, nil)
		if err != nil {
			return ""
		}
		pubKeyHash := wallet.HashPubKey(append(rawPubKey.X.Bytes(), rawPubKey.Y.Bytes()...))
		tx := blockchain.NewUTXOTransaction(w, to, amount, pubKeyHash, hex.EncodeToString(ct), &UTXOSet)
		if tx == nil {
			fmt.Println("Fail to create transaction!")

			return ""
		}
		proposal := proposal{tx.ID, []byte(to), []byte(newIHash), amount}
		sendProposal(strings.Split(os.Getenv("KNOWNNODE"), "_")[0], proposal)
		sendTx("127.0.0.1:3000", tx)

		return hex.EncodeToString(ct)
	}
}
