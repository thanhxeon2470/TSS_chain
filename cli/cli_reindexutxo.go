package cli

import (
	"fmt"

	"github.com/thanhxeon2470/testchain/blockchain"
)

func (cli *CLI) ReindexUTXO() {
	bc := blockchain.NewBlockchain()
	UTXOSet := blockchain.UTXOSet{bc}
	UTXOSet.Reindex()
	count := UTXOSet.CountTransactions()
	fmt.Printf("Done! There are %d transactions in the UTXO set.\n", count)
}
