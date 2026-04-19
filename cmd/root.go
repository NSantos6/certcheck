package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "certcheck",
	Short: "Monitor SSL certificates and domain registration expiry",
	Long:  "certcheck helps you track SSL certificate and domain registration expiry dates — a free alternative to paid monitoring services.",
}

func Execute() error {
	return rootCmd.Execute()
}
