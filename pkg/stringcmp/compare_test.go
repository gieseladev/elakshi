package stringcmp

import "testing"

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
			if got := ContainsWords(tt.args.s, tt.args.substring); got != tt.want {
				t.Errorf("ContainsWords() = %v, want %v", got, tt.want)
			}
		})
	}
}
