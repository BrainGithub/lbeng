package utils

import (
    "bytes"
    "crypto/aes"
)

//ECBEncrypt
func ECBEncrypt(text []byte) []byte {
    key := "abcdefgh12345678"
    return encrypt(pPKCS7Pad(text), key)
}

//ECBDecrypt
func ECBDecrypt(text []byte) []byte {
    key := "abcdefgh12345678"
    if len(text) % 16 == 0 {
        return decrypt(text, key)
    } else {
        return text
    }
}

// 加密
func encrypt(plaintext []byte, key string) []byte {
    cipher, err := aes.NewCipher([]byte(key[:aes.BlockSize]))
    if err != nil {
        panic(err.Error())
    }

    if len(plaintext)%aes.BlockSize != 0 {
        panic("Need a multiple of the blocksize 16")
    }

    ciphertext := make([]byte, 0)
    text := make([]byte, 16)
    for len(plaintext) > 0 {
        // 每次运算一个block
        cipher.Encrypt(text, plaintext)
        plaintext = plaintext[aes.BlockSize:]
        ciphertext = append(ciphertext, text...)
    }
    return ciphertext
}

// 解密
func decrypt(ciphertext []byte, key string) []byte {
    cipher, err := aes.NewCipher([]byte(key[:aes.BlockSize]))
    if err != nil {
        panic(err.Error())
    }

    if len(ciphertext)%aes.BlockSize != 0 {
        panic("Need a multiple of the blocksize 16")
    }

    plaintext := make([]byte, 0)
    text := make([]byte, 16)
    for len(ciphertext) > 0 {
        cipher.Decrypt(text, ciphertext)
        ciphertext = ciphertext[aes.BlockSize:]
        plaintext = append(plaintext, text...)
    }
    return plaintext
}

// Padding补全' '
func pPKCS7Pad(data []byte) []byte {
    padding := aes.BlockSize - len(data)%aes.BlockSize
    padtext := bytes.Repeat([]byte{byte(' ')}, padding)
    return append(data, padtext...)
}

func pPKCS7UPad(data []byte) []byte {
    padLength := int(data[len(data)-1])
    return data[:len(data)-padLength]
}
