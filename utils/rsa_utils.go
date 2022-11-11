package utils

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"

	log "web-server/alog"
	"crypto/rand"
)

const (
	RSA1024_DECRYPT_BLOCK_SIZE = 128 // 1024位RSA私钥每次解密最大长度
	RSA2048_DECRYPT_BLOCK_SIZE = 256 // 2048位RSA私钥每次解密最大长度
)

func getPublicKey(key string) (*rsa.PublicKey, error) {
	blockPublic, _ := pem.Decode([]byte(key))
	pubKeyIfe, err := x509.ParsePKIXPublicKey(blockPublic.Bytes)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	pubKey, isok := pubKeyIfe.(*rsa.PublicKey)
	if !isok {
		log.Error("interface 2 *rsa.PublicKey failed")
		return nil, errors.New("interface 2 *rsa.PublicKey failed")
	}

	//k := (pubKey.N.BitLen()+7)/8 - 11

	//log.Debugf("get pubKey successful, 最大长度为:[%d]\n", k)

	return pubKey, nil
}

func getPrivateKeyPKCS1(key string) (*rsa.PrivateKey, error) {
	blockPublic, _ := pem.Decode([]byte(key))
	pubKeyIfe, err := x509.ParsePKCS1PrivateKey(blockPublic.Bytes)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return pubKeyIfe, nil
}

func getPrivateKeyPKCS8(key string) (*rsa.PrivateKey, error) {
	blockPublic, _ := pem.Decode([]byte(key))
	tmp, err := x509.ParsePKCS8PrivateKey(blockPublic.Bytes)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	pubKeyIfe := tmp.(*rsa.PrivateKey)
	return pubKeyIfe, nil
}

func encryptPublic(msgContent, publicKey string) ([]byte, error) {
	key, err := getPublicKey(publicKey)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	after, err := rsa.EncryptPKCS1v15(rand.Reader, key, []byte(msgContent))
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return after, nil
}

// 公钥加密
func RSAPublicEncrypt(content, publicKey string) (string, error) {
	tmp, err := encryptPublic(content, publicKey)
	return base64.StdEncoding.EncodeToString(tmp), err
}

func doRsaDecrypt(content string, blockSize int, key *rsa.PrivateKey) (string, error) {
	contentByte, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		log.Errorf("decryptPrivate content base64 decode error[%v] content[%s]", err, content)
		return "", err
	}

	if blockSize <= 0 {
		blockSize = RSA1024_DECRYPT_BLOCK_SIZE
	}

	text := ""
	length := len(contentByte)
	for i := 0; i < length; i += blockSize {
		end := i + blockSize
		if length < end {
			end = length
		}
		if tmp, err := rsa.DecryptPKCS1v15(rand.Reader, key, contentByte[i:end]); err != nil {
			return "", err
		} else {
			text += string(tmp)
		}
	}

	return text, err
}

// 私钥解密(PKCS1格式证书)
func RSAPrivateDecryptPKCS1(content, privateKey string, blockSize int) (string, error) {
	key, err := getPrivateKeyPKCS1(privateKey)
	if err != nil {
		log.Errorf("decryptPrivate private key[%s] is error: %v", privateKey, err)
		return "", err
	}
	return doRsaDecrypt(content, blockSize, key)
}

// 私钥解密(PKCS8格式证书)
func RSAPrivateDecryptPKCS8(content, privateKey string, blockSize int) (string, error) {
	key, err := getPrivateKeyPKCS8(privateKey)
	if err != nil {
		log.Errorf("decryptPrivate private key[%s] is error: %v", privateKey, err)
		return "", err
	}
	return doRsaDecrypt(content, blockSize, key)
}
