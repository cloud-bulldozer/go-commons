package ssh

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/crypto/ssh"
)

var _ = Describe("ssh.go verify key creation", func() {
	Context("Create SSH pair and validate", func() {
		It("create valid SSH keys", func() {
			private, public, err := GenerateSSHKeyPair("", "", "")
			Expect(err).To(BeNil())
			err = validateSSHKeyPair(private, public)
			Expect(err).To(BeNil())
		})
	})

	Context("Validate SSH key name and location", func() {
		It("create with file name", func(){
			fileName := "foo"
			private, public, err := GenerateSSHKeyPair("", "", fileName)
			Expect(err).To(BeNil())
			Expect(filepath.Base(private)).To(Equal(fileName))
			Expect(filepath.Base(public)).To(Equal(fmt.Sprintf("%s.pub", fileName)))
		})

		It("create with path", func ()  {
			keyPairPath := "/tmp/cloud-bulldozer-unittest"
			os.RemoveAll(keyPairPath)
			err := os.Mkdir(keyPairPath, 0755)
			Expect(err).To(BeNil())
			private, public, err := GenerateSSHKeyPair(keyPairPath, "", "")
			Expect(err).To(BeNil())
			Expect(filepath.Dir(private)).To(Equal(keyPairPath))
			Expect(filepath.Dir(public)).To(Equal(keyPairPath))

		})

		It("create with template", func ()  {
			dirPattern := "cloud-bulldozer-unittest"
			private, public, err := GenerateSSHKeyPair("", fmt.Sprintf("%s-*", dirPattern), "")
			Expect(err).To(BeNil())
			Expect(extractFirstPartOfLastDir(private)).To(Equal(dirPattern))
			Expect(extractFirstPartOfLastDir(public)).To(Equal(dirPattern))
		})
	})
})

// validateSSHKeyPair checks if the private and public key files exist, are valid, and match.
func validateSSHKeyPair(privateKeyPath, publicKeyPath string) error {
	// Read private key
	privateKeyData, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read private key file: %v", err)
	}

	// Parse private key
	privateKey, err := parsePrivateKey(privateKeyData)
	if err != nil {
		return fmt.Errorf("invalid private key: %v", err)
	}

	// Read public key
	publicKeyData, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read public key file: %v", err)
	}

	// Parse public key
	publicKey, _, _, _, err := ssh.ParseAuthorizedKey(publicKeyData)
	if err != nil {
		return fmt.Errorf("invalid public key format: %v", err)
	}

	// Extract public key from private key
	derivedPublicKey, err := extractPublicKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to extract public key: %v", err)
	}

	// Compare derived public key with provided public key
	if !bytes.Equal(derivedPublicKey.Marshal(), publicKey.Marshal()) {
		return fmt.Errorf("public key does not match private key")
	}

	return nil
}

// parsePrivateKey attempts to parse the given PEM-encoded private key data.
func parsePrivateKey(privateKeyData []byte) (interface{}, error) {
	block, _ := pem.Decode(privateKeyData)
	if block == nil {
		return nil, fmt.Errorf("no valid PEM data found")
	}

	// Try parsing as different key types
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		switch key := key.(type) {
		case *rsa.PrivateKey, *ecdsa.PrivateKey, ed25519.PrivateKey:
			return key, nil
		default:
			return nil, fmt.Errorf("unsupported private key type")
		}
	}
	if key, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	return nil, fmt.Errorf("failed to parse private key")
}

func extractPublicKey(privateKey interface{}) (ssh.PublicKey, error) {
	var pub interface{}

	switch key := privateKey.(type) {
	case *rsa.PrivateKey:
		pub = &key.PublicKey
	case *ecdsa.PrivateKey:
		pub = &key.PublicKey
	case ed25519.PrivateKey:
		pub = key.Public()
	default:
		return nil, fmt.Errorf("unsupported private key type")
	}

	return ssh.NewPublicKey(pub)
}

// Extracts the first part of the last directory name split by the last "-"
func extractFirstPartOfLastDir(path string) (string, error) {
	// Get the absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %v", err)
	}

	// Get the directory part of the path
	dirPath := filepath.Dir(absPath)

	// Get the last directory from the path
	lastDir := filepath.Base(dirPath)

	// Find the last occurrence of "-"
	lastDashIndex := strings.LastIndex(lastDir, "-")
	if lastDashIndex == -1 {
		return lastDir, nil // No "-" found, return full last directory name
	}

	// Extract the first part before the last "-"
	firstPart := lastDir[:lastDashIndex]
	return firstPart, nil
}
