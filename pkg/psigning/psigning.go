package psigning

import (
	"errors"
	"fmt"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/intelsdi-x/pulse/core/perror"
	"golang.org/x/crypto/openpgp"
)

type SigningManager struct{}

var (
	ErrPluginNotFound        = errors.New("Plugin not found")
	ErrKeyringFileNotFound   = errors.New("Keyring file (.gpg) not found")
	ErrUnableToReadKeyring   = errors.New("Unable to read keyring")
	ErrSignedFileNotFound    = errors.New("Signed file not found")
	ErrSignatureFileNotFound = errors.New("Signature file (.asc) not found")
	ErrCheckSignature        = errors.New("Error checking signature")
)

//ValidateSignature is exported for plugin authoring
func (s *SigningManager) ValidateSignature(keyringFile, signedFile, signatureFile string) perror.PulseError {
	smLogger := log.WithFields(log.Fields{
		"_block":  "ValidateSignature",
		"_module": "psigning",
	})
	keyringf, err := os.Open(keyringFile)
	if err != nil {
		fields := map[string]interface{}{
			"error": err,
			"file":  keyringFile,
		}
		pe := perror.New(ErrKeyringFileNotFound, fields)
		smLogger.WithFields(fields).Error(ErrKeyringFileNotFound)
		return pe
	}
	defer keyringf.Close()

	keyring, err := openpgp.ReadKeyRing(keyringf)
	if err != nil {
		fields := map[string]interface{}{
			"error": err,
		}
		pe := perror.New(ErrUnableToReadKeyring, fields)
		smLogger.WithFields(fields).Error(ErrUnableToReadKeyring)
		return pe
	}

	signed, err := os.Open(signedFile)
	if err != nil {
		fields := map[string]interface{}{
			"error": err,
			"file":  signedFile,
		}
		pe := perror.New(ErrSignedFileNotFound, fields)
		smLogger.WithFields(fields).Error(ErrSignedFileNotFound)
		return pe
	}
	defer signed.Close()

	signature, err := os.Open(signatureFile)
	if err != nil {
		fields := map[string]interface{}{
			"error": err,
			"file":  signatureFile,
		}
		pe := perror.New(ErrSignatureFileNotFound, fields)
		smLogger.WithFields(fields).Error(ErrSignatureFileNotFound)
		return pe
	}
	defer signature.Close()

	checked, err := openpgp.CheckArmoredDetachedSignature(keyring, signed, signature)
	if err != nil {
		fields := map[string]interface{}{
			"error": err,
		}
		pe := perror.New(ErrCheckSignature, fields)
		smLogger.WithFields(fields).Error(ErrCheckSignature)
		return pe
	}

	var signedby string
	for k := range checked.Identities {
		signedby = signedby + k
	}
	fmt.Printf("Signature made %v using RSA key ID %v\nGood signature from %v\n", time.Now().Format(time.RFC1123), checked.PrimaryKey.KeyIdShortString(), signedby)
	return nil
}
