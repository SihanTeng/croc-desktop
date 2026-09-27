package croc

import (
	"bytes"
	"math"
	"os"
	"testing"

	"github.com/schollz/croc/v11/src/models"
)

func TestDesktopHookRunsAfterReceiveValidation(t *testing.T) {
	for _, tc := range []struct {
		name string
		file FileInfo
		text bool
	}{
		{"text filename", FileInfo{Name: "croc-text-test", FolderRemote: ".", Size: 1}, true},
		{"text size", FileInfo{Name: "croc-stdin-test", FolderRemote: ".", Size: maxTextTransferBytes + 1}, true},
		{"traversal", FileInfo{Name: "outside", FolderRemote: "../", Size: 1}, false},
		{"symlink", FileInfo{Name: "link", FolderRemote: ".", Symlink: "../../outside"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := &Client{Options: Options{NoPrompt: true}}
			called := false
			client.SetHooks(&Hooks{OnAcceptRequest: func(AcceptRequest) bool { called = true; return true }})
			done, err := client.processSenderInfo(SenderInfo{FilesToTransfer: []FileInfo{tc.file}, SendingText: tc.text})
			if err == nil || !done {
				t.Fatalf("unsafe offer accepted: done=%v, err=%v", done, err)
			}
			if called {
				t.Fatal("unsafe metadata reached the GUI acceptance callback")
			}
		})
	}
}

func TestDesktopRelayOnlyDoesNotAdvertiseTailcat(t *testing.T) {
	client := &Client{Options: Options{DisableTailcat: true}}
	if client.localTailcatSupported() {
		t.Fatal("relay-only client advertises Tailcat")
	}
	client.startTailcatPreparation()
	if client.tailcat.prepareReady != nil {
		t.Fatal("relay-only client started Tailcat preparation")
	}
}

func TestDesktopDeclinedArchiveSurvivesReconnect(t *testing.T) {
	t.Chdir(t.TempDir())
	writeTestZip(t, "keep.zip", map[string]string{"mine.txt": "mine"})
	original, err := os.ReadFile("keep.zip")
	if err != nil {
		t.Fatal(err)
	}
	offered := FileInfo{Name: "keep.zip", FolderRemote: ".", Size: 99999, TempFile: true}
	client := overwriteTestClient(t, offered, Options{HashAlgorithm: defaultHashAlgorithm, NoPrompt: true})
	prompts := 0
	client.SetHooks(&Hooks{
		OnAcceptRequest: func(AcceptRequest) bool { return true },
		OnOverwriteRequest: func(string, float64) bool {
			prompts++
			return false
		},
	})
	if err := client.updateIfRecipientHasFileInfo(); err != nil {
		t.Fatal(err)
	}
	client.resetLifecycle()
	if _, err := client.processSenderInfo(SenderInfo{FilesToTransfer: []FileInfo{offered}}); err != nil {
		t.Fatal(err)
	}
	if err := client.extractReceivedArchives(); err != nil {
		t.Fatal(err)
	}
	if prompts != 1 {
		t.Fatalf("overwrite prompts = %d, want exactly one", prompts)
	}
	got, err := os.ReadFile("keep.zip")
	if err != nil || !bytes.Equal(got, original) {
		t.Fatalf("declined archive was changed or removed: %v", err)
	}
	if _, err := os.Stat("mine.txt"); !os.IsNotExist(err) {
		t.Fatalf("declined archive was extracted: %v", err)
	}
}

func TestDesktopResumePercentage(t *testing.T) {
	t.Chdir(t.TempDir())
	chunk := models.TCP_BUFFER_SIZE / 2
	payload := append(bytes.Repeat([]byte{1}, 2*chunk), make([]byte, chunk)...)
	if err := os.WriteFile("partial.bin", payload, 0600); err != nil {
		t.Fatal(err)
	}
	client := &Client{}
	called := false
	client.SetHooks(&Hooks{OnOverwriteRequest: func(name string, percent float64) bool {
		called = true
		if name != "partial.bin" || math.Abs(percent-200.0/3) > 0.01 {
			t.Errorf("resume prompt = %q %.2f%%, want partial.bin 66.67%%", name, percent)
		}
		return true
	}})
	accepted, err := client.askReceiveOverwrite(FileInfo{Name: "partial.bin", FolderRemote: ".", Size: int64(len(payload))}, true)
	if err != nil || !accepted || !called {
		t.Fatalf("resume response = %v, %v; callback=%v", accepted, err, called)
	}
}
