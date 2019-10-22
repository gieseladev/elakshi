package stringcmp

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestContainsSurrounded(t *testing.T) {
	a := assert.New(t)

	a.True(ContainsSurrounded("hello world", "hello"))
	a.True(ContainsSurrounded("hello world", "world"))
	a.True(ContainsSurrounded("hello world", "hello world"))

	a.True(ContainsSurrounded("a b c d", "b c"))

	a.False(ContainsSurrounded("a bc d", "b"))
	a.False(ContainsSurrounded("a bc d", "c"))
}

func TestStringContainsWords(t *testing.T) {
	type args struct {
		s         string
		substring string
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "simple 1",
			args: args{
				s:         "hell owo rld",
				substring: "hello world",
			},
			want: true,
		},
		{
			name: "simple 2",
			args: args{
				s:         "hell of world",
				substring: "hello world",
			},
			want: false,
		},
		{
			name: "contained 1",
			args: args{
				s:         "my what is this over there, sugar crush perhaps?",
				substring: "sugarcrush",
			},
			want: true,
		},
		{
			name: "contained 2",
			args: args{
				s:         "can't come up with more tests",
				substring: "with tests",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContainsSurroundedIgnoreSpace(tt.args.s, tt.args.substring); got != tt.want {
				t.Errorf("ContainsSurroundedIgnoreSpace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetWordsFocusedString(t *testing.T) {
	a := assert.New(t)

	a.Equal("test", GetWordsFocusedString("[test]"))
	a.Equal("h3ll0 w0rl", GetWordsFocusedString("h3ll0__:__w0rl$"))
}
