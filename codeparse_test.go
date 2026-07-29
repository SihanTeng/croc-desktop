package main

import "testing"

func TestNormalizeCodeInput(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"bare code", "4523-ivan-twist-order", "4523-ivan-twist-order"},
		{"bare code with spaces", "  4523-ivan-twist-order \n", "4523-ivan-twist-order"},
		{"custom single-word code", "secret1", "secret1"},
		{"linux hint", `CROC_SECRET="4523-ivan-twist-order" croc`, "4523-ivan-twist-order"},
		{"linux hint unquoted", `CROC_SECRET=4523-ivan-twist-order croc`, "4523-ivan-twist-order"},
		{"windows hint", "croc 4523-ivan-twist-order", "4523-ivan-twist-order"},
		{"command with flags", "croc --relay croc.example.com:9009 4523-ivan-twist-order", "4523-ivan-twist-order"},
		{"share link", "https://getcroc.com/?code=4523-ivan-twist-order&relay=croc.schollz.com:9009", "4523-ivan-twist-order"},
		{"share link path style ignored", "https://example.com/download", ""},
		{"empty", "   ", ""},
		{"two bare words", "hello world", ""},
		{"link without code", "https://example.com/?foo=bar", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := normalizeCodeInput(c.input)
			if c.want == "" {
				if err == nil {
					t.Fatalf("expected an error, got code %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != c.want {
				t.Fatalf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestFriendlyTransferError(t *testing.T) {
	cases := []struct{ in, want string }{
		{"password mismatch", "The relay rejected the connection (password mismatch) — check the relay password in Settings"},
		{"handshake failed: could not secure channel", "Couldn't establish a secure channel — double-check the code and try again"},
		{"could not reconnect to any relay: croc.example.com", "Couldn't reach the relay — check the relay address in Settings and your network connection"},
		{"found no addresses to connect", "Couldn't reach the relay — check the relay address in Settings and your network connection"},
		{"write failed: not enough disk space", "Not enough disk space to receive the files"},
		{"some other error", "some other error"},
	}
	for _, c := range cases {
		if got := friendlyTransferError(c.in); got != c.want {
			t.Errorf("friendlyTransferError(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
