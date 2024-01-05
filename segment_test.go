package gompressor

import (
	"reflect"
	"testing"
)

func Test_meta_setPosLen(t *testing.T) {
	type args struct {
		size uint8
	}
	tests := []struct {
		name string
		m    meta
		args args
		want meta
	}{
		{
			name: "empty case",
			m:    meta(0),
			args: args{size: 0},
			want: meta(0),
		},
		{
			name: "first set",
			m:    meta(0b00000111),
			args: args{size: 0b11111},
			want: meta(0xff),
		},
		{
			name: "second set",
			m:    meta(0xff),
			args: args{size: 0b11011},
			want: meta(0b11011111),
		},
		{
			name: "clear set",
			m:    meta(0xff),
			args: args{size: 0},
			want: meta(0b00000111),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.setPosLen(tt.args.size); got != tt.want {
				t.Errorf("meta.setPosLen() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_meta_getPosLen(t *testing.T) {
	tests := []struct {
		name string
		m    meta
		want byte
	}{
		{
			name: "empty",
			m:    meta(0),
			want: 0,
		},
		{
			name: "size one",
			m:    meta(0b00001000),
			want: 1,
		},
		{
			name: "size 31",
			m:    meta(0b11111000),
			want: 31,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.getPosLen(); got != tt.want {
				t.Errorf("meta.getPosLen() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_segment_addPos(t *testing.T) {
	type fields struct {
		flags    meta
		pos      []uint32
		repeat   uint16
		buffer   []byte
		previous *segment
		next     *segment
	}
	type args struct {
		pos []uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *segment
		wantErr bool
	}{
		{
			name:    "pos overflow from args",
			fields:  fields{},
			args:    args{make([]uint32, 32)},
			want:    &segment{},
			wantErr: true,
		},
		{
			name: "pos overflow from total",
			fields: fields{
				pos: make([]uint32, 16),
			},
			args: args{make([]uint32, 16)},
			want: &segment{
				pos: make([]uint32, 16),
			},
			wantErr: true,
		},
		{
			name: "no initial pos",
			fields: fields{
				flags:  meta(0),
				pos:    []uint32{},
				repeat: 1,
				buffer: []byte{1},
			},
			args: args{pos: []uint32{1, 2}},
			want: &segment{
				flags:  meta(0b00010000),
				pos:    []uint32{1, 2},
				repeat: 1,
				buffer: []byte{1},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &segment{
				flags:    tt.fields.flags,
				pos:      tt.fields.pos,
				repeat:   tt.fields.repeat,
				buffer:   tt.fields.buffer,
				previous: tt.fields.previous,
				next:     tt.fields.next,
			}
			got, err := s.addPos(tt.args.pos)
			if (err != nil) != tt.wantErr {
				t.Errorf("segment.addPos() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("segment.addPos() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_meta_SetIsRepeat2Bytes(t *testing.T) {
	tests := []struct {
		name string
		arg  bool
		m    meta
		want meta
	}{
		{
			name: "false to false",
			m:    meta(0b11111000),
			arg:  false,
			want: meta(0b11111000),
		},
		{
			name: "false to true",
			m:    meta(0b11111000),
			arg:  true,
			want: meta(0b11111100),
		},
		{
			name: "true to false",
			m:    meta(0b11111100),
			arg:  false,
			want: meta(0b11111000),
		},
		{
			name: "true to false",
			m:    meta(0b11111100),
			arg:  true,
			want: meta(0b11111100),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.setIsRepeat2Bytes(tt.arg); got != tt.want {
				t.Errorf("meta.SetIsRepeat2Bytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_meta_setType(t *testing.T) {
	type args struct {
		t segType
	}
	tests := []struct {
		name string
		m    meta
		args args
		want meta
	}{
		{
			name: "empty",
			m:    meta(0b11111100),
			args: args{typeUncompressed},
			want: meta(0b11111100),
		},
		{
			name: "repeat",
			m:    meta(0b11111100),
			args: args{typeRepeat},
			want: meta(0b11111101),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.setType(tt.args.t); got != tt.want {
				t.Errorf("meta.setType() = %v, want %v", got, tt.want)
			}
		})
	}
}
