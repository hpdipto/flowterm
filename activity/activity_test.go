//go:build darwin

package activity

import "testing"

func TestParseFront(t *testing.T) {
	tests := []struct {
		name       string
		out        string
		wantBundle string
		wantName   string
		wantErr    bool
	}{
		{
			name: "named app",
			out: `"Rex Beta" ASN:0x0-0xe95e95: (in front) 
    bundleID="com.superlogical.rex"
    bundle path=[ NULL ] 
`,
			wantBundle: "com.superlogical.rex",
			wantName:   "Rex Beta",
		},
		{
			name: "bundle without name",
			out: `ASN:0x0-0x1: (in front) 
    bundleID="com.example.agent"
`,
			wantBundle: "com.example.agent",
			wantName:   "",
		},
		{
			name: "missing bundle id",
			out: `"Rex Beta" ASN:0x0-0xe95e95: (in front) 
    bundle path=[ NULL ] 
`,
			wantErr: true,
		},
		{
			name:    "empty output",
			out:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundle, name, err := parseFront([]byte(tt.out))

			if tt.wantErr {
				if err == nil {
					t.Fatalf("err = nil, want error (bundle %q name %q)", bundle, name)
				}
				return
			}

			if err != nil {
				t.Fatalf("err = %v, want nil", err)
			}
			if bundle != tt.wantBundle {
				t.Errorf("bundle = %q, want %q", bundle, tt.wantBundle)
			}
			if name != tt.wantName {
				t.Errorf("name = %q, want %q", name, tt.wantName)
			}
		})
	}
}
