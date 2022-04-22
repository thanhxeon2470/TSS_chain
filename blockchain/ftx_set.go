package blockchain

import (
	// "blockchain_go/blockchain"
	"log"

	"github.com/boltdb/bolt"
)

const ftxBucket = "filealive"

// FTX is File transaction
type FTXset struct {
	Blockchain *Blockchain
}

func (f FTXset) FindFTX(pubKeyHash []byte) []string {
	var listFTX []string
	db := f.Blockchain.db
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ftxBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			out := DeserializeIPFS(v)
			if out.IsLockedWithKey(pubKeyHash) {
				listFTX = append(listFTX, string(out.IpfsHash))
			}
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return listFTX

}

func UpdateFTX() {

}
