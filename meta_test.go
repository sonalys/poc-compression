package gompressor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMetadata_Set(t *testing.T) {
	type args struct {
		mask  Mask
		value byte
	}
	tests := []struct {
		name string
		m    Metadata
		args args
		want Metadata
	}{
		{
			name: "empty",
			m:    Metadata(0),
			args: args{
				mask:  0,
				value: 0,
			},
			want: Metadata(0),
		},
		{
			name: "set one bit",
			m:    Metadata(0b10),
			args: args{
				mask:  0b01,
				value: 1,
			},
			want: Metadata(0b11),
		},
		{
			name: "clear one bit",
			m:    Metadata(0b11),
			args: args{
				mask:  0b10,
				value: 0,
			},
			want: Metadata(0b01),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.m.Set(tt.args.mask, tt.args.value)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_SegmentMetaSet(t *testing.T) {
	meta := NewMetadata().
		SetType(TypeRepeatingGroup).
		SetRepSize(1).
		SetPosLenSize(0).
		SetBufLenSize(2).
		SetPosSize(3)

	require.EqualValues(t, TypeRepeatingGroup, meta.GetType())
	require.EqualValues(t, 1, meta.GetRepSize())
	require.EqualValues(t, 0, meta.GetPosLenSize())
	require.EqualValues(t, 2, meta.GetBufLenSize())
	require.EqualValues(t, 3, meta.GetPosSize())
}
