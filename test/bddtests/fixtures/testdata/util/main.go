/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

func privateKeyToPEM(privateKey *ecdsa.PrivateKey) ([]byte, error) {
	keyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	block := pem.Block{
		Type:    "PRIVATE KEY",
		Headers: nil,
		Bytes:   keyBytes,
	}

	return pem.EncodeToMemory(&block), nil
}

func publicKeyToPEM(publicKey crypto.PublicKey) ([]byte, error) {
	keyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, err
	}

	block := pem.Block{
		Type:    "PUBLIC KEY",
		Headers: nil,
		Bytes:   keyBytes,
	}

	return pem.EncodeToMemory(&block), nil
}

func generateKeyPair() ([]byte, []byte, error) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	privPEMBytes, err := privateKeyToPEM(privKey)
	if err != nil {
		return nil, nil, err
	}

	pubPEMBytes, err := publicKeyToPEM(privKey.Public())
	if err != nil {
		return nil, nil, err
	}

	return privPEMBytes, pubPEMBytes, nil
}

func generateKeyFiles(dirPath string) error {
	err := os.MkdirAll(dirPath, 0777)
	if err != nil {
		return err
	}

	privKeyFile := filepath.Join(dirPath, "private.key")
	pubKeyFile := filepath.Join(dirPath, "public.key")

	privPEMBytes, pubPEMBytes, err := generateKeyPair()
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(privKeyFile, privPEMBytes, 0600)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(pubKeyFile, pubPEMBytes, 0600)
	if err != nil {
		return err
	}

	return nil
}

// main generates a signing key pair to the directory specified in the argument
func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Expecting output directory")
		return
	}

	dir := args[0]

	if err := generateKeyFiles(dir); err != nil {
		fmt.Printf("Unable to generate key pair: %s\n", err)
	} else {
		fmt.Printf("Successfully generated key pair to directory: %s\n", dir)
	}
}
