package p2p

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/multiformats/go-multiaddr"
)

// DiscoveryInterval is how often we re-publish our mDNS records.
const DiscoveryInterval = time.Hour

// DiscoveryServiceTag is used in our mDNS advertisements to discover other chat peers.
const DiscoveryServiceTag = "room-TSS-mDNS"
const DiscoveryRendezvousTag = "tss"

var data2Send = make(chan []byte, 4096)
var Data2Handle = make(chan []byte, 4096)

func InitP2P(port int, bootsNodes []string, useKey bool) {
	// parse some flags to set our nickname and the room to join
	ctx := context.Background()
	prk, _, err := GenerateKeyPairP2P()
	if err != nil {
		log.Println(err)
		return
	}
	// create a new libp2p Host that listens on a random TCP port
	if useKey {
		prkStr := os.Getenv("P2P_KEY")

		if prkStr != "" {
			prkBytes, err := hex.DecodeString(prkStr)
			if err != nil {
				log.Println(err)
				return
			}
			prk, err = crypto.UnmarshalPrivateKey(prkBytes)
			if err != nil {
				log.Println(err)
				return
			}
		}
	}

	h, err := makeHost(port, prk)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Start node : ", h.ID())
	for _, peer := range bootsNodes {
		err = connectPeer(ctx, h, peer)
		if err != nil {
			log.Println(err)
		}
	}

	// if len(addrs) != 1 {
	// 	log.Println("didn't expect change in returned addresses.")
	// }
	// create a new PubSub service using the GossipSub router
	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		log.Println(err)
	}

	// setup local mDNS discovery
	if err := setupDiscovery(h); err != nil {
		log.Println(err)
	}

	// if err := DiscoveryPeers(ctx, h); err != nil {
	// 	log.Println(err)
	// }

	// use the nickname from the cli flag, or a default if blank
	nick := defaultNick(h.ID())

	// join the room P2P
	r, err := JoinRoom(ctx, ps, h.ID(), nick, "tss-1.0")
	if err != nil {
		panic(err)
	}
	go r.handleEvents()

}

func Send2Peers(d []byte) {
	data2Send <- d
}

// printErr is like fmt.Printf, but writes to stderr.
func printErr(m string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, m, args...)
}

// defaultNick generates a nickname based on the last 8 chars of a peer ID.
func defaultNick(p peer.ID) string {
	return shortID(p)
}

// shortID returns the last 8 chars of a base58-encoded peer id.
func shortID(p peer.ID) string {
	pretty := p.Pretty()
	return pretty[len(pretty)-8:]
}

// discoveryNotifee gets notified when we find a new peer via mDNS discovery
type discoveryNotifee struct {
	h host.Host
}

// HandlePeerFound connects to peers discovered via mDNS. Once they're connected,
// the PubSub system will automatically start interacting with them if they also
// support PubSub.
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	fmt.Printf("discovered new peer %s\n", pi.ID.Pretty())
	err := n.h.Connect(context.Background(), pi)
	if err != nil {
		fmt.Printf("error connecting to peer %s: %s\n", pi.ID.Pretty(), err)
	}
}

// setupDiscovery creates an mDNS discovery service and attaches it to the libp2p Host.
// This lets us automatically discover peers on the same LAN and connect to them.
func setupDiscovery(h host.Host) error {
	// setup mDNS discovery to find local peers
	s := mdns.NewMdnsService(h, DiscoveryServiceTag, &discoveryNotifee{h: h})
	return s.Start()
}

func GenerateKeyPairP2P() (crypto.PrivKey, crypto.PubKey, error) {
	// Creates a new RSA key pair for this host.
	return crypto.GenerateKeyPairWithReader(crypto.Secp256k1, 32, rand.Reader)
}

func makeHost(port int, prvKey crypto.PrivKey) (host.Host, error) {
	// 0.0.0.0 will listen on any interface device.
	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))
	// libp2p.New constructs a new libp2p Host.
	// Other options can be added here.
	return libp2p.New(
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(prvKey),
	)
}

func connectPeer(ctx context.Context, h host.Host, destination string) error {
	log.Println("This node's multiaddresses:")
	for _, la := range h.Addrs() {
		log.Printf(" - %v/p2p/%s\n", la, h.ID().Pretty())
	}
	log.Println()
	if destination == "" {
		return fmt.Errorf("No destination!\n")
	}

	// Turn the destination into a multiaddr.
	maddr, err := multiaddr.NewMultiaddr(destination)
	if err != nil {
		log.Println(err)
		return err
	}
	return DiscoveryPeers(ctx, h, []multiaddr.Multiaddr{maddr})
	// // Extract the peer ID from the multiaddr.
	// info, err := peer.AddrInfoFromP2pAddr(maddr)
	// if err != nil {
	// 	log.Println(err)
	// 	return err
	// }

	// // Add the destination's peer multiaddress in the peerstore.
	// // This will be used during connection and stream creation by libp2p.
	// h.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)
	// err = h.Connect(ctx, *info)
	// return err
}
