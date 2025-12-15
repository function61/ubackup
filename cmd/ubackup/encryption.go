package main

import (
	"io"
	"os"

	"github.com/function61/gokit/os/osutil"
	"github.com/function61/ubackup/pkg/backupfile"
	"github.com/spf13/cobra"
)

func decryptEntry() *cobra.Command {
	decryptAndDecompress := func(pathToPrivateKey string, input io.Reader, output io.Writer) error {
		privateKeyFile, err := os.ReadFile(pathToPrivateKey)
		if err != nil {
			return err
		}

		plaintextDecompressed, err := backupfile.CreateDecryptorAndDecompressor(
			string(privateKeyFile),
			input)
		if err != nil {
			return err
		}

		_, err = io.Copy(output, plaintextDecompressed)
		return err
	}

	return &cobra.Command{
		Use:   "decrypt-and-decompress [pathToPrivateKey]",
		Short: "Decrypts an encrypted backup file (from stdin) with your private key",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			osutil.ExitIfError(decryptAndDecompress(args[0], os.Stdin, os.Stdout))
		},
	}
}

func decryptionKeyGenerateEntry() *cobra.Command {
	return &cobra.Command{
		Use:   "decryption-key-generate",
		Short: "Generate private key for backup decryption",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			osutil.ExitIfError(backupfile.DecryptionKeyGenerate(os.Stdout))
		},
	}
}

func decryptionKeyToEncryptionKeyEntry() *cobra.Command {
	return &cobra.Command{
		Use:   "decryption-key-to-encryption-key",
		Short: "Prints encryption key (= public key) of decryption key (= private key)",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			osutil.ExitIfError(backupfile.DecryptionKeyToEncryptionKey(os.Stdin, os.Stdout))
		},
	}
}
