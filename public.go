package main

import (
	"github.com/drand/drand/core"
	"github.com/drand/drand/net"
	"github.com/drand/drand/protobuf/drand"
	"github.com/nikkolasg/slog"
	"github.com/urfave/cli"
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
	defaultManager := net.NewCertManager()
	if c.IsSet("tls-cert") {
		defaultManager.Add(c.String("tls-cert"))
	}

	ids := getNodes(c)
	group := getGroup(c)
	if group.PublicKey == nil {
		slog.Fatalf("drand: group file must contain the distributed public key!")
	}

	public := group.PublicKey
	client := core.NewGrpcClientFromCert(defaultManager)
	isTLS := !c.Bool("tls-disable")
	var resp *drand.PublicRandResponse
	var err error
	for _, id := range ids {
		if c.IsSet("round") {
			resp, err = client.Public(id.Addr, public, isTLS, c.Int("round"))
		} else {
			resp, err = client.LastPublic(id.Addr, public, isTLS)
		}
		if err == nil {
			slog.Infof("drand: public randomness retrieved from %s", id.Addr)
			break
		}
		slog.Printf("drand: could not get public randomness from %s: %s", id.Addr, err)
	}

	printJSON(resp)
	return nil
}

func getCokeyCmd(c *cli.Context) error {
	defaultManager := net.NewCertManager()
	if c.IsSet("tls-cert") {
		defaultManager.Add(c.String("tls-cert"))
	}
	ids := getNodes(c)
	client := core.NewGrpcClientFromCert(defaultManager)
	var dkey *drand.DistKeyResponse
	var err error
	for _, id := range ids {
		dkey, err = client.DistKey(id.Addr, !c.Bool("tls-disable"))
		if err == nil {
			break
		}
		slog.Printf("drand: error fetching distributed key from %s : %s",
			id.Addr, err)
	}
	if dkey == nil {
		slog.Fatalf("drand: can't retrieve dist. key from all nodes")
	}
	printJSON(dkey)
	return nil
}
