// Copyright 2025 The go-commons Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

const (
	sshKeyBits = 2048
	defaultSSHKeyFileName = "ssh"
)

// GenerateSSHKeyPair generates an RSA based SSH key pair and saves them to the specified path
// If sshKeyPairPath is not set, the keys are saved to a temp directory. Specify a pattern for the temp dir using tmpDirPattern
// The private key is saved under the name passed in `{sshKeyFileName}` and the public key as `{sshKeyFileName}.pub`
// Return full path of `privateKey` and `publicKey`
func GenerateSSHKeyPair(sshKeyPairPath, tmpDirPattern, sshKeyFileName string) (string, string, error) {
	if sshKeyPairPath == "" {
		tempDir, err := os.MkdirTemp("", tmpDirPattern)
		if err != nil {
			log.Fatalln("Error creating temporary directory:", err)
		}
		sshKeyPairPath = tempDir
	}
	if sshKeyFileName == "" {
		sshKeyFileName = defaultSSHKeyFileName
	}
	privateKeyPath := path.Join(sshKeyPairPath, sshKeyFileName)
	publicKeyPath := path.Join(sshKeyPairPath, strings.Join([]string{sshKeyFileName, "pub"}, "."))

	// Generate RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, sshKeyBits)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate private key: %w", err)
	}

	// Encode the private key to PEM format
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Write the private key to a file
	err = os.WriteFile(privateKeyPath, privateKeyPEM, 0600)
	if err != nil {
		return "", "", fmt.Errorf("failed to write private key to file: %w", err)
	}

	// Generate the public key in OpenSSH authorized_keys format
	sshPublicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate public key: %w", err)
	}

	// Write the public key to a file
	err = os.WriteFile(publicKeyPath, ssh.MarshalAuthorizedKey(sshPublicKey), 0644)
	if err != nil {
		return "", "", fmt.Errorf("failed to write public key to file: %w", err)
	}

	log.Infof("SSH keys saved to [%s]", sshKeyPairPath)

	return privateKeyPath, publicKeyPath, nil
}
