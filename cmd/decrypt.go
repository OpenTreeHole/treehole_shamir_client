package cmd

import (
	"errors"
	"fmt"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/spf13/cobra"
	"os"
	"treehole_shamir_client/utils"
)

var userID int
var privateFilename string
var Password string
var ShareFileName []string

// decryptCmd represents the decrypt command
var decryptCmd = &cobra.Command{
	Use:   "decrypt",
	Short: "decrypt a pgp message",
	Long:  `get pgp message from auth server and decrypt with local private key`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if ShareFileName != nil {
			for _, filename := range ShareFileName {
				data, err := os.ReadFile(filename)
				if err != nil {
					return err
				}

				err = utils.UploadShares(data, authUrl)
				if err != nil {
					return err
				}
			}

			return nil
		}

		data, err := os.ReadFile(privateFilename)
		if err != nil {
			return err
		}

		key, err := crypto.NewKeyFromArmored(string(data))
		if err != nil {
			return err
		}

		if !key.IsPrivate() {
			return errors.New("not private key, please check your private key file")
		}

		if Password != "" {
			utils.Key, err = key.Unlock([]byte(Password))
			if err != nil {
				return fmt.Errorf("密码错误: %s", err)
			}
		} else {
			utils.Key = key
		}

		utils.KeyRing, err = crypto.NewKeyRing(utils.Key)
		if err != nil {
			return err
		}

		if userID < 0 {
			return fmt.Errorf("invalid user_id %v", userID)
		}
		if userID == 0 {
			err := utils.DecryptAllUser(authUrl)
			if err != nil {
				return err
			}
		} else {
			err := utils.DecryptByUserID(userID, authUrl)
			if err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(decryptCmd)

	decryptCmd.Flags().IntVarP(
		&userID, "user-id", "u",
		0, "targeted user's user_id, optional")

	decryptCmd.Flags().StringVarP(
		&privateFilename, "key", "k",
		"private.key", "specific private key filename")

	decryptCmd.Flags().StringVarP(
		&Password, "pass", "p",
		"", "your private key password",
	)

	decryptCmd.Flags().StringSliceVarP(
		&ShareFileName, "file", "f",
		nil, "specific share data filename",
	)
}
