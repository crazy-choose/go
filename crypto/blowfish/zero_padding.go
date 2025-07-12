package blowfish

import (
	"bytes"
	"errors"
	"github.com/andreburgaud/crypt2go/padding"
)

type ZeroPadder struct {
	blockSize int
}

func newZeroPadding() padding.Padding {
	return &ZeroPadder{blockSize: 8}
}

func (p *ZeroPadder) Pad(buf []byte) ([]byte, error) {
	bufLen := len(buf)
	padLen := p.blockSize - (bufLen % p.blockSize)
	padText := bytes.Repeat([]byte{byte(0)}, padLen)
	return append(buf, padText...), nil
}

func (p *ZeroPadder) Unpad(buf []byte) ([]byte, error) {
	bufLen := len(buf)
	if bufLen == 0 {
		return nil, errors.New("blowfish/zeropadding: invalid padding size")
	}

	for i := len(buf) - 1; i >= 0; i-- {
		if buf[i] != 0 {
			return buf[:i+1], nil
		}
	}

	return nil, errors.New("blowfish/zeropadding: unpad failed")
}
