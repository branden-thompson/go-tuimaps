// Command answer-key computes the independent answer key for the M1 scenarios. It shares no code with the library.
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
// for a usage error. It is replaced by work package WP-11.
func run(stderr io.Writer) int {
	fmt.Fprintln(stderr, "answer-key: not built yet (work package WP-11)")
	return 2
}
