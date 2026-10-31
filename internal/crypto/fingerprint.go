package crypto

import (
	"crypto/sha256"
	"encoding/base32"
)

func Fingerprint(pub []byte) string {
	h := sha256.Sum256(pub)
	return "SHA256:" + base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(h[:])
}
