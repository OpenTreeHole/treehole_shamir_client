package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"os"
)

// uploadCmd represents the upload command
var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "upload public keys to auth",
	Long:  `upload public keys to auth shamir server`,
	Args:  cobra.ExactArgs(7),
	RunE: func(cmd *cobra.Command, args []string) error {
		var publicKeys []string
		for _, filename := range args {
			data, err := os.ReadFile(filename)
			if err != nil {
				return err
			}
			publicKeys = append(publicKeys, string(data))
		}

		data := map[string]any{"data": publicKeys}
		body, err := json.Marshal(data)
		if err != nil {
			return err
		}

		res, err := http.DefaultClient.Post(
			authUrl+"/api/shamir/key",
			"application/json",
			bytes.NewReader(body),
		)
		if err != nil {
			return err
		}

		body, err = io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		_ = res.Body.Close()

		if res.StatusCode != 200 {
			return fmt.Errorf("upload public key error: %s", string(body))
		}
		fmt.Printf("upload public key success")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(uploadCmd)

}
