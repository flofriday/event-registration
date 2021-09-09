package main

import (
	"crypto/rand"
	"encoding/base64"
)

/* Create a 128 bit long UUID and return it base64 encoded*/
func createUuid() (string, error) {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(bytes), nil
}
