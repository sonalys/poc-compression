package utils

import "fmt"

func IntegrityCheck(in, out []byte) error {
	if len(in) != len(out) {
		return fmt.Errorf("output size is different. exp %d != got %d", len(in), len(out))
	}
	for i := range in {
		if in[i] != out[i] {
			return fmt.Errorf("output is different at pos %d exp %d != got %d", i, in[i], out[i])
		}
	}
	return nil
}
