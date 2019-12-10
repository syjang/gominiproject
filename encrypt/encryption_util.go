package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
	"io/ioutil"
)

// PKCS7 errors.
var (
	// ErrInvalidBlockSize indicates hash blocksize <= 0.
	ErrInvalidBlockSize = errors.New("invalid blocksize")

	// ErrInvalidPKCS7Data indicates bad input to PKCS7 pad or unpad.
	ErrInvalidPKCS7Data = errors.New("invalid PKCS7 data (empty or not padded)")

	// ErrInvalidPKCS7Padding indicates PKCS7 unpad fails to bad input.
	ErrInvalidPKCS7Padding = errors.New("invalid padding on input")
)

const baseKey = "alcheraencrytion"

func FileEncrypter(srcName, dstName string) error {
	input, err := ioutil.ReadFile(srcName)
	if err != nil {
		return err
	}

	key := []byte(baseKey)
	codec, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	// make stream mode
	iv := key[:aes.BlockSize]
	cs := cipher.NewCBCEncrypter(codec, iv) // ciper stream

	buf := bytes.NewBuffer(input)
	buf.Write(make([]byte, 16-(len(input)%16)))

	ctxt := make([]byte, buf.Len())
	cs.CryptBlocks(ctxt, buf.Bytes()) // the envelop is sealed
	if err := ioutil.WriteFile(dstName, ctxt, 0600); err != nil {
		return err
	}

	return nil
}

func FileDecrypter(srcName string, dstName string, srcSize int) error {
	input, err := ioutil.ReadFile(srcName)
	if err != nil {
		return err
	}

	key := []byte(baseKey)
	codec, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	// make stream mode
	iv := key[:aes.BlockSize]
	ps := cipher.NewCBCDecrypter(codec, iv) // plain stream

	ptxt := make([]byte, len(input))

	ps.CryptBlocks(ptxt, input) // unseal the envelop
	if err := ioutil.WriteFile(dstName, ptxt[:srcSize], 0600); err != nil {
		return err
	}

	return nil
}

func BufferDecrypter(key []byte, input []byte, srcSize int) ([]byte, error) {

	if len(input)%aes.BlockSize != 0 {
		err := errors.New("input not full blocks")
		return nil, err
	}

	codec, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// make stream mode
	iv := key[:aes.BlockSize]
	ps := cipher.NewCBCDecrypter(codec, iv) // plain stream

	ptxt := make([]byte, len(input))

	ps.CryptBlocks(ptxt, input) // unseal the envelop

	return ptxt[:srcSize], nil
}

func BufferDecrypterWithPKCS7(key []byte, input []byte) ([]byte, error) {

	if len(input)%aes.BlockSize != 0 {
		err := errors.New("input not full blocks")
		return nil, err
	}

	codec, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// make stream mode
	iv := key[:aes.BlockSize]
	ps := cipher.NewCBCDecrypter(codec, iv) // plain stream

	ptxt := make([]byte, len(input))

	ps.CryptBlocks(ptxt, input) // unseal the envelop

	return pkcs7Unpad(ptxt, aes.BlockSize)
}

// Reference
// https://github.com/go-web/tokenizer/blob/master/pkcs7.go

func pkcs7Unpad(b []byte, blocksize int) ([]byte, error) {
	if blocksize <= 0 {
		return nil, ErrInvalidBlockSize
	}
	if b == nil || len(b) == 0 {
		return nil, ErrInvalidPKCS7Data
	}
	if len(b)%blocksize != 0 {
		return nil, ErrInvalidPKCS7Padding
	}
	c := b[len(b)-1]
	n := int(c)
	if n == 0 || n > len(b) {
		return nil, ErrInvalidPKCS7Padding
	}
	for i := 0; i < n; i++ {
		if b[len(b)-n+i] != c {
			return nil, ErrInvalidPKCS7Padding
		}
	}
	return b[:len(b)-n], nil
}

// reference
// https://github.com/hyperledger/fabric-sdk-go/blob/master/internal/github.com/hyperledger/fabric/bccsp/sw/aes.go

func pkcs7Padding(src []byte) []byte {
	padding := aes.BlockSize - len(src)%aes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

func pkcs7UnPadding(src []byte) ([]byte, error) {
	length := len(src)
	unpadding := int(src[length-1])

	if unpadding > aes.BlockSize || unpadding == 0 {
		return nil, errors.New("Invalid pkcs7 padding (unpadding > aes.BlockSize || unpadding == 0)")
	}

	pad := src[len(src)-unpadding:]
	for i := 0; i < unpadding; i++ {
		if pad[i] != byte(unpadding) {
			return nil, errors.New("Invalid pkcs7 padding (pad[i] != unpadding)")
		}
	}

	return src[:(length - unpadding)], nil
}

func aesCBCEncrypt(key, s []byte) ([]byte, error) {
	return aesCBCEncryptWithRand(rand.Reader, key, s)
}

func aesCBCEncryptWithRand(prng io.Reader, key, s []byte) ([]byte, error) {
	if len(s)%aes.BlockSize != 0 {
		return nil, errors.New("Invalid plaintext. It must be a multiple of the block size")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(s))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(prng, iv); err != nil {
		return nil, err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], s)

	return ciphertext, nil
}

func aesCBCEncryptWithIV(IV []byte, key, s []byte) ([]byte, error) {
	if len(s)%aes.BlockSize != 0 {
		return nil, errors.New("Invalid plaintext. It must be a multiple of the block size")
	}

	if len(IV) != aes.BlockSize {
		return nil, errors.New("Invalid IV. It must have length the block size")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(s))
	copy(ciphertext[:aes.BlockSize], IV)

	mode := cipher.NewCBCEncrypter(block, IV)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], s)

	return ciphertext, nil
}

func aesCBCDecrypt(key, src []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(src) < aes.BlockSize {
		return nil, errors.New("Invalid ciphertext. It must be a multiple of the block size")
	}
	iv := src[:aes.BlockSize]
	src = src[aes.BlockSize:]

	if len(src)%aes.BlockSize != 0 {
		return nil, errors.New("Invalid ciphertext. It must be a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)

	mode.CryptBlocks(src, src)

	return src, nil
}

// AESCBCPKCS7Encrypt combines CBC encryption and PKCS7 padding
func AESCBCPKCS7Encrypt(key, src []byte) ([]byte, error) {
	// First pad
	tmp := pkcs7Padding(src)

	// Then encrypt
	return aesCBCEncrypt(key, tmp)
}

// AESCBCPKCS7Decrypt combines CBC decryption and PKCS7 unpadding
func AESCBCPKCS7Decrypt(key, src []byte) ([]byte, error) {
	// First decrypt
	pt, err := aesCBCDecrypt(key, src)
	if err == nil {
		return pkcs7UnPadding(pt)
	}
	return nil, err
}

// AESCBCPKCS7Encrypt combines CBC encryption and PKCS7 padding using as prng the passed to the function
func AESCBCPKCS7EncryptWithRand(prng io.Reader, key, src []byte) ([]byte, error) {
	// First pad
	tmp := pkcs7Padding(src)

	// Then encrypt
	return aesCBCEncryptWithRand(prng, key, tmp)
}

// AESCBCPKCS7Encrypt combines CBC encryption and PKCS7 padding, the IV used is the one passed to the function
func AESCBCPKCS7EncryptWithIV(IV []byte, key, src []byte) ([]byte, error) {
	// First pad
	tmp := pkcs7Padding(src)

	// Then encrypt
	return aesCBCEncryptWithIV(IV, key, tmp)
}
