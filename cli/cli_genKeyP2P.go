package cli

import (
	"encoding/hex"
	"fmt"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/opentracing/opentracing-go/log"
	"github.com/thanhxeon2470/TSS_chain/p2p"
)

func (cli *CLI) GenerateKeyPairP2P() {
	prk, _, err := p2p.GenerateKeyPairP2P()
	if err != nil {
		log.Error(err)
		return
	}
	prkMarshal, err := crypto.MarshalPrivateKey(prk)
	if err != nil {
		log.Error(err)
		return
	}

	fmt.Println(hex.EncodeToString(prkMarshal))
	h, err := libp2p.New(libp2p.Identity(prk))
	if err != nil {
		log.Error(err)
		return
	}
	fmt.Println(h.ID())
	h.Close()
	if err != nil {
		log.Error(err)
	}
	return
}
