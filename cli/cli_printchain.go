package cli

import (
	"blockchain_go/blockchain"
	"fmt"
	"strconv"
)

func (cli *CLI) PrintChain() {
	bc := blockchain.NewBlockchain()
	defer bc.DB.Close()

	bci := bc.Iterator()

	for {
		blk := bci.Next()

		fmt.Printf("============ Block %x ============\n", blk.Hash)
		fmt.Printf("Height: %d\n", blk.Height)
		fmt.Printf("Prev. block: %x\n", blk.PrevBlockHash)
		pow := blockchain.NewProofOfWork(blk)
		fmt.Printf("PoW: %s\n\n", strconv.FormatBool(pow.Validate()))
		for _, tx := range blk.Transactions {
			fmt.Println(tx)
		}
		fmt.Printf("\n\n")

		if len(blk.PrevBlockHash) == 0 {
			break
		}
	}
}
