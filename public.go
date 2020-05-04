package main

import (
	"errors"
	gonet "net"

	"github.com/drand/drand/core"
	"github.com/drand/drand/net"
	"github.com/drand/drand/protobuf/drand"
	"github.com/nikkolasg/slog"
	"github.com/urfave/cli/v2"
)

func getPrivateCmd(c *cli.Context) error {
	if !c.Args().Present() {
		slog.Fatal("Get private takes a group file as argument.")
	}
	defaultManager := net.NewCertManager()
	if c.IsSet("tls-cert") {
		defaultManager.Add(c.String("tls-cert"))
	}
	ids := getNodes(c)
	client := core.NewGrpcClientFromCert(defaultManager)
	var resp []byte
	var err error
	for _, public := range ids {
		resp, err = client.Private(public)
		if err == nil {
			slog.Infof("drand: successfully retrieved private randomness "+
				"from %s", public.Addr)
			break
		}
		slog.Infof("drand: error contacting node %s: %s", public.Addr, err)
	}
	if resp == nil {
		slog.Fatalf("drand: zero successful contacts with nodes")
	}

	type private struct {
		Randomness []byte
	}

	printJSON(&private{resp})
	return nil
}

func getPublicRandomness(c *cli.Context) error {
	if !c.Args().Present() {
		slog.Fatal("Get public command takes a group file as argument.")
	}
	client := core.NewGrpcClient()
	if c.IsSet(tlsCertFlag.Name) {
		defaultManager := net.NewCertManager()
		defaultManager.Add(c.String(tlsCertFlag.Name))
		client = core.NewGrpcClientFromCert(defaultManager)
	}

	ids := getNodes(c)
	group := getGroup(c)
	if group.PublicKey == nil {
		slog.Fatalf("drand: group file must contain the distributed public key!")
	}

	public := group.PublicKey
	var resp *drand.PublicRandResponse
	var err error
	var foundCorrect bool
	for _, id := range ids {
		if c.IsSet(roundFlag.Name) {
			resp, err = client.Public(id.Addr, public, id.TLS, c.Int(roundFlag.Name))
		} else {
			resp, err = client.LastPublic(id.Addr, public, id.TLS)
		}
		if err == nil {
			foundCorrect = true
			slog.Infof("drand: public randomness retrieved from %s", id.Addr)
			break
		}
		slog.Printf("drand: could not get public randomness from %s: %s", id.Addr, err)
	}
	if !foundCorrect {
		return errors.New("drand: could not verify randomness")
	}

	printJSON(resp)
	return nil
}

func getCokeyCmd(c *cli.Context) error {
	var client = core.NewGrpcClient()
	if c.IsSet(tlsCertFlag.Name) {
		defaultManager := net.NewCertManager()
		certPath := c.String(tlsCertFlag.Name)
		defaultManager.Add(certPath)
		client = core.NewGrpcClientFromCert(defaultManager)
	}
	var dkey *drand.DistKeyResponse
	for _, addr := range c.Args().Slice() {
		_, _, err := gonet.SplitHostPort(addr)
		if err != nil {
			fatal("invalid address given: %s", err)
		}
		dkey, err = client.DistKey(addr, !c.Bool("tls-disable"))
		if err == nil {
			break
		}
		slog.Printf("drand: error fetching distributed key from %s : %s",
			addr, err)
	}
	if dkey == nil {
		slog.Fatalf("drand: can't retrieve dist. key from all nodes")
	}
	printJSON(dkey)
	return nil
}
