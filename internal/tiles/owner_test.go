//go:build unix

package tiles

import (
	"io/fs"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"testing"
)

// otherUsers is a directory's file info as another user's.
type otherUsers struct{ fs.FileInfo }

func (o otherUsers) Sys() any {
	st := *o.FileInfo.Sys().(*syscall.Stat_t)
	st.Uid++
	return &st
}

// TestARootAnotherUserOwnsIsRefused (D-134): a cache root belongs to the
// user running the program; one another user owns is refused, though its
// mode lets no one else write - its owner could plant tiles in it.
func TestARootAnotherUserOwnsIsRefused(t *testing.T) {
	info, err := os.Stat(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if !ownedHere(info) {
		t.Fatal("a directory this test made reads as another user's")
	}
	if ownedHere(otherUsers{info}) {
		t.Error("a directory another user owns reads as this user's")
	}
	if _, refused := rootRefusal(t.TempDir(), otherUsers{info}, true); !refused {
		t.Error("OpenDisk's check passes a root another user owns")
	}
}

// TestARootWithAnACLIsRefused (D-134): on macOS an access control list can
// let other users write to a root whose mode bits say they cannot; such a
// root is refused.
func TestARootWithAnACLIsRefused(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("NOT RUN: access control lists are checked on macOS")
	}
	dir := t.TempDir()
	if err := exec.Command("chmod", "+a", "everyone allow add_file", dir).Run(); err != nil {
		t.Fatal(err)
	}
	if d, err := OpenDisk(dir, 1<<20); err == nil {
		d.Release()
		t.Fatal("a root carrying an access control list was opened")
	}
	plain := t.TempDir()
	d, err := OpenDisk(plain, 1<<20)
	if err != nil {
		t.Fatalf("a plain root was refused: %v", err)
	}
	d.Release()
}
