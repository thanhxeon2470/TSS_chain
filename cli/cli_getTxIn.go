package cli

import (
	"encoding/hex"
	"log"

	"github.com/thanhxeon2470/TSS_chain/blockchain"
	"github.com/thanhxeon2470/TSS_chain/utils"
	"github.com/thanhxeon2470/TSS_chain/wallet"
)

func (cli *CLI) GetTxIn(addr string, amount int) blockchain.TXInputs {
	if !wallet.ValidateAddress(addr) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	bc := blockchain.NewBlockchainView()
	defer bc.DB.Close()

	UTXOSet := blockchain.UTXOSet{bc}

	pubKeyHash := utils.Base58Decode([]byte(addr))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-wallet.AddressChecksumLen]

	acc, validOutputs := UTXOSet.FindSpendableOutputs(pubKeyHash, amount)

	res := blockchain.TXInputs{nil}
	if acc < amount {
		log.Panic("ERROR: Not enough funds")
		return res
	}
	// Build a list of inputs
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			input := blockchain.TXInput{txID, out, nil, nil}
			res.Inputs = append(res.Inputs, input)

			// fmt.Println(hex.EncodeToString(txID), " ==== ", out)
		}
	}

	return res

}
