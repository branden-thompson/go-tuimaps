// Command tuimaps is the standalone terminal map: keys, pointer, a describe mode and a one-shot headless render.
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
// for a usage error. It is replaced by work package WP-13.
func run(stderr io.Writer) int {
	fmt.Fprintln(stderr, "tuimaps: not built yet (work package WP-13)")
	return 2
}
