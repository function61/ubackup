// File format for ubackup. Basically just: `ageEncrypt(gzip(plaintext))`
package backupfile

import (
	"compress/gzip"
	"io"
	"strings"

	"filippo.io/age"
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
func CreateEncryptorAndCompressor(ageRecipients string, sink io.Writer) (io.WriteCloser, error) {
	recipients, err := age.ParseRecipients(strings.NewReader(ageRecipients))
	if err != nil {
		return nil, err
	}

	encryptedWriter, err := age.Encrypt(sink, recipients...)
	if err != nil {
		return nil, err
	}

	return &encryptorAndCompressor{encryptedWriter, gzip.NewWriter(encryptedWriter)}, nil
}

func CreateDecryptorAndDecompressor(ageIdentities string, ciphertextAndCompressedInput io.Reader) (io.Reader, error) {
	identities, err := age.ParseIdentities(strings.NewReader(ageIdentities))
	if err != nil {
		return nil, err
	}

	compressedPlaintextReader, err := age.Decrypt(ciphertextAndCompressedInput, identities...)
	if err != nil {
		return nil, err
	}

	return gzip.NewReader(compressedPlaintextReader)
}
