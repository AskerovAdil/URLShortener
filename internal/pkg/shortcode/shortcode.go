package shortcode

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func Generate(length int) (string, error) {
	if length <= 0 {
		length = 8
	}

	out := make([]byte, length)
	max := big.NewInt(int64(len(alphabet)))

	for i := range out {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", fmt.Errorf("random int: %w", err)
		}
		out[i] = alphabet[n.Int64()]
	}

	return string(out), nil
}
