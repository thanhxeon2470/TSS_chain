package p2p

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/multiformats/go-multiaddr"
)

func NewDHT(ctx context.Context, host host.Host, bootstrapPeers []multiaddr.Multiaddr) (*dht.IpfsDHT, error) {
	var options []dht.Option

	// if no bootstrap peers give this peer act as a bootstraping node
	// other peers can use this peers ipfs address for peer discovery via dht
	if len(bootstrapPeers) == 0 {
		options = append(options, dht.Mode(dht.ModeServer))
	}

	kdht, err := dht.New(ctx, host, options...)
	if err != nil {
		return nil, err
	}

	if err = kdht.Bootstrap(ctx); err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	for _, peerAddr := range bootstrapPeers {
		peerinfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)

		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := host.Connect(ctx, *peerinfo); err != nil {
				log.Printf("Error while connecting to node %q: %-v", peerinfo, err)
			} else {
				log.Printf("Connection established with bootstrap node: %q", *peerinfo)
			}
		}()
	}
	wg.Wait()

	return kdht, nil
}

func Discover(ctx context.Context, h host.Host, dht *dht.IpfsDHT, rendezvous string) {
	var routingDiscovery = routing.NewRoutingDiscovery(dht)

	routingDiscovery.Advertise(ctx, rendezvous)

	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			peers, err := routingDiscovery.FindPeers(ctx, rendezvous)
			if err != nil {
				log.Fatal(err)
			}
			p := <-peers
			// fmt.Println(p)
			if p.ID == h.ID() {
				continue
			}
			if h.Network().Connectedness(p.ID) != network.Connected {
				_, err := h.Network().DialPeer(ctx, p.ID)
				fmt.Printf("Connected to peer %s\n", p.ID.Pretty())
				if err != nil {
					continue
				}
			}
		}
	}
}

func DiscoveryPeers(ctx context.Context, host host.Host, discoveryPeers []multiaddr.Multiaddr) error {
	// // Find multiaddress
	// discoveryPeers := []multiaddr.Multiaddr{}
	// for _, p := range host.Network().Peers() {
	// 	discoveryPeers = append(discoveryPeers, host.Network().Peerstore().Addrs(p)...)
	// }
	// discoveryPeers = host.Addrs()

	// setup DHT with discovery server
	// this peer could run behind the nat(with private ip address)
	dht, err := NewDHT(ctx, host, discoveryPeers)
	if err != nil {
		return err
	}

	// setup peer discovery
	go Discover(ctx, host, dht, DiscoveryRendezvousTag)
	return nil
}
