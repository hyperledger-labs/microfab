package microfab

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:           "microfab",
	Short:         "microfab...",
	Long:          "Microfab Launch Control",
	Version:       "c0",
	SilenceUsage:  true,
	SilenceErrors: true,
}

var defaultCfg = `{"endorsing_organizations":[{"name":"org1"}],"channels":[{"name":"mychannel","endorsing_organizations":["org1"]},{"name":"appchannel","endorsing_organizations":["org1"]}],"capability_level":"V2_5"}`

var cfg string
var mspdir string
var force bool
var cfgFile string

// Execute the microfab command
func Execute() {
	viper.AutomaticEnv()
	viper.ReadInConfig()

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("%s\n", err)
	}
}

func init() {

	rootCmd.AddGroup(&cobra.Group{ID: "mf", Title: "microfab"})
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(connectCmd)
	rootCmd.AddCommand(pingCmd)

}
