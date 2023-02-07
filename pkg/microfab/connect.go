package microfab

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path"

	"github.com/hyperledger-labs/microfab/pkg/client"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var connectCmd = &cobra.Command{
	Use:     "connect",
	Short:   "Writes out connection details for use by the Peer CLI and SDKs",
	GroupID: "mf",
	RunE: func(cmd *cobra.Command, args []string) error {
		return connect()
	},
}

func init() {
	connectCmd.PersistentFlags().BoolVarP(&force, "force", "f", false, "Force overwriting details directory")
	connectCmd.PersistentFlags().StringVar(&mspdir, "msp", "_mfcfg", "msp output directory")
}

func connect() error {

	urlStr := "http://console.127-0-0-1.nip.io:8080"
	testURL, err := url.Parse(urlStr)
	if err != nil {
		return errors.Wrapf(err, "Unable to parse URL")
	}

	rootDir := path.Clean(mspdir)

	log.Printf("Connecting to URL '%s'\n", urlStr)
	log.Printf("Identity and Configuration '%s'\n", rootDir)

	// check to see if the directory exists, and if it does emptry
	cfgExists, err := Exists(rootDir)
	if err != nil {
		return err
	}

	if cfgExists {
		empty, err := isEmpty(rootDir)
		if err != nil {
			return err
		}

		if !empty && !force {
			return errors.Errorf("Config directory '%s' is not empty, use --force to overwrite", rootDir)
		}
	} else {
		os.MkdirAll(rootDir, 0755)
	}

	mfc, err := client.New(testURL, false)
	if err != nil {
		return errors.Wrapf(err, "Unable to create client")
	}

	orgs, err := mfc.GetOrganizations()
	if err != nil {
		return errors.Wrapf(err, "Unable to get Organizations")
	}

	for _, org := range orgs {

		id, err := mfc.GetIdentity(org)
		if err != nil {
			return errors.Wrapf(err, "Unable to get Identity")
		}

		idRoot := path.Join(rootDir, org, id.ID, "msp")
		os.MkdirAll(path.Join(idRoot, "admincerts"), 0755)
		os.MkdirAll(path.Join(idRoot, "cacerts"), 0755)
		os.MkdirAll(path.Join(idRoot, "keystore"), 0755)
		os.MkdirAll(path.Join(idRoot, "signcerts"), 0755)

		os.WriteFile(path.Join(idRoot, "admincerts", "cert.pem"), id.Certificate, 0644)
		os.WriteFile(path.Join(idRoot, "signcerts", "cert.pem"), id.Certificate, 0644)
		os.WriteFile(path.Join(idRoot, "keystore", "cert_sk"), id.PrivateKey, 0644)
		os.WriteFile(path.Join(idRoot, "cacerts", "cert.pem"), id.CA, 0644)

		// get the peers, if there's no peer then move on
		peer, err := mfc.GetPeer(org)
		if err != nil {
			continue
		}

		u, err := url.Parse(peer.APIURL)
		if err != nil {
			return errors.Wrapf(err, "Unable to prase APIURL")
		}

		f, err := os.Create(path.Join(rootDir, fmt.Sprintf("%s.env", org)))
		if err != nil {
			return errors.Wrapf(err, "Unable to form path for context")
		}

		f.WriteString(fmt.Sprintf("export CORE_PEER_ADDRESS=%s\n", u.Host))
		f.WriteString(fmt.Sprintf("export CORE_PEER_LOCALMSPID=%s\n", peer.MSPID))
		f.WriteString(fmt.Sprintf("export CORE_PEER_MSPCONFIGPATH=%s\n", idRoot))
		f.Sync()

		log.Printf("For %s context run  'source %s'", org, f.Name())

	}
	return nil

}

func isEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}
