package main

import (
	"reflect"
	"testing"
)

func Test_compress(t *testing.T) {
	type args struct {
		in             []byte
		compressionMap map[string][]byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{"Basic test", args{[]byte("testcompress"), map[string][]byte{"es": []byte{128}}}, []byte{'t', 128, 't', 'c', 'o', 'm', 'p', 'r', 128, 's'}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := compress(tt.args.in, tt.args.compressionMap); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Compress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_decompress(t *testing.T) {
	type args struct {
		filename       string
		compressionMap map[string][]byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{"Basic test", args{"testcompress.txt", map[string][]byte{"es": []byte{128}}}, []byte("testcompress")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := decompress(tt.args.filename, tt.args.compressionMap); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Decompress() = %v, want %v", got, tt.want)
			}
		})
	}
}
