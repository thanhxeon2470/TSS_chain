package blockchain

import (
	// "github.com/thanhxeon2470/testchain/blockchain"
	"encoding/hex"
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
	db := f.Blockchain.DB
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ftxBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			ipfsList := DeserializeIPFS(v)

			for _, ipfs := range ipfsList.TXIpfsList {
				if ipfs.IsLockedWithKey(pubKeyHash) {
					listFTX = append(listFTX, string(ipfs.IpfsHash))
				}
			}

		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return listFTX

}

func (f FTXset) ReindexFTX() {
	db := f.Blockchain.DB
	bucketName := []byte(ftxBucket)

	// Renew bucket
	err := db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket(bucketName)
		if err != nil && err != bolt.ErrBucketNotFound {
			log.Panic(err)
		}

		_, err = tx.CreateBucket(bucketName)
		if err != nil {
			log.Panic(err)
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	FTX := f.Blockchain.FindFTX()

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)

		for txID, outs := range FTX {
			key, err := hex.DecodeString(txID)
			if err != nil {
				log.Panic(err)
			}

			err = b.Put(key, outs.SerializeIPFS())
			if err != nil {
				log.Panic(err)
			}
		}

		return nil
	})
}

// Update updates the UTXO set with transactions from the Block
// The Block is considered to be the tip of a blockchain
func (f FTXset) UpdateFTX(block *Block) {
	db := f.Blockchain.DB

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ftxBucket))

		for _, tx := range block.Transactions {
			// if tx.IsCoinbase() == false {
			// 	for _, vin := range tx.Vin {
			// 		updatedOuts := TXOutputs{}
			// 		outsBytes := b.Get(vin.Txid)
			// 		outs := DeserializeOutputs(outsBytes)

			// 		for outIdx, out := range outs.Outputs {
			// 			if outIdx != vin.Vout {
			// 				updatedOuts.Outputs = append(updatedOuts.Outputs, out)
			// 			}
			// 		}

			// 		if len(updatedOuts.Outputs) == 0 {
			// 			err := b.Delete(vin.Txid)
			// 			if err != nil {
			// 				log.Panic(err)
			// 			}
			// 		} else {
			// 			err := b.Put(vin.Txid, updatedOuts.Serialize())
			// 			if err != nil {
			// 				log.Panic(err)
			// 			}
			// 		}

			// 	}
			// }

			newIPFSList := TXIpfsList{}
			for _, ipfs := range tx.Ipfs {
				newIPFSList.TXIpfsList = append(newIPFSList.TXIpfsList, ipfs)
			}

			err := b.Put(tx.ID, newIPFSList.SerializeIPFS())
			if err != nil {
				log.Panic(err)
			}
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}
