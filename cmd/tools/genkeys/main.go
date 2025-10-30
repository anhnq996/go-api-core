package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	keysDir := filepath.Join("keys")
	privPath := filepath.Join(keysDir, "private.pem")
	pubPath := filepath.Join(keysDir, "public.pem")

	if err := os.MkdirAll(keysDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir keys: %v\n", err)
		os.Exit(1)
	}

	// Generate RSA private key
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate key: %v\n", err)
		os.Exit(1)
	}

	// Write private key (PKCS8) PEM
	privDer, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal private key: %v\n", err)
		os.Exit(1)
	}
	privFile, err := os.Create(privPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create private key: %v\n", err)
		os.Exit(1)
	}
	defer privFile.Close()
	if err := pem.Encode(privFile, &pem.Block{Type: "PRIVATE KEY", Bytes: privDer}); err != nil {
		fmt.Fprintf(os.Stderr, "write private key: %v\n", err)
		os.Exit(1)
	}

	// Write public key (PKIX) PEM
	pubDer, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal public key: %v\n", err)
		os.Exit(1)
	}
	pubFile, err := os.Create(pubPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create public key: %v\n", err)
		os.Exit(1)
	}
	defer pubFile.Close()
	if err := pem.Encode(pubFile, &pem.Block{Type: "PUBLIC KEY", Bytes: pubDer}); err != nil {
		fmt.Fprintf(os.Stderr, "write public key: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Generated:")
	fmt.Println(" -", privPath)
	fmt.Println(" -", pubPath)
}
