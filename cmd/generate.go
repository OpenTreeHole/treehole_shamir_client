package cmd

import (
	"fmt"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/spf13/cobra"
	"os"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "generate a pgp key by identity name",
	Long: `Generate a pgp key 

treehole_shamir_client generate your_name your_email

eg: 
treehole_shamir_client generate jingyijun jingyijun@fduhole.com`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return fmt.Errorf("error: your name or email not available")
		}

		var (
			privateFilename = "private.key"
			publicFilename  = "public.key"
		)

		if len(args) == 3 {
			privateFilename = args[2] + "-" + privateFilename
			publicFilename = args[2] + "-" + publicFilename
		}

		key, err := crypto.GenerateKey(args[0], args[1], "rsa", 4096)
		if err != nil {
			return fmt.Errorf("key generating error: %s", err)
		}

		armoredPrivateKey, err := key.Armor()
		if err != nil {
			return fmt.Errorf("generate armored private key error: %s", err.Error())
		}

		armoredPublicKey, err := key.GetArmoredPublicKey()
		if err != nil {
			return fmt.Errorf("generate armored private key error: %s", err.Error())
		}

		privateKeyFile, err := os.Create(privateFilename)
		if err != nil {
			return fmt.Errorf("create file private.key error: %s", err.Error())
		}

		_, err = privateKeyFile.WriteString(armoredPrivateKey)
		if err != nil {
			return fmt.Errorf("write private key error: %s", err.Error())
		}

		err = privateKeyFile.Close()
		if err != nil {
			return fmt.Errorf("error close private.key: %s", err.Error())
		}

		publicKeyFile, err := os.Create(publicFilename)
		if err != nil {
			return fmt.Errorf("create file public.key error: %s", err.Error())
		}

		_, err = publicKeyFile.WriteString(armoredPublicKey)
		if err != nil {
			return fmt.Errorf("write public key error: %s", err.Error())
		}

		err = publicKeyFile.Close()
		if err != nil {
			return fmt.Errorf("error close public.key: %s", err.Error())
		}

		fmt.Printf(
			`generate key success
identity name: %s
store in %s and %s`,
			key.GetEntity().PrimaryIdentity().Name,
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

	generateCmd.Flags().BoolP("output", "o", false, "name prefix of private.key and public.key")
}
