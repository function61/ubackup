package main

import (
	"compress/gzip"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"github.com/function61/gokit/cryptoutil"
	"github.com/function61/gokit/pkencryptedstream"
	"github.com/spf13/cobra"
	"io"
	"os"
)

func decrypt(pathToPrivateKey string, ciphertextInput io.Reader, plaintextOutput io.Writer) error {
	pkeyFile, err := os.Open(pathToPrivateKey)
	if err != nil {
		return err
	}
	defer pkeyFile.Close()

	privateKey, err := cryptoutil.ParsePemPkcs1EncodedRsaPrivateKey(pkeyFile)
	if err != nil {
		return err
	}

	compressedPlaintextReader, err := pkencryptedstream.Reader(ciphertextInput, privateKey)
	if err != nil {
		return err
	}

	plaintextReader, err := gzip.NewReader(compressedPlaintextReader)
	if err != nil {
		return err
	}

	if _, err := io.Copy(plaintextOutput, plaintextReader); err != nil {
		return err
	}

	return nil
}

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
		cryptoutil.PemTypeRsaPublicKey)); err != nil {
		return err
	}

	return nil
}

func decryptEntry() *cobra.Command {
	return &cobra.Command{
		Use:   "decrypt [pathToPrivateKey]",
		Short: "Decrypts an encrypted backup file with your private key",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := decrypt(args[0], os.Stdin, os.Stdout); err != nil {
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
