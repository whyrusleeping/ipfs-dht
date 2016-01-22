package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	ds "github.com/ipfs/go-ipfs/Godeps/_workspace/src/github.com/ipfs/go-datastore"

	ma "github.com/ipfs/go-ipfs/Godeps/_workspace/src/github.com/jbenet/go-multiaddr"
	"github.com/ipfs/go-ipfs/metrics"
	ci "github.com/ipfs/go-ipfs/p2p/crypto"
	"github.com/ipfs/go-ipfs/p2p/host/basic"
	"github.com/ipfs/go-ipfs/p2p/net/swarm"
	"github.com/ipfs/go-ipfs/p2p/peer"
	"github.com/ipfs/go-ipfs/routing/dht"
	"github.com/ipfs/go-ipfs/util/ipfsaddr"

	"golang.org/x/net/context"
)

func fail(i interface{}) {
	fmt.Println(i)
	os.Exit(1)
}

func main() {
	key := flag.String("keyfile", "", "specify file containing ipfs private key")
	bootstrap := flag.String("bootstrap", "", "specify file containing bootstrap nodes")
	listen := flag.String("listen", "/ip4/0.0.0.0/tcp/7000", "multiaddr to listen on")
	flag.Parse()

	var priv ci.PrivKey
	var pub ci.PubKey
	var local peer.ID

	if *key != "" {
		data, err := ioutil.ReadFile(*key)
		if err != nil {
			fail(err)
		}

		fpriv, err := ci.UnmarshalPrivateKey(data)
		if err != nil {
			fail(err)
		}

		fpub := priv.GetPublic()
		fid, err := peer.IDFromPublicKey(fpub)
		if err != nil {
			fail(err)
		}

		priv = fpriv
		pub = fpub
		local = fid
	} else {
		var err error
		priv, pub, err = ci.GenerateKeyPair(ci.RSA, 2048)
		if err != nil {
			fail(err)
		}

		local, err = peer.IDFromPublicKey(pub)
		if err != nil {
			fail(err)
		}
	}

	a, err := ma.NewMultiaddr(*listen)
	if err != nil {
		fail(err)
	}

	fmt.Println("Using Peer ID: ", local.Pretty())
	fmt.Println("listening on ", a)

	ps := peer.NewPeerstore()
	ps.AddPrivKey(local, priv)
	ps.AddPubKey(local, pub)

	var bsaddrs []ma.Multiaddr
	if *bootstrap != "" {
		content, err := ioutil.ReadFile(*bootstrap)
		if err != nil {
			fail(err)
		}

		for _, pb := range bytes.Split(content, []byte("\n")) {
			pbs := string(pb)
			if !strings.Contains(pbs, "/") {
				continue
			}
			baddr, err := ma.NewMultiaddr(pbs)
			if err != nil {
				fail(err)
			}

			bsaddrs = append(bsaddrs, baddr)
		}
	}

	if len(bsaddrs) > 0 {
		fmt.Println("Bootstrapping to:")
		for _, b := range bsaddrs {
			fmt.Printf("  - %s\n", b)
		}
		fmt.Println()
	}

	s, err := swarm.NewNetwork(context.Background(), []ma.Multiaddr{a}, local, ps, metrics.NewBandwidthCounter())
	if err != nil {
		fail(err)
	}
	host := basichost.New(s)
	dstore := ds.NewMapDatastore()

	for _, bsaddr := range bsaddrs {
		iaddr, err := ipfsaddr.ParseMultiaddr(bsaddr)
		if err != nil {
			fmt.Println("error parsing bootstrap: ", err)
			continue
		}

		ps.AddAddr(iaddr.ID(), iaddr.Transport(), peer.PermanentAddrTTL)

		ctx, _ := context.WithTimeout(context.Background(), time.Second*10)

		err = host.Connect(ctx, ps.PeerInfo(iaddr.ID()))
		if err != nil {
			fmt.Println("error connecting to peer: %s", err)
			continue
		}
		fmt.Printf("dial to %s succeeded!\n", iaddr.ID())
	}

	idht := dht.NewDHT(context.Background(), host, dstore)
	_ = idht

	<-make(chan struct{})
}
