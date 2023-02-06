package cmd

import (
	"fmt"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/gopenpgp/v2/helper"
	"github.com/spf13/cobra"
	"os"
)

var prefix string

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "generate a pgp key by identity name",
	Long: `Generate a pgp key 

treehole_shamir_client generate your_name your_email

eg: 
treehole_shamir_client generate jingyijun jingyijun@fduhole.com 123465789`,
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 3 {
			return fmt.Errorf("error: your name or email or password not available")
		}

		var (
			privateFilename = "private.key"
			publicFilename  = "public.key"
		)

		if prefix != "" {
			privateFilename = prefix + "-" + privateFilename
			publicFilename = prefix + "-" + publicFilename
		}

		armoredPrivateKey, err := helper.GenerateKey(args[0], args[1], []byte(args[2]), "rsa", 4096)
		if err != nil {
			return fmt.Errorf("key generating error: %s", err)
		}

		key, err := crypto.NewKeyFromArmored(armoredPrivateKey)
		if err != nil {
			return err
		}

		armoredPublicKey, err := key.GetArmoredPublicKey()
		if err != nil {
			return fmt.Errorf("generate armored private key error: %s", err.Error())
		}

		err = os.WriteFile(privateFilename, []byte(armoredPrivateKey), 0666)
		if err != nil {
			return err
		}

		err = os.WriteFile(publicFilename, []byte(armoredPublicKey), 0666)
		if err != nil {
			return err
		}

		fmt.Printf(
			`generate key success
identity name: %s
password: %s
store in %s and %s`,
			key.GetEntity().PrimaryIdentity().Name,
			args[2],
			privateFilename,
			publicFilename,
		)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// generateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// generateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	generateCmd.Flags().StringVarP(&prefix, "output", "o", "", "name prefix of private.key and public.key")
}
