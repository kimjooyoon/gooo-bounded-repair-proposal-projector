package projector

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"regexp"
)

var digestPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
var commitPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

func DigestBytes(value []byte) string {
	digest := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func DigestJSON(value any) (string, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return DigestBytes(raw), nil
}

func validDigest(value string) bool {
	return digestPattern.MatchString(value)
}

func validCommit(value string) bool {
	return commitPattern.MatchString(value)
}

func validateDigest(value, code string) error {
	if !validDigest(value) {
		return errors.New(code)
	}
	return nil
}
