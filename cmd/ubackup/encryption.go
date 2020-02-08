package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"github.com/function61/gokit/cryptoutil"
	"github.com/function61/ubackup/pkg/backupfile"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"os"
)

func decryptionKeyGenerate(out io.Writer) error {
	// using 4096 to be super safe, though 2048 seems to be what's currently used
	privKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	privKeyBytes := cryptoutil.MarshalPemBytes(
		x509.MarshalPKCS1PrivateKey(privKey),
		cryptoutil.PemTypeRsaPrivateKey)

	if _, err := out.Write(privKeyBytes); err != nil {
		return err
	}

	return nil
}

func decryptionKeyToEncryptionKey(privKeyIn io.Reader, pubKeyOut io.Writer) error {
	privKey, err := cryptoutil.ParsePemPkcs1EncodedRsaPrivateKey(privKeyIn)
	if err != nil {
		return err
	}

	if _, err := pubKeyOut.Write(cryptoutil.MarshalPemBytes(
		x509.MarshalPKCS1PublicKey(&privKey.PublicKey),
		cryptoutil.PemTypeRsaPublicKey),
	); err != nil {
		return err
	}

	return nil
}

func decryptEntry() *cobra.Command {
	decryptAndDecompress := func(pathToPrivateKey string) error {
		privateKeyFile, err := ioutil.ReadFile(pathToPrivateKey)
		if err != nil {
			return err
		}

		if err := backupfile.DecryptAndDecompress(
			bytes.NewBuffer(privateKeyFile),
			os.Stdin,
			os.Stdout,
		); err != nil {
			return err
		}

		return nil
	}

	return &cobra.Command{
		Use:   "decrypt-and-decompress [pathToPrivateKey]",
		Short: "Decrypts an encrypted backup file with your private key",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := decryptAndDecompress(args[0]); err != nil {
				panic(err)
			}
		},
	}
}

func decryptionKeyGenerateEntry() *cobra.Command {
	return &cobra.Command{
		Use:   "decryption-key-generate",
		Short: "Generate RSA private key for backup decryption",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if err := decryptionKeyGenerate(os.Stdout); err != nil {
				panic(err)
			}
		},
	}
}

func decryptionKeyToEncryptionKeyEntry() *cobra.Command {
	return &cobra.Command{
		Use:   "decryption-key-to-encryption-key",
		Short: "Prints encryption key of decryption key",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if err := decryptionKeyToEncryptionKey(os.Stdin, os.Stdout); err != nil {
				panic(err)
			}
		},
	}
}
