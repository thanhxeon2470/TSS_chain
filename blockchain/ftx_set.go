package blockchain

import (
	// "github.com/thanhxeon2470/TSS_chain/blockchain"

	"encoding/hex"
	"log"

	"github.com/boltdb/bolt"
	"github.com/thanhxeon2470/TSS_chain/utils"
	"github.com/thanhxeon2470/TSS_chain/wallet"
)

const FtxBucket = "filealive"

// FTX is File transaction
type FTXset struct {
	Blockchain *Blockchain
}
type InfoIPFS struct {
	Author bool
	Exp    int64
}

//FindFTX to find all file this pubKeyHash cant access, but if map to true is Author else none
func (f FTXset) FindFTX(pubKeyHash []byte) map[string]InfoIPFS {
	listFTX := make(map[string]InfoIPFS)
	db := f.Blockchain.DB
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(FtxBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			ipfsList := DeserializeIPFS(v)

			for _, ipfs := range ipfsList.TXIpfsList {
				if ipfs.IsLockedWithKey(pubKeyHash) {
					if ipfs.IsOwner(pubKeyHash) {
						listFTX[ipfs.IpfsHash] = InfoIPFS{true, ipfs.Exp}
					} else {
						listFTX[ipfs.IpfsHash] = InfoIPFS{false, ipfs.Exp}
					}
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

//FindIPFS to find user allow and owner of this hash file
func (f FTXset) FindIPFS(ipfsHash string) map[string]bool {
	listUserAllow := make(map[string]bool)
	db := f.Blockchain.DB
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(FtxBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			ipfsList := DeserializeIPFS(v)

			for _, ipfs := range ipfsList.TXIpfsList {
				if ipfs.IpfsHash == ipfsHash {
					for _, userPubKeyHash := range ipfs.PubKeyHash {
						versionedPayload := append([]byte{wallet.Version}, userPubKeyHash...)
						checksum := wallet.Checksum(versionedPayload)

						fullPayload := append(versionedPayload, checksum...)
						address := utils.Base58Encode(fullPayload)
						author := false

						listUserAllow[string(address)] = author
					}
					pubKeyHash := wallet.HashPubKey(ipfs.PubKeyOwner)

					versionedPayload := append([]byte{wallet.Version}, pubKeyHash...)
					checksum := wallet.Checksum(versionedPayload)

					fullPayload := append(versionedPayload, checksum...)
					address := utils.Base58Encode(fullPayload)
					listUserAllow[string(address)] = true

				}
				// if ipfs.IsLockedWithKey(pubKeyHash) {
				// 	if ipfs.IsOwner(pubKeyHash) {
				// 		listFTX[ipfs.IpfsHash] = InfoIPFS{true, ipfs.Exp}
				// 	} else {
				// 		listFTX[ipfs.IpfsHash] = InfoIPFS{false, ipfs.Exp}
				// 	}
				// }
			}

		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return listUserAllow
}

func (f FTXset) ReindexFTX() {
	db := f.Blockchain.DB
	bucketName := []byte(FtxBucket)

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
		b := tx.Bucket([]byte(FtxBucket))

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
