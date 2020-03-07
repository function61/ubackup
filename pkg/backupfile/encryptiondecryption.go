// File format for ubackup. Basically just: pkencryptedstream(gzip(plaintext))
// pkencryptedstream is aesKeyEnvelope(rsaOaep(aesKey, pubKey)) + iv + aesCtr(plaintextGzipped)
package backupfile

import (
	"compress/gzip"
	"github.com/function61/gokit/cryptoutil"
	"github.com/function61/gokit/pkencryptedstream"
	"io"
)

type encryptorAndCompressor struct {
	pkencryptedStream io.WriteCloser
	gzipWriter        io.WriteCloser
}

func (f *encryptorAndCompressor) Write(buf []byte) (int, error) {
	return f.gzipWriter.Write(buf)
}

func (f *encryptorAndCompressor) Close() error {
	// gzipWriter does not close the underlying io.Writer
	if err := f.gzipWriter.Close(); err != nil {
		return err
	}

	// is an cipher.StreamWriter which calls close on the underlying io.Writer
	return f.pkencryptedStream.Close()
}

// you need to call .Close() on the returned WriteCloser for the gzip header and encryption
// process to finish gracefully
func CreateEncryptorAndCompressor(rsaPublicKeyPemPkcs1 string, sink io.Writer) (io.WriteCloser, error) {
	publicKey, err := cryptoutil.ParsePemPkcs1EncodedRsaPublicKey([]byte(rsaPublicKeyPemPkcs1))
	if err != nil {
		return nil, err
	}

	encryptedWriter, err := pkencryptedstream.Writer(sink, publicKey)
	if err != nil {
		return nil, err
	}

	return &encryptorAndCompressor{encryptedWriter, gzip.NewWriter(encryptedWriter)}, nil
}

func CreateDecryptorAndDecompressor(rsaPrivateKeyPemPkcs1 string, ciphertextAndCompressedInput io.Reader) (io.Reader, error) {
	privateKey, err := cryptoutil.ParsePemPkcs1EncodedRsaPrivateKey([]byte(rsaPrivateKeyPemPkcs1))
	if err != nil {
		return nil, err
	}

	compressedPlaintextReader, err := pkencryptedstream.Reader(ciphertextAndCompressedInput, privateKey)
	if err != nil {
		return nil, err
	}

	return gzip.NewReader(compressedPlaintextReader)
}
