// Command gen-assets builds the embedded zoom 0 to 3 tiles from the pinned planet archive.
package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	os.Exit(run(os.Stderr))
}

// run reports that the command is not built yet and returns the exit code
// for a usage error. It is replaced by work package WP-04.
func run(stderr io.Writer) int {
	fmt.Fprintln(stderr, "gen-assets: not built yet (work package WP-04)")
	return 2
}
