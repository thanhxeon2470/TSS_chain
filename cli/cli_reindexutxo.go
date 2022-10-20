package cli

import (
	"fmt"

	"github.com/thanhxeon2470/TSS_chain/blockchain"
)

func (cli *CLI) ReindexUTXO() {
	bc := blockchain.NewBlockchain()
	UTXOSet := blockchain.UTXOSet{Blockchain: bc}
	FTXSet := blockchain.FTXset{Blockchain: bc}
	UTXOSet.Reindex()
	FTXSet.ReindexFTX()
	count := UTXOSet.CountTransactions()
	fmt.Printf("Done! There are %d transactions in the UTXO set.\n", count)
}
