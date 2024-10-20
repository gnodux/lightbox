package cryptlib

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"errors"
)

// PKCS7Padding PKCS7 填充模式
func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	//Repeat()函数的功能是把切片[]byte{byte(padding)}复制padding个，然后合并成新的字节切片返回
	paddingText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, paddingText...)
}

// PKCS7UnPadding 填充的反向操作，删除填充字符串
func PKCS7UnPadding(origData []byte) ([]byte, error) {
	//获取数据长度
	length := len(origData)
	if length == 0 {
		return nil, errors.New("加密字符串错误！")
	} else {
		//获取填充字符串长度
		unPadding := int(origData[length-1])
		//截取切片，删除填充字节，并且返回明文
		return origData[:(length - unPadding)], nil
	}
}

var defaultKey = []byte{'1', '9', '7', '8', '1', '1', '0', '7', '2', '0', '1', '8', '1', '2', '2', '4'}
var defaultPadding = []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p'}

func KeyPadding(key []byte) []byte {
	if len(key) == 0 {
		return defaultKey
	}
	l := len(key)
	switch l {
	case 16, 24, 32:
		return key
	default:
		if l < 16 {
			return append(key, defaultPadding[0:16-l]...)
		}
		if l < 24 {
			return append(key, defaultPadding[0:24-l]...)
		}
		if l < 32 {
			return append(key, defaultPadding[0:32-l]...)
		}
		return key[0:32]
	}
}

func AesEncrypt(origData []byte, key []byte) ([]byte, error) {

	key = KeyPadding(key)
	//创建加密算法实例
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	//获取块的大小
	blockSize := block.BlockSize()
	//对数据进行填充，让数据长度满足需求
	origData = PKCS7Padding(origData, blockSize)

	//采用AES加密方法中CBC加密模式
	blocMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	encrypted := make([]byte, len(origData))
	//执行加密
	blocMode.CryptBlocks(encrypted, origData)
	return encrypted, nil
}

// AesDecrypt 实现解密
func AesDecrypt(encrypted []byte, key []byte) ([]byte, error) {
	key = KeyPadding(key)
	//创建加密算法实例
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	//获取块大小
	blockSize := block.BlockSize()
	//创建加密客户端实例
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(encrypted))
	//这个函数也可以用来解密
	blockMode.CryptBlocks(origData, encrypted)
	//去除填充字符串
	origData, err = PKCS7UnPadding(origData)
	if err != nil {
		return nil, err
	}
	return origData, err
}
