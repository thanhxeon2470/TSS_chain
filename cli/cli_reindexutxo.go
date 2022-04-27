package cli

import (
	"fmt"
	"testchain/blockchain"
)

func (cli *CLI) ReindexUTXO() {
	bc := blockchain.NewBlockchain()
	UTXOSet := blockchain.UTXOSet{bc}
	UTXOSet.Reindex()
	FTXSet := blockchain.FTXset{bc}
	FTXSet.ReindexFTX()
	count := UTXOSet.CountTransactions()
	fmt.Printf("Done! There are %d transactions in the UTXO set.\n", count)
}
