package provider

import "testing"

func TestNormalizeSMBDirectDest(t *testing.T) {
	tests := []struct {
		name       string
		rawProduct string
		code       string
		bankPrefix string
		dest       string
		want       string
	}{
		{
			name:       "plain target already prefixed once",
			rawProduct: "BIFASTOPEN2:002",
			code:       "BIFASTOPEN2",
			bankPrefix: "002",
			dest:       "002601072763509",
			want:       "002601072763509",
		},
		{
			name:       "dest accidentally starts with raw product prefix",
			rawProduct: "BIFASTOPEN2:002",
			code:       "BIFASTOPEN2",
			bankPrefix: "002",
			dest:       "BIFASTOPEN2:002002601072763509",
			want:       "002601072763509",
		},
		{
			name:       "dest without bank prefix gets prefixed",
			rawProduct: "BIFASTOPEN2:002",
			code:       "BIFASTOPEN2",
			bankPrefix: "002",
			dest:       "601072763509",
			want:       "002601072763509",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeSMBDirectDest(tc.rawProduct, tc.code, tc.bankPrefix, tc.dest)
			if got != tc.want {
				t.Fatalf("normalizeSMBDirectDest() = %q, want %q", got, tc.want)
			}
		})
	}
}
