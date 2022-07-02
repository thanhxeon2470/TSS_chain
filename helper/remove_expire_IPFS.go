package helper

import (
	"encoding/hex"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/thanhxeon2470/TSS_chain/blockchain"
)

var RemoveAt = time.Now().Unix()

// Remove IPFS is expired || This func run each 8 hours
func RemoveExpireIPFS(bc *blockchain.Blockchain) []string {
	var listRM []string
	if time.Now().Unix()-RemoveAt > 28800 {
		RemoveAt = time.Now().Unix()

		db := bc.DB
		err := db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(blockchain.FtxBucket))
			c := b.Cursor()

			for k, v := c.First(); k != nil; k, v = c.Next() {
				ipfsList := blockchain.DeserializeIPFS(v)

				for _, ipfs := range ipfsList.TXIpfsList {
					if ipfs.Exp < time.Now().Unix() {
						IpfsHashENC := hex.EncodeToString(ipfs.IpfsHashENC)
						getCMD := exec.Command("ipfs-cluster-ctl", "pin", "rm", IpfsHashENC)
						stdout, err := getCMD.Output()
						if err != nil {
							log.Panic("Error remove file")
							break
						}
						str := string(stdout)
						if !strings.Contains(str, IpfsHashENC) {
							fmt.Print("Cant remove file from ipfs")
						} else {
							listRM = append(listRM, IpfsHashENC)
						}
					}
				}

			}

			return nil
		})
		if err != nil {
			log.Panic(err)
		}
	}

	return listRM
}
