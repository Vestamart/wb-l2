package unpack

import "testing"

func TestUnpack(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "basic",
			input: "a4bc2d5e",
			want:  "aaaabccddddde",
		},
		{
			name:  "no digits",
			input: "abcd",
			want:  "abcd",
		},
		{
			name:    "only digits",
			input:   "45",
			wantErr: true,
		},
		{
			name:  "empty",
			input: "",
			want:  "",
		},
		{
			name:  "escaped digits",
			input: `qwe\4\5`,
			want:  "qwe45",
		},
		{
			name:  "escaped then count",
			input: `qwe\45`,
			want:  "qwe44444",
		},
		{
			name:    "dangling escape",
			input:   `abc\`,
			wantErr: true,
		},
		{
			name:  "zero count",
			input: "a0b1c2",
			want:  "bcc",
		},
		{
			name:  "multi-digit count",
			input: "a12",
			want:  "aaaaaaaaaaaa",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := Unpack(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Unpack(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Fatalf("Unpack(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
