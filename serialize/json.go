package serialize

import (
	"github.com/bytedance/sonic"
)

func Marshal(input any) ([]byte, error) {
	return sonic.Marshal(input)
}

func Unmarshal(input []byte, output any) error {
	return sonic.Unmarshal(input, output)
}

