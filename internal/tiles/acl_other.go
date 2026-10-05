//go:build !darwin

package tiles

// hasACL is false where the root's mode bits are the whole of who may write
// to it, or where (on Windows) the root is not checked at all.
func hasACL(string) bool { return false }
