package cli

import (
	"blockchain_go/blockchain"
	"fmt"
)

func (cli *CLI) reindexUTXO(nodeID string) {
	bc := blockchain.NewBlockchain(nodeID)
	UTXOSet := blockchain.UTXOSet{bc}
	UTXOSet.Reindex()

	count := UTXOSet.CountTransactions()
	fmt.Printf("Done! There are %d transactions in the UTXO set.\n", count)
}
