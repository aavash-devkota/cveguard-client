package cmd

import (
	"cveguard-client/cmd/add"
	"cveguard-client/cmd/ls"
	"cveguard-client/cmd/remove"
	"cveguard-client/cmd/scan"
	"log"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cveguard-client",
	Short: "Client program for CVEGuard",
	Long: `CVEGuard-Client is a CLI client for CVEGuard that integrates with the CVEGuard web app.
This application is required to send the information about the packages installed in a project.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	initConfig()

	rootCmd.AddCommand(add.AddCmd)
	rootCmd.AddCommand(ls.LsCmd)
	rootCmd.AddCommand(remove.RemoveCmd)
	rootCmd.AddCommand(scan.ScanCmd)
}

func initConfig() {
	// Form the configuration file path
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln("[ERROR] Could not read config file:", err)
	}
	configFilePath := filepath.Join(userHomeDir, ".cveguard-client.yaml")

	// Create the config file if it does not exist
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		if err := viper.SafeWriteConfigAs(configFilePath); err != nil {
			log.Fatalln("[ERROR] Error writing config file:", err)
		}
	}

	// Setup viper config file location
	viper.AddConfigPath(userHomeDir)
	viper.SetConfigName(".cveguard-client")
	viper.SetConfigType("yaml")

	// Read configuration
	err = viper.ReadInConfig()
	if err != nil {
		log.Fatalln("[ERROR] Could not read config:", err)
	}
	viper.WatchConfig()

	// Create a Client ID if it does not exist
	if viper.Get("client_id") == nil {
		newClientId := uuid.New().String()
		viper.Set("client_id", newClientId)
		err := viper.WriteConfig()
		if err != nil {
			log.Fatalln("[ERROR] Could not write config file:", err)
		}
	}
}
