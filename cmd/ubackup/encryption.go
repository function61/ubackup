package main

import (
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

func decryptionKeyToEncryptionKey(privKeyPemReader io.Reader, pubKeyOut io.Writer) error {
	privKeyPem, err := ioutil.ReadAll(privKeyPemReader)
	if err != nil {
		return err
	}

	privKey, err := cryptoutil.ParsePemPkcs1EncodedRsaPrivateKey(privKeyPem)
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
	decryptAndDecompress := func(pathToPrivateKey string, input io.Reader, output io.Writer) error {
		privateKeyFile, err := ioutil.ReadFile(pathToPrivateKey)
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
			exitIfError(decryptAndDecompress(args[0], os.Stdin, os.Stdout))
		},
	}
}

func decryptionKeyGenerateEntry() *cobra.Command {
	return &cobra.Command{
		Use:   "decryption-key-generate",
		Short: "Generate RSA private key for backup decryption",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			exitIfError(decryptionKeyGenerate(os.Stdout))
		},
	}
}

func decryptionKeyToEncryptionKeyEntry() *cobra.Command {
	return &cobra.Command{
		Use:   "decryption-key-to-encryption-key",
		Short: "Prints encryption key (= public key) of decryption key (= private key)",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			exitIfError(decryptionKeyToEncryptionKey(os.Stdin, os.Stdout))
		},
	}
}
