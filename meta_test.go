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
			require.Equal(t, tt.want, *got)
		})
	}
}
