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

var cfg string
var mspdir string
var force bool

// Execute the microfab command
func Execute() {
	viper.AutomaticEnv()
	viper.ReadInConfig()

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("%s\n", err)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfg, "config", "", "Microfab config")
	rootCmd.PersistentFlags().StringVar(&mspdir, "msp", "_mfcfg", "msp output directory")
	rootCmd.PersistentFlags().BoolVar(&force, "force", false, "Force overwriting msp directory")

	viper.BindPFlag("MICROFAB_CONFIG", rootCmd.PersistentFlags().Lookup("config"))

	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(connectCmd)
	rootCmd.AddCommand(pingCmd)

}
