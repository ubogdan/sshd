package util

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

//AesEncrypt aes 加密
//orig 原始字符串
//key 对称密钥 密钥长度必须 16/24/32 长度
//返回值 加密之后的b64-url-encoding 字符串

//AesEncryptBytes 加密bytes
func AesEncrypt(origData []byte, key string) (string, error) {
	// 转成字节数组
	k := []byte(key)

	// 分组秘钥
	block, err := aes.NewCipher(k)
	if err != nil {
		return "", fmt.Errorf("key 长度必须 16/24/32长度: %s", err)
	}
	// 获取秘钥块的长度
	blockSize := block.BlockSize()
	// 补全码
	origData = pkcs7Padding(origData, blockSize)
	// 加密模式
	blockMode := cipher.NewCBCEncrypter(block, k[:blockSize])
	// 创建数组
	cryted := make([]byte, len(origData))
	// 加密
	blockMode.CryptBlocks(cryted, origData)
	//使用RawURLEncoding 不要使用StdEncoding
	//不要使用StdEncoding  放在url参数中会导致错误
	return base64.RawURLEncoding.EncodeToString(cryted), nil

}

//AesDecrypt aes 解密
//cryted 加密之后的b64-url-encoding 字符串
//key 对称密钥 密钥长度必须 16/24/32 长度
//返回 解密之后的string
func AesDecrypt(cryted string, key string) (string, error) {
	//使用RawURLEncoding 不要使用StdEncoding
	//不要使用StdEncoding  放在url参数中回导致错误
	crytedByte, err := base64.RawURLEncoding.DecodeString(cryted)
	if err != nil {
		return "", fmt.Errorf("base64.RawURLEncoding: %s", err)
	}
	k := []byte(key)

	// 分组秘钥
	block, err := aes.NewCipher(k)
	if err != nil {
		return "", fmt.Errorf("key 长度必须 16/24/32长度: %s", err)
	}
	// 获取秘钥块的长度
	blockSize := block.BlockSize()
	// 加密模式
	blockMode := cipher.NewCBCDecrypter(block, k[:blockSize])
	// 创建数组
	orig := make([]byte, len(crytedByte))
	// 解密
	blockMode.CryptBlocks(orig, crytedByte)
	// 去补全码
	orig = pkcs7UnPadding(orig)
	return string(orig), nil
}

//pkcs7Padding 补码
func pkcs7Padding(ciphertext []byte, blocksize int) []byte {
	padding := blocksize - len(ciphertext)%blocksize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

//pkcs7UnPadding 去码
func pkcs7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

//AesEncryptJson aes 加密json
func AesEncryptJson(obj interface{}, key string) (s string, err error) {
	origData, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	return AesEncrypt(origData, key)
}
