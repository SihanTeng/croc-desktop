package main

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// codeToken matches a code phrase embedded in a larger paste: hyphenated
// words, possibly with a numeric prefix (4523-ivan-twist-order).
var codeToken = regexp.MustCompile(`^[A-Za-z0-9]+(-[A-Za-z0-9]+)+$`)

// crocSecretEnv matches the CROC_SECRET=… prefix of CLI copy-paste hints.
var crocSecretEnv = regexp.MustCompile(`(?i)CROC_SECRET=["']?([^\s"']+)`)

// bareCode matches a whole-input code phrase (custom codes may be a single
// word; StartReceive enforces the minimum length).
var bareCode = regexp.MustCompile(`^[A-Za-z0-9-]+$`)

// normalizeCodeInput extracts a croc code phrase from anything the user might
// paste: a bare code, a CLI hint ("croc <code>", "CROC_SECRET=… croc"), or a
// share link carrying a ?code= parameter.
func normalizeCodeInput(input string) (string, error) {
	s := strings.TrimSpace(input)
	if s == "" {
		return "", fmt.Errorf("no code given")
	}
	// share links carry the code as a query parameter
	if strings.Contains(s, "://") {
		u, err := url.Parse(s)
		if err == nil {
			if code := u.Query().Get("code"); code != "" {
				return code, nil
			}
		}
		return "", fmt.Errorf("no code found in link")
	}
	// "CROC_SECRET=… croc" command hints
	if m := crocSecretEnv.FindStringSubmatch(s); m != nil {
		return m[1], nil
	}
	// "croc <code>" (possibly with flags): last code-looking token wins
	fields := strings.Fields(s)
	for i := len(fields) - 1; i >= 0; i-- {
		if codeToken.MatchString(fields[i]) {
			return fields[i], nil
		}
	}
	// a single word that doesn't look like a phrase may be a custom code
	if len(fields) == 1 && bareCode.MatchString(fields[0]) {
		return fields[0], nil
	}
	return "", fmt.Errorf("no code found — expected something like 4523-ivan-twist-order")
}

// NormalizeCode exposes normalizeCodeInput to the frontend so pasted receive
// instructions "just work".
func (a *App) NormalizeCode(input string) (string, error) {
	return normalizeCodeInput(input)
}
