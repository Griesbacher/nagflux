package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

//https://gist.github.com/DeanThompson/17056cc40b4899e3e7f4
type ecb struct {
	b         cipher.Block
	blockSize int
}

func newECB(b cipher.Block) *ecb {
	return &ecb{
		b:         b,
		blockSize: b.BlockSize(),
	}
}

//AESECBDecrypter can decrypt aes ecb.
type AESECBDecrypter ecb

//NewAESECBDecrypter generates a new AESECBDecrypter
func NewAESECBDecrypter(key []byte) (*AESECBDecrypter, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return (*AESECBDecrypter)(newECB(block)), nil
}

//CryptBlocks encrypts the given array and saves it into dst
func (d *AESECBDecrypter) CryptBlocks(dst, src []byte) {
	if len(src)%d.blockSize != 0 {
		panic("crypto/cipher: input not full blocks")
	}
	for len(src) > 0 {
		d.b.Decrypt(dst, src[:d.blockSize])
		src = src[d.blockSize:]
		dst = dst[d.blockSize:]
	}
}

//Decypt decrpts the given array by using base64decodeing and aes-ecb.
func (d *AESECBDecrypter) Decypt(data []byte) ([]byte, error) {
	raw := make([]byte, len(data))
	decoded, err := base64.StdEncoding.Decode(raw, data)
	if err != nil {
		return nil, err
	}
	raw = raw[:decoded]
	dest := make([]byte, decoded)
	d.CryptBlocks(dest, raw)
	return dest, nil
}
