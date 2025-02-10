package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

func GetPrivateKey(path string) (*rsa.PrivateKey, error) {
	certPem, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(certPem)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("failed to decode PEM block containing public key")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func GetPublicKey(path string) (*rsa.PublicKey, error) {
	certFile, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read crypto key: %w", err)
	}

	block, _ := pem.Decode(certFile)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("failed to decode crypto key")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	pubKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to get public key")
	}

	return pubKey, nil
}

// Encrypt - split raw message by chunks and encrypt
func Encrypt(msg []byte, key *rsa.PublicKey) ([]byte, error) {
	if key == nil {
		return nil, errors.New("public key is nil")
	}

	hash := sha256.New()
	chunkSize := key.Size() - 2*hash.Size() - 2
	if chunkSize <= 0 {
		return nil, errors.New("invalid key size")
	}

	var encrypted []byte
	for start := 0; start < len(msg); start += chunkSize {
		end := start + chunkSize
		if end > len(msg) {
			end = len(msg)
		}
		chunk, err := rsa.EncryptOAEP(hash, rand.Reader, key, msg[start:end], nil)
		if err != nil {
			return nil, err
		}
		encrypted = append(encrypted, chunk...)
	}
	return encrypted, nil
}

// Decrypt message
func Decrypt(msg []byte, key *rsa.PrivateKey) ([]byte, error) {
	if key == nil {
		return nil, errors.New("private key is nil")
	}

	chunkSize := key.Size()
	if len(msg)%chunkSize != 0 {
		return nil, errors.New("invalid data size")
	}

	hash := sha256.New()
	var decrypted []byte
	for start := 0; start < len(msg); start += chunkSize {
		end := start + chunkSize
		chunk, err := rsa.DecryptOAEP(hash, rand.Reader, key, msg[start:end], nil)
		if err != nil {
			return nil, err
		}
		decrypted = append(decrypted, chunk...)
	}
	return decrypted, nil
}
