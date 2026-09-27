package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// CROC_TEST_CLI must point to an unmodified upstream v11 CLI, not the patched
// e2e peer. CI installs the exact upstream version independently of replace.
func TestOfficialCLIInterop(t *testing.T) {
	binary := os.Getenv("CROC_TEST_CLI")
	if binary == "" {
		t.Skip("set CROC_TEST_CLI to an upstream croc v11.5.4 binary")
	}
	for _, direction := range []string{"desktop-send", "desktop-receive"} {
		for _, text := range []bool{false, true} {
			kind := "file"
			if text {
				kind = "text"
			}
			t.Run(direction+"-"+kind, func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
				defer cancel()
				app, events := newTestApp()
				events.onAccept = func() { app.RespondAccept(true) }
				payload := "croc v11 upstream interoperability 🐊"
				source := filepath.Join(t.TempDir(), "interop.txt")
				if err := os.WriteFile(source, []byte(payload), 0600); err != nil {
					t.Fatal(err)
				}
				outputDir := t.TempDir()
				args := []string{"--relay", "127.0.0.1:" + testRelayPorts[0], "--pass", "pass123", "--yes"}
				var code string
				if direction == "desktop-send" {
					var err error
					if text {
						code, err = app.StartSendText(payload)
					} else {
						code, err = app.StartSend([]string{source})
					}
					if err != nil {
						t.Fatal(err)
					}
					args = append(args, "--out", outputDir)
				} else {
					code = fmt.Sprintf("1234-cli-%d", time.Now().UnixNano())
					args = append(args, "send", "--no-local", "--transport", "relay")
					if text {
						args = append(args, "--text", payload)
					} else {
						args = append(args, source)
					}
				}
				cmd := exec.CommandContext(ctx, binary, args...)
				cmd.Dir = t.TempDir()
				cmd.Env = append(os.Environ(), "CROC_SECRET="+code, "CROC_CONFIG_DIR="+t.TempDir())
				var out bytes.Buffer
				cmd.Stdout = &out
				cmd.Stderr = &out
				if err := cmd.Start(); err != nil {
					t.Fatal(err)
				}
				defer func() {
					if cmd.ProcessState == nil {
						cancel()
						_ = cmd.Wait()
					}
				}()
				if direction == "desktop-receive" {
					if err := app.StartReceive(code, outputDir); err != nil {
						t.Fatal(err)
					}
				}
				received := waitDone(t, events, direction)
				if err := cmd.Wait(); err != nil {
					t.Fatalf("upstream CLI: %v\n%s", err, out.String())
				}
				if text {
					if direction == "desktop-send" {
						if !strings.Contains(out.String(), payload) {
							t.Fatalf("upstream CLI did not display text: %s", out.String())
						}
					} else if !received.IsText || received.Text != payload {
						t.Fatalf("desktop text = %+v", received)
					}
					entries, err := os.ReadDir(outputDir)
					if err != nil || len(entries) != 0 {
						t.Fatalf("text artifacts were not cleaned: %v %v", entries, err)
					}
				} else {
					got, err := os.ReadFile(filepath.Join(outputDir, "interop.txt"))
					if err != nil || string(got) != payload {
						t.Fatalf("file contents = %q, %v", got, err)
					}
				}
			})
		}
	}
}
