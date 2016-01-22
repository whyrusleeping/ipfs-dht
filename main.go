package main

import (
	"fmt"
	"os"

	ds "github.com/ipfs/go-ipfs/Godeps/_workspace/src/github.com/ipfs/go-datastore"

	ma "github.com/ipfs/go-ipfs/Godeps/_workspace/src/github.com/jbenet/go-multiaddr"
	"github.com/ipfs/go-ipfs/metrics"
	ci "github.com/ipfs/go-ipfs/p2p/crypto"
	"github.com/ipfs/go-ipfs/p2p/host/basic"
	"github.com/ipfs/go-ipfs/p2p/net/swarm"
	"github.com/ipfs/go-ipfs/p2p/peer"
	"github.com/ipfs/go-ipfs/routing/dht"

	"golang.org/x/net/context"
)

func fail(i interface{}) {
	fmt.Println(i)
	os.Exit(1)
}

func main() {
	a, err := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/7000")
	if err != nil {
		fail(err)
	}

	priv, pub, err := ci.GenerateKeyPair(ci.RSA, 2048)
	if err != nil {
		fail(err)
	}

	local, err := peer.IDFromPublicKey(pub)
	if err != nil {
		fail(err)
	}

	fmt.Println("PEER ID: ", local.Pretty())

	ps := peer.NewPeerstore()

	ps.AddPrivKey(local, priv)
	ps.AddPubKey(local, pub)

	s, err := swarm.NewNetwork(context.Background(), []ma.Multiaddr{a}, local, ps, metrics.NewBandwidthCounter())
	if err != nil {
		fail(err)
	}

	host := basichost.New(s)
	dstore := ds.NewMapDatastore()

	idht := dht.NewDHT(context.Background(), host, dstore)
	_ = idht

	<-make(chan struct{})
}