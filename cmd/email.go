package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"treehole_shamir_client/utils"

	"github.com/spf13/cobra"
)

var shareFile string

// emailCmd represents the email command
var emailCmd = &cobra.Command{
	Use:   "email",
	Short: "get user email from shares",
	Long: `get user email from shares in a specific file

The file should in this format
[
	"123123123123\n345345345345",
	"123123123123\n456456456456"
]`,
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := os.ReadFile(shareFile)
		if err != nil {
			return err
		}

		var shares utils.Shares
		err = json.Unmarshal(data, &shares)
		if err != nil {
			return err
		}

		fmt.Println(utils.Decrypt(shares))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(emailCmd)

	emailCmd.Flags().StringVarP(&shareFile, "file", "f", "shares.json", "set shares file")
}
