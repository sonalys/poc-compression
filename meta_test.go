package gompressor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_MetaEncoding(t *testing.T) {
	m := Meta{
		Type:          TypeRepeatingGroup,
		InvertBitmask: true,
		RepeatSize:    MaxSizeUint16,
		PosLenSize:    MaxSizeUint8,
		BufLenSize:    MaxSizeUint16,
		PosSize:       MaxSizeUint8,
	}
	b := m.ToByte()
	got := NewMeta2(b)
	require.Equal(t, m, got)
}
