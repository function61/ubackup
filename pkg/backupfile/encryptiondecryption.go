package backupfile

import (
	"compress/gzip"
	"github.com/function61/gokit/cryptoutil"
	"github.com/function61/gokit/pkencryptedstream"
	"io"
)

type encryptoAndCompressor struct {
	pkencryptedStream io.WriteCloser
	gzipWriter        io.WriteCloser
}

func (f *encryptoAndCompressor) Write(buf []byte) (int, error) {
	return f.gzipWriter.Write(buf)
}

func (f *encryptoAndCompressor) Close() error {
	if err := f.gzipWriter.Close(); err != nil {
		return err
	}

	return f.pkencryptedStream.Close()
}

// you need to call .Close() on the returned WriteCloser for the gzip header and encryption
// process to finish gracefully
func CreateEncryptorAndCompressor(rsaPublicKeyPemPkcs1 io.Reader, sink io.Writer) (io.WriteCloser, error) {
	publicKey, err := cryptoutil.ParsePemPkcs1EncodedRsaPublicKey(rsaPublicKeyPemPkcs1)
	if err != nil {
		return nil, err
	}

	encryptedWriter, err := pkencryptedstream.Writer(sink, publicKey)
	if err != nil {
		return nil, err
	}

	return &encryptoAndCompressor{encryptedWriter, gzip.NewWriter(encryptedWriter)}, nil
}

func DecryptAndDecompress(rsaPrivateKeyPemPkcs1 io.Reader, ciphertextInput io.Reader, plaintextOutput io.Writer) error {
	privateKey, err := cryptoutil.ParsePemPkcs1EncodedRsaPrivateKey(rsaPrivateKeyPemPkcs1)
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
