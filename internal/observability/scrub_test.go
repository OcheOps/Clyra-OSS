package observability

import "testing"

func TestScrubString_RedactsKnownShapes(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"google key in url (real 39-char length)",
			`Post ".../gemini-1.5-pro:generateContent?key=AIzaSyA1B2C3D4E5F6G7H8I9J0K1L2M3N4O5P6Q": dial tcp: lookup ...`,
			`Post ".../gemini-1.5-pro:generateContent?key=[REDACTED]": dial tcp: lookup ...`},
		{"openai",
			"authorization: sk-abcDEF1234567890abcDEF1234567890abc", "authorization: [REDACTED]"},
		{"anthropic",
			"key sk-ant-api03-abcDEF_1234567890abcDEF1234567890abc", "key [REDACTED]"},
		{"clyra pat",
			"token clyra_pat_abcdef0123456789abcdef0123456789", "token [REDACTED]"},
		{"aws key id",
			"aws_access_key_id=AKIAIOSFODNN7EXAMPLE in logs", "aws_access_key_id=[REDACTED] in logs"},
		{"polar whsec",
			"whsec_abcDEF0123456789ABCdefg received", "[REDACTED] received"},
		{"bearer header",
			"Authorization: Bearer abcDEF0123456789ABCdefg0123456789", "Authorization: [REDACTED]"},
		{"no match passes through",
			"a benign request /api/v1/health", "a benign request /api/v1/health"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := scrubString(c.in); got != c.want {
				t.Fatalf("\nin:   %q\nwant: %q\ngot:  %q", c.in, c.want, got)
			}
		})
	}
}

func TestScrubValue_ScrubsErrorMessages(t *testing.T) {
	err := errorsNew("tls error AIzaSyA1B2C3D4E5F6G7H8I9J0K1L2M3N4O5P6Q")
	got, ok := scrubValue(err).(error)
	if !ok {
		t.Fatal("expected error back")
	}
	if got.Error() != "tls error [REDACTED]" {
		t.Fatalf("got %q", got.Error())
	}
}

// Tiny local alias so this test file doesn't need to import "errors".
func errorsNew(s string) error { return simpleErr(s) }

type simpleErr string

func (e simpleErr) Error() string { return string(e) }
