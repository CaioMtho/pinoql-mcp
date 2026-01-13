package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
)

const (
	dekSize = 32
)

type Envelope struct {
	CiphertextHex string
	WrappedDEKHex string
}

type CryptoManager struct {
	MasterKey []byte
}

func NewCryptoManager(masterKey []byte) (*CryptoManager, error) {
	if len(masterKey) != 32 {
		return nil, errors.New("master key must be 32 bytes")
	}
	return &CryptoManager{MasterKey: masterKey}, nil
}

func randomBytes(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	return b, err
}

func newGCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func (cm *CryptoManager) Encrypt(plaintext []byte) (*Envelope, error) {
	dek, err := randomBytes(dekSize)
	if err != nil {
		return nil, err
	}

	dataGCM, err := newGCM(dek)
	if err != nil {
		return nil, err
	}

	dataNonce, err := randomBytes(dataGCM.NonceSize())
	if err != nil {
		return nil, err
	}

	ciphertext := dataGCM.Seal(
		dataNonce,
		dataNonce,
		plaintext,
		nil,
	)

	kekGCM, err := newGCM(cm.MasterKey)
	if err != nil {
		return nil, err
	}

	kekNonce, err := randomBytes(kekGCM.NonceSize())
	if err != nil {
		return nil, err
	}

	wrappedDEK := kekGCM.Seal(
		kekNonce,
		kekNonce,
		dek,
		nil,
	)

	return &Envelope{
		CiphertextHex: hex.EncodeToString(ciphertext),
		WrappedDEKHex: hex.EncodeToString(wrappedDEK),
	}, nil
}

func (cm *CryptoManager) Decrypt(env *Envelope) ([]byte, error) {
	ciphertext, err := hex.DecodeString(env.CiphertextHex)
	if err != nil {
		return nil, err
	}

	wrappedDEK, err := hex.DecodeString(env.WrappedDEKHex)
	if err != nil {
		return nil, err
	}

	kekGCM, err := newGCM(cm.MasterKey)
	if err != nil {
		return nil, err
	}

	if len(wrappedDEK) < kekGCM.NonceSize() {
		return nil, errors.New("invalid wrapped DEK")
	}

	kekNonce := wrappedDEK[:kekGCM.NonceSize()]
	encDEK := wrappedDEK[kekGCM.NonceSize():]

	dek, err := kekGCM.Open(nil, kekNonce, encDEK, nil)
	if err != nil {
		return nil, err
	}

	dataGCM, err := newGCM(dek)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < dataGCM.NonceSize() {
		return nil, errors.New("invalid ciphertext")
	}

	dataNonce := ciphertext[:dataGCM.NonceSize()]
	encData := ciphertext[dataGCM.NonceSize():]

	plaintext, err := dataGCM.Open(nil, dataNonce, encData, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
