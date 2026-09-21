package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/branden-thompson/go-tuimaps/internal/archive"
	"github.com/branden-thompson/go-tuimaps/internal/fetch"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stderr))
}

// run is the command: read the pinned planet archive over the network and
// write the embedded tiles, or check that a directory of them still agrees
// with itself. It is one of the two programs here that reach the network,
// and only when a person runs it (NFR-11).
func run(args []string, stderr io.Writer) int {
	if stderr == nil {
		return 2
	}
	flags := flag.NewFlagSet("gen-assets", flag.ContinueOnError)
	flags.SetOutput(stderr)
	address := flags.String("url", "", "the address of the pinned planet archive; it is never written down")
	name := flags.String("name", "", "the archive's versioned name, recorded in the pin")
	out := flags.String("out", "", "the directory to write: tiles, PIN and HASHES")
	verify := flags.String("verify", "", "check a directory written earlier, and do nothing else")
	if flags.Parse(args) != nil {
		return 2
	}
	var err error
	switch {
	case *verify != "":
		err = Verify(*verify)
	default:
		err = fromNetwork(context.Background(), *address, *name, *out, fetch.Options{Token: "gen-assets"})
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

// fromNetwork generates from the archive at address: a few dozen range
// requests, each held to a 206 with exactly the range asked for.
func fromNetwork(ctx context.Context, address, name, out string, opts fetch.Options) error {
	if ctx == nil {
		return errors.New("gen-assets: no context")
	}
	if address == "" || name == "" || out == "" {
		return errors.New("gen-assets: -url, -name and -out are all required")
	}
	source, err := fetch.ForSource(address, opts)
	if err != nil {
		return err
	}
	info, err := source.Describe(ctx, address, archive.HeaderLen)
	if err != nil {
		return err
	}
	read, err := archive.OverNetwork(source.Fetch, address)
	if err != nil {
		return err
	}
	planet, err := archive.Open(ctx, read, info.Length)
	if err != nil {
		return err
	}
	return Generate(ctx, planet.Tile, Pin{Archive: name, Length: info.Length, EntityTag: info.EntityTag.String()}, out)
}
