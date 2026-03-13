package aliyun

import "testing"

func TestParseCASUploadCertID(t *testing.T) {
	cases := []struct {
		name string
		body string
		want string
	}{
		{
			name: "top-level cert id",
			body: `{"CertId":"abc"}`,
			want: "abc",
		},
		{
			name: "nested certificate id",
			body: `{"Data":{"CertificateId":"xyz"}}`,
			want: "xyz",
		},
		{
			name: "empty",
			body: `{}`,
			want: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseCASUploadCertID([]byte(tc.body))
			if err != nil {
				t.Fatalf("parseCASUploadCertID returned error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestParseCASCertificateList(t *testing.T) {
	body := `{"Data":{"CertificateList":[{"CertId":"id-1","Fingerprint":"fp-1"}]}}`
	items, err := parseCASCertificateList([]byte(body))
	if err != nil {
		t.Fatalf("parseCASCertificateList returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].ID != "id-1" {
		t.Fatalf("unexpected id: %s", items[0].ID)
	}
}
