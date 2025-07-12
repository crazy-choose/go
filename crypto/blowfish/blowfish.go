package blowfish

import (
	"crypto/cipher"
	"github.com/andreburgaud/crypt2go/ecb"
	"github.com/andreburgaud/crypt2go/padding"
	"golang.org/x/crypto/blowfish"
)

type BlowFish struct {
	cipher    *blowfish.Cipher
	padder    padding.Padding
	encrypter cipher.BlockMode
	decrypter cipher.BlockMode
}

func New(key string) (*BlowFish, error) {
	block, err := blowfish.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}

	return &BlowFish{
		cipher:    block,
		padder:    newZeroPadding(),
		encrypter: ecb.NewECBEncrypter(block),
		decrypter: ecb.NewECBDecrypter(block),
	}, nil
}

func (b *BlowFish) Encrypt(pt []byte) ([]byte, error) {
	pt, err := b.padder.Pad(pt) // pad last block of plaintext if block size less than block cipher size
	if err != nil {
		return nil, err
	}
	ct := make([]byte, len(pt))
	b.encrypter.CryptBlocks(ct, pt)
	return ct, nil
}

func (b *BlowFish) Decrypt(ct []byte) ([]byte, error) {
	pt := make([]byte, len(ct))
	b.decrypter.CryptBlocks(pt, ct)
	pt, err := b.padder.Unpad(pt) // pad last block of plaintext if block size less than block cipher size
	if err != nil {
		return nil, err
	}
	return pt, nil
}
