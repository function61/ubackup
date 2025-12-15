package backupfile

import (
	"io"

	"filippo.io/age"
)

func DecryptionKeyGenerate(out io.Writer) error {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return err
	}

	if _, err := out.Write([]byte(identity.String())); err != nil {
		return err
	}

	return nil
}

func DecryptionKeyToEncryptionKey(privKeyReader io.Reader, pubKeyOut io.Writer) error {
	identities, err := age.ParseIdentities(privKeyReader)
	if err != nil {
		return err
	}

	for _, identity := range identities {
		if _, err := pubKeyOut.Write([]byte(identity.(*age.X25519Identity).Recipient().String())); err != nil {
			return err
		}
	}

	return nil
}
