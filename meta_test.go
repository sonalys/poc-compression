package gompressor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_MetaEncoding(t *testing.T) {
	t.Run("repeat group meta", func(t *testing.T) {
		m := MetaRepeatGroup{
			Type:       TypeRepeatingGroup,
			InvertMask: true,
			PosLenSize: MaxSizeUint8,
			BufLenSize: MaxSizeUint16,
			PosSize:    MaxSizeUint8,
		}
		b := m.ToByte()
		got := NewRepeatGroupMeta(b)
		require.Equal(t, m, got)
	})

	t.Run("same char meta", func(t *testing.T) {
		m := MetaSameChar{
			Type:       TypeRepeatingGroup,
			RepeatSize: MaxSizeUint16,
			SinglePos:  true,
			PosLenSize: MaxSizeUint8,
			PosSize:    MaxSizeUint8,
		}
		b := m.ToByte()
		got := NewSameCharMeta(b)
		require.Equal(t, m, got)
	})
}
