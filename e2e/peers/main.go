// Test peers for the browser E2E suite: a croc sender, an auto-accepting
// receiver, and a relay, spawned by the playwright tests as subprocesses.
//
// Usage:
//
//	peers relay [--pass pw] <controlPort>            (uses controlPort..controlPort+4)
//	peers send [--relay host:port] [--pass pw] <code> <file>
//	peers recv [--relay host:port] [--pass pw] <code> <outdir>
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/schollz/croc/v10/src/croc"
	"github.com/schollz/croc/v10/src/models"
	"github.com/schollz/croc/v10/src/tcp"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	switch os.Args[1] {
	case "relay":
		cmdRelay(os.Args[2:])
	case "send":
		cmdSend(os.Args[2:])
	case "recv":
		cmdRecv(os.Args[2:])
	default:
		usage()
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: peers relay [--pass pw] <controlPort>")
	fmt.Fprintln(os.Stderr, "       peers send [--relay host:port] [--pass pw] <code> <file>")
	fmt.Fprintln(os.Stderr, "       peers recv [--relay host:port] [--pass pw] <code> <outdir>")
	os.Exit(2)
}

// baseOptions mirrors buildCrocOptions(defaultSettings()) in the app.
func baseOptions(isSender bool, relay, pass string) croc.Options {
	opts := croc.Options{
		IsSender:         isSender,
		RelayAddress:     models.DEFAULT_RELAY,
		RelayAddress6:    models.DEFAULT_RELAY6,
		RelayPorts:       []string{"9009", "9010", "9011", "9012"},
		RelayPassword:    models.DEFAULT_PASSPHRASE,
		Curve:            "p256",
		HashAlgorithm:    "xxhash",
		NoPrompt:         true,
		DisableClipboard: true,
	}
	if relay != "" {
		opts.RelayAddress = relay
		opts.RelayAddress6 = "" // custom relay: same selection logic as the CLI
	}
	if pass != "" {
		opts.RelayPassword = pass
	}
	return opts
}

func transferFlags(args []string) (*flag.FlagSet, *string, *string) {
	fs := flag.NewFlagSet("transfer", flag.ContinueOnError)
	relay := fs.String("relay", "", "relay address (default: public relay)")
	pass := fs.String("pass", "", "relay password (default: croc default)")
	_ = fs.Parse(args)
	return fs, relay, pass
}

func cmdSend(args []string) {
	fs, relay, pass := transferFlags(args)
	if fs.NArg() != 2 {
		usage()
	}
	code, path := fs.Arg(0), fs.Arg(1)
	opts := baseOptions(true, *relay, *pass)
	opts.SharedSecret = code
	filesInfo, emptyFolders, totalFolders, err := croc.GetFilesInfoWithExactExclusions([]string{path}, false, false, nil, nil)
	if err != nil {
		fatal(err)
	}
	cr, err := croc.NewCtx(context.Background(), opts)
	if err != nil {
		fatal(err)
	}
	fmt.Println("SENDER READY")
	if err := cr.Send(filesInfo, emptyFolders, totalFolders); err != nil {
		fatal(err)
	}
	fmt.Println("SEND OK")
}

func cmdRecv(args []string) {
	fs, relay, pass := transferFlags(args)
	if fs.NArg() != 2 {
		usage()
	}
	code, outDir := fs.Arg(0), fs.Arg(1)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fatal(err)
	}
	opts := baseOptions(false, *relay, *pass)
	opts.SharedSecret = code
	cr, err := croc.NewCtx(context.Background(), opts)
	if err != nil {
		fatal(err)
	}
	oldWd, _ := os.Getwd()
	if err := os.Chdir(outDir); err != nil {
		fatal(err)
	}
	fmt.Println("RECEIVER READY")
	if err := cr.Receive(); err != nil {
		fatal(err)
	}
	_ = os.Chdir(oldWd)
	fmt.Println("RECEIVE OK")
}

func cmdRelay(args []string) {
	fs := flag.NewFlagSet("relay", flag.ContinueOnError)
	pass := fs.String("pass", "pass123", "relay password")
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		usage()
	}
	base, err := strconv.Atoi(fs.Arg(0))
	if err != nil {
		fatal(err)
	}
	ctx := context.Background()
	banner := ""
	for i := 1; i <= 4; i++ {
		if i > 1 {
			banner += ","
		}
		banner += strconv.Itoa(base + i)
	}
	for i := 0; i <= 4; i++ {
		port := strconv.Itoa(base + i)
		var banners []string
		if i == 0 {
			banners = append(banners, banner)
		}
		go func(p string, b ...string) {
			fmt.Fprintf(os.Stderr, "relay port %s exited: %v\n", p, tcp.RunCtx(ctx, "error", "", p, *pass, b...))
		}(port, banners...)
	}
	fmt.Println("RELAY READY")
	select {} // run until killed
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
