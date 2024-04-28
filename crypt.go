package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"sync"
)

type bufferStruct struct {
	buf *[]byte
}

var (
	secretKey = []byte("cUev3OsnWurHVWOBl9vzrh0LQwjNUbsw")
	block     cipher.Block
	blockErr  error
	pool      = sync.Pool{
		New: func() interface{} {
			b := make([]byte, aes.BlockSize+1024)
			return &bufferStruct{buf: &b}
		},
	}
)

func init() {
	var err error
	block, err = aes.NewCipher(secretKey)
	if err != nil {
		panic(err)
	}
}

func Encrypt(text string) (string, error) {
	if block == nil {
		return "", blockErr
	}

	bs := pool.Get().(*bufferStruct)
	defer pool.Put(bs)

	iv := (*bs.buf)[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	plaintext := []byte(text)
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream((*bs.buf)[aes.BlockSize:], plaintext)

	encodedStr := hex.EncodeToString((*bs.buf)[:aes.BlockSize+len(plaintext)])
	return encodedStr, nil
}

func Decrypt(encodedText string) (string, error) {
	if block == nil {
		return "", blockErr
	}

	data, err := hex.DecodeString(encodedText)
	if err != nil {
		return "", err
	}
	if len(data) < aes.BlockSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	bs := pool.Get().(*bufferStruct)
	defer pool.Put(bs)

	iv := data[:aes.BlockSize]
	ciphertext := data[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(*bs.buf, ciphertext)

	return string((*bs.buf)[:len(ciphertext)]), nil
}
