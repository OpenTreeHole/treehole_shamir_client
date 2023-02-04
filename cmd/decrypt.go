package cmd

import (
	"github.com/spf13/cobra"
)

var userID int

// decryptCmd represents the decrypt command
var decryptCmd = &cobra.Command{
	Use:   "decrypt",
	Short: "decrypt a pgp message",
	Long:  `get pgp message from auth server and decrypt with local private key`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	rootCmd.AddCommand(decryptCmd)

	decryptCmd.Flags().IntVarP(&userID, "user-id", "u", 0, "targeted user's user_id, optional, default all users")
}
