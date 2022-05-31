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
	"os/exec"
	"strings"

	"github.com/thanhxeon2470/TSS_chain/blockchain"
	"github.com/thanhxeon2470/TSS_chain/wallet"

	"github.com/ethereum/go-ethereum/crypto/ecies"
)

func (cli *CLI) Share(prkFrom, to string, amount int, pubkeyallowuser string, iHashEncode []byte) []byte {
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
	UTXOSet := blockchain.UTXOSet{bc}

	curve := elliptic.P256()
	pubKey, err := hex.DecodeString(pubkeyallowuser)
	if err != nil {
		return nil
	}

	// Decode to get hash file
	priKey := ecies.ImportECDSA(&w.PrivateKey)
	iHash, err := priKey.Decrypt(iHashEncode, nil, nil)
	var newIHash string = ""
	getCMD := exec.Command("ipfs", "get", string(iHash))
	stdout, err := getCMD.Output()
	if err != nil {
		return nil
	}
	str := string(stdout)
	if strings.Contains(str, string(iHash)) {
		source, err := os.Open(string(iHash))
		if err != nil {
			return nil
		}
		defer source.Close()

		destination, err := os.Create(string(iHash) + "copy")
		if err != nil {
			return nil
		}
		buf := make([]byte, 1024)
		for {
			n, err := source.Read(buf)
			if err != nil && err != io.EOF {
				return nil
			}
			if n == 0 {
				if _, err := destination.Write(pubKey); err != nil {
					return nil
				}
				break
			}

			if _, err := destination.Write(buf[:n]); err != nil {
				return nil
			}
		}
		getCMD := exec.Command("ipfs", "add", string(iHash)+"copy")
		stdout, err := getCMD.Output()
		if err != nil {
			return nil
		}
		str := string(stdout)
		if strings.Contains(str, "added") {
			newIHash = strings.Split(str, " ")[1]
		}

		// err = os.Remove(string(iHash))
		// if err != nil {
		// 	return nil
		// }
		// err = os.Remove(string(iHash) + "copy")
		// if err != nil {
		// 	return nil
		// }
	} else {
		fmt.Print("Cant get file from ipfs")
		return nil
	}

	// Encode to new apllow user
	if newIHash != "" {
		return nil
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
			return nil
		}
		pubKeyHash := wallet.HashPubKey(append(rawPubKey.X.Bytes(), rawPubKey.Y.Bytes()...))
		tx := blockchain.NewUTXOTransaction(w, to, amount, pubKeyHash, hex.EncodeToString(ct), &UTXOSet)
		if tx == nil {
			fmt.Println("Fail to create transaction!")

			return nil
		}
		proposal := proposal{tx.ID, []byte(to), []byte(newIHash), amount}
		sendProposal(strings.Split(os.Getenv("KNOWNNODE"), "_")[0], proposal)
		sendTx("127.0.0.1:3000", tx)

		return ct
	}
}
