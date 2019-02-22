package utils

import (
	"crypto/rand"
	"errors"
	"io"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/nacl/secretbox"
)

//CreateHashPassphrase Hashes users password for storage
func CreateHashPassphrase(passphrase []byte) ([]byte, error) {
	hashedPasspharse, err := bcrypt.GenerateFromPassword(passphrase, bcrypt.DefaultCost)
	if err != nil {
		return []byte{}, errors.New("[Crypto] Failed to create hash for the password provided")
	}
	return hashedPasspharse, nil
}

//VerifyPassphrase ...
func VerifyPassphrase(passphrase []byte, hashedPasspharse []byte) bool {
	err := bcrypt.CompareHashAndPassword(hashedPasspharse, passphrase)
	if err != nil {
		return false
	}
	return true
}

//Encrypt ...
func Encrypt(data []byte, passphrase []byte) []byte {
	secretKey := getSecretKeyFromPassphrase(passphrase)
	// You must use a different nonce for each message you encrypt with the
	// same key. Since the nonce here is 192 bits long, a random value
	// provides a sufficiently small probability of repeats.
	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		panic(err)
	}
	// This encrypts "hello world" and appends the result to the nonce.
	encrypted := secretbox.Seal(nonce[:], data, &nonce, &secretKey)
	return encrypted
}

//Decrypt ...
func Decrypt(encrypted []byte, passphrase []byte) []byte {
	secretKey := getSecretKeyFromPassphrase(passphrase)
	// When you decrypt, you must use the same nonce and key you used to
	// encrypt the message. One way to achieve this is to store the nonce
	// alongside the encrypted message. Above, we stored the nonce in the first
	// 24 bytes of the encrypted text.

	var decryptNonce [24]byte
	copy(decryptNonce[:], encrypted[:24])
	decrypted, ok := secretbox.Open(nil, encrypted[24:], &decryptNonce, &secretKey)
	if !ok {
		panic("decryption error")
	}
	return decrypted
}

//internal
func getSecretKeyFromPassphrase(passphrase []byte) [32]byte {
	secretKeyBytes := passphrase

	var secretKey [32]byte
	copy(secretKey[:], secretKeyBytes)
	return secretKey
}
