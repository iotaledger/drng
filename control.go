package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/drand/drand/core"
	"github.com/drand/drand/key"
	"github.com/drand/drand/net"
	control "github.com/drand/drand/protobuf/drand"

	json "github.com/nikkolasg/hexjson"
	"github.com/urfave/cli"
)

// shareCmd decides whether the command is for a DKG or for a resharing and
// dispatch to the respective sub-commands.
func shareCmd(c *cli.Context) error {
	if !c.Args().Present() {
		fatal("drand: needs at least one group.toml file argument")
	}
	groupPath := c.Args().First()
	groupPath, err := filepath.Abs(groupPath)
	if err != nil {
		fatal("can't open group path absolute path from %s", c.Args().First())
	}
	testEmptyGroup(groupPath)

	if c.IsSet(oldGroupFlag.Name) {
		testEmptyGroup(c.String(oldGroupFlag.Name))
		fmt.Println("drand: old group file given for resharing protocol")
		return initReshare(c, groupPath)
	}

	conf := contextToConfig(c)
	fs := key.NewFileStore(conf.ConfigFolder())
	_, errG := fs.LoadGroup()
	_, errS := fs.LoadShare()
	_, errD := fs.LoadDistPublic()
	// XXX place that logic inside core/ directly with only one method
	freshRun := errG != nil || errS != nil || errD != nil
	if freshRun {
		fmt.Println("drand: no current distributed key -> running DKG protocol.")
		err = initDKG(c, groupPath)
	} else {
		fmt.Println("drand: found distributed key -> running resharing protocol.")
		err = initReshare(c, groupPath)
	}
	return err
}

// initDKG indicates to the daemon to start the DKG protocol, as a leader or
// not. The method waits until the DKG protocol finishes or an error occured.
// If the DKG protocol finishes successfully, the beacon randomness loop starts.
func initDKG(c *cli.Context, groupPath string) error {
	// still trying to load it ourself now for the moment
	// just to test if it's a valid thing or not
	conf := contextToConfig(c)
	client, err := net.NewControlClient(conf.ControlPort())
	if err != nil {
		fatal("drand: error creating control client: %s", err)
	}

	if c.IsSet(userEntropyOnlyFlag.Name) && !c.IsSet(sourceFlag.Name) {
		fmt.Print("drand: userEntropyOnly needs to be used with the source flag, which is not specified here. userEntropyOnly flag is ignored.")
	}
	entropyInfo := entropyInfoFromReader(c)

	fmt.Print("drand: waiting the end of DKG protocol ... " +
		"(you can CTRL-C to not quit waiting)")

	_, err = client.InitDKG(groupPath, c.Bool(leaderFlag.Name), c.String(timeoutFlag.Name), entropyInfo)
	if err != nil {
		fatal("drand: initdkg %s", err)
	}
	return nil
}

// initReshare indicates to the daemon to start the resharing protocol, as a
// leader or not. The method waits until the resharing protocol finishes or
// an error occured. TInfofhe "old group" toml is inferred either from the local
// informations that the drand node is keeping (saved in filesystem), and can be
// superseeded by the command line flag "old-group".
// If the DKG protocol finishes successfully, the beacon randomness loop starts.
// NOTE: If the contacted node is not present in the new list of nodes, the
// waiting *can* be infinite in some cases. It's an issue that is low priority
// though.
func initReshare(c *cli.Context, newGroupPath string) error {
	var isLeader = c.Bool(leaderFlag.Name)
	var oldGroupPath string

	if c.IsSet(oldGroupFlag.Name) {
		oldGroupPath = c.String(oldGroupFlag.Name)
	}
	if oldGroupPath == "" {
		fmt.Print("drand: old group path not specified. Using daemon's own group if possible.")
	}

	client := controlClient(c)
	fmt.Println("drand: initiating resharing protocol. Waiting to the end ...")
	_, err := client.InitReshare(oldGroupPath, newGroupPath, isLeader, c.String(timeoutFlag.Name))
	if err != nil {
		fatal("drand: error resharing: %s", err)
	}
	return nil
}

func getShare(c *cli.Context) error {
	client := controlClient(c)
	resp, err := client.Share()
	if err != nil {
		fatal("drand: could not request the share: %s", err)
	}
	printJSON(resp)
	return nil
}

func pingpongCmd(c *cli.Context) error {
	client := controlClient(c)
	if err := client.Ping(); err != nil {
		fatal("drand: can't ping the daemon ... %s", err)
	}
	fmt.Printf("drand daemon is alive on port %s", controlPort(c))
	return nil
}

func showGroupCmd(c *cli.Context) error {
	client := controlClient(c)
	r, err := client.GroupFile()
	if err != nil {
		fatal("drand: fetching group file error: %s", err)
	}

	if c.IsSet(outFlag.Name) {
		filePath := c.String(outFlag.Name)
		err := ioutil.WriteFile(filePath, []byte(r.GroupToml), 0750)
		if err != nil {
			fatal("drand: can't write to file: %s", err)
		}
		fmt.Printf("group file written to %s", filePath)
	} else {
		fmt.Printf("\n\n%s", r.GroupToml)
	}
	return nil
}

func showCokeyCmd(c *cli.Context) error {
	client := controlClient(c)
	resp, err := client.CollectiveKey()
	if err != nil {
		fatal("drand: could not request drand.cokey: %s", err)
	}
	printJSON(resp)
	return nil
}

func showPrivateCmd(c *cli.Context) error {
	client := controlClient(c)
	resp, err := client.PrivateKey()
	if err != nil {
		fatal("drand: could not request drand.private: %s", err)
	}

	printJSON(resp)
	return nil
}

func showPublicCmd(c *cli.Context) error {
	client := controlClient(c)
	resp, err := client.PublicKey()
	if err != nil {
		fatal("drand: could not request drand.public: %s", err)
	}

	printJSON(resp)
	return nil
}

func showShareCmd(c *cli.Context) error {
	client := controlClient(c)
	resp, err := client.Share()
	if err != nil {
		fatal("drand: could not request drand.share: %s", err)
	}

	printJSON(resp)
	return nil
}

func controlPort(c *cli.Context) string {
	port := c.String("control")
	if port == "" {
		port = core.DefaultControlPort
	}
	return port
}

func controlClient(c *cli.Context) *net.ControlClient {
	port := controlPort(c)
	client, err := net.NewControlClient(port)
	if err != nil {
		fatal("drand: can't instantiate control client: %s", err)
	}
	return client
}

func printJSON(j interface{}) {
	buff, err := json.MarshalIndent(j, "", "    ")
	if err != nil {
		fatal("drand: could not JSON marshal: %s", err)
	}
	fmt.Println(string(buff))
}
func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func entropyInfoFromReader(c *cli.Context) *control.EntropyInfo {
	if c.IsSet(sourceFlag.Name) {
		_, err := os.Lstat(c.String(sourceFlag.Name))
		if err != nil {
			fatal("drand: cannot use given entropy source: %s", err)
		}
		source := c.String(sourceFlag.Name)
		ei := &control.EntropyInfo{
			Script:   source,
			UserOnly: c.Bool(userEntropyOnlyFlag.Name),
		}
		return ei
	}
	return nil
}
