package encryptionUtils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/mergermarket/go-pkcs7"
	"github.com/pkg/errors"
	"github.com/thanhpk/randstr"
	"golang.org/x/crypto/hkdf"
	"io"
	"strings"
)

func Encrypt(plainText []byte) (*bytes.Buffer,string,error){
	ikmEnc := randstr.String(16)
	infoEnc := randstr.String(16)
	authenticationTagEnc := randstr.String(16)
	hkdfAuthenticationKey := getHkdfKey([]byte(ikmEnc),[]byte(infoEnc),16)
	plaintextEnc, err := pkcs7.Pad(plainText, aes.BlockSize)
	if err != nil {
		return nil, "", errors.New("plaintext could not be padded")
	}
	if len(plaintextEnc)%aes.BlockSize != 0 {
		return nil, "", errors.New("plaintext is not a multiple of the block size")
	}
	block, err := aes.NewCipher(hkdfAuthenticationKey)
	if err != nil {
		return nil, "", errors.New("error generating a block")
	}
	iv := make([]byte, 16)
	_,_= rand.Read(iv)

	ciphertext := make([]byte, len(plaintextEnc))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext,plaintextEnc)

	// mac sign
	hkdfAuthenticationKeyNew := getHkdfKey([]byte(ikmEnc),[]byte(authenticationTagEnc),32)
	h := hmac.New(sha256.New, hkdfAuthenticationKeyNew)
	h.Write(ciphertext)
	finalMac := h.Sum(nil)
	length := 1 + len(iv) + 1 + len(finalMac) + len(ciphertext)
	finalBuffer := make([]byte, length)
	finalBuffer[0] = byte(len(iv))
	copy(finalBuffer[1:17], iv)
	finalBuffer[17] =  byte(len(finalMac))
	copy(finalBuffer[18:50], finalMac)
	copy(finalBuffer[50:], ciphertext)
	return bytes.NewBuffer([]byte(fmt.Sprintf(`{"data":"%v"}`, base64.StdEncoding.EncodeToString(finalBuffer)))), fmt.Sprintf("%s|%s|%s|",ikmEnc,infoEnc,authenticationTagEnc), nil
}

func Decrypt(key, message string) (string, error) {
	keyArray := strings.Split(key, "|")
	if len(keyArray) < 3{
		return "", errors.New("Invalid key")
	}
	if len(keyArray[0]) < 16 || len(keyArray[1]) < 16 ||len(keyArray[2]) < 16 {
		fmt.Println("Invalid X-MessageID")
		return "", errors.New("Invalid key")
	}
	// decode the base64 string from response body to bytes
	data,err := base64.StdEncoding.DecodeString(message)
	if err!= nil{
		return "", errors.New("Error decoding message.")
	}
	//extract needed info
	initialKeyMaterial := keyArray[0]
	information := keyArray[1]
	authenticationTag := keyArray[2]

	initializationVector := data[1:17]
	macAuth := data[18:50]
	cipherText := data[50:]

	//authenticateMac
	if ! macAuthenticate(initialKeyMaterial, authenticationTag, cipherText,macAuth){
		return "", errors.New("Error validating the mac")
	}

	hkdfAuthenticationKey := getHkdfKey([]byte(initialKeyMaterial), []byte(information), 16)
	block, err := aes.NewCipher(hkdfAuthenticationKey)
	if err != nil {
		return "", errors.New("Error generating cipher")
	}
	if len(cipherText) < aes.BlockSize {
		return "", errors.New("cipher text too short")
	}
	if len(cipherText)%aes.BlockSize != 0 {
		return "", errors.New("ciphertext is not a multiple of the block size")
	}
	mode := cipher.NewCBCDecrypter(block, initializationVector)
	mode.CryptBlocks(cipherText, cipherText)
	return string(cipherText), nil
}

func macAuthenticate(ikm string, authenticationTag string, cipher []byte,mac []byte) bool {
	hkdfAuthenticationKey:= getHkdfKey([]byte(ikm), []byte(authenticationTag), 32)
	h := hmac.New(sha256.New, hkdfAuthenticationKey)
	h.Write(cipher)
	computedMac := h.Sum(nil)
	return hmac.Equal(computedMac, mac)
}

func getHkdfKey(initialKeyMaterial, info []byte, size int) []byte{
	hkdfHash := hkdf.New(sha256.New, initialKeyMaterial, nil, info)
	hkdfKey := make([]byte, size)
	if _, err := io.ReadFull(hkdfHash, hkdfKey); err != nil {
		fmt.Println(err)
	}
	return hkdfKey
}

