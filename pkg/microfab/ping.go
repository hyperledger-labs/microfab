package microfab

import (
	"log"
	"net/url"

	"github.com/hyperledger-labs/microfab/pkg/client"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var pingCmd = &cobra.Command{
	Use:     "ping",
	Short:   "Pings the microfab image to see if it's running",
	GroupID: "mf",
	RunE: func(cmd *cobra.Command, args []string) error {
		return ping()
	},
}

func ping() error {

	testURL, err := url.Parse("http://console.127-0-0-1.nip.io:8080")
	if err != nil {
		return errors.Errorf("Unable to parse URL %s", testURL.String())
	}

	mfc, err := client.New(testURL, false)
	if err != nil {
		return errors.Wrapf(err, "Unable to connect create client to connect to Microfab")

	}
	err = mfc.Ping()
	if err != nil {
		return errors.Wrapf(err, "Unable to connect to runing microfab")
	}
	log.Println("Microfab ping successful")
	return nil
}
