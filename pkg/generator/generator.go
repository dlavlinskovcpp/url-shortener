package generator

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"

type Generator struct{}

func New() *Generator {
	return &Generator{}
}

func (g *Generator) Generate(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("length must be greater than zero")
	}

	out := make([]byte, length)
	max := big.NewInt(int64(len(charset)))

	for i := range out {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		out[i] = charset[n.Int64()]
	}

	return string(out), nil
}
