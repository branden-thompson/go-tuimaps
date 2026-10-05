//go:build darwin

package tiles

import (
	"syscall"
	"unsafe"
)

// attrCmnExtendedSecurity asks getattrlist(2) for a file's access control
// list, if it has one.
const attrCmnExtendedSecurity = 0x00400000

// hasACL reports whether a directory carries an access control list, which
// can grant other users what its mode bits do not show. A directory the
// call cannot answer for is taken to carry one: the caller refuses it.
func hasACL(dir string) bool {
	path, err := syscall.BytePtrFromString(dir)
	if err != nil {
		return true
	}
	list := struct {
		bitmapCount uint16
		reserved    uint16
		common      uint32
		vol         uint32
		dir         uint32
		file        uint32
		fork        uint32
	}{bitmapCount: 5, common: attrCmnExtendedSecurity}
	var buf struct {
		length uint32
		offset int32
		size   uint32
		data   [64]byte
	}
	_, _, errno := syscall.Syscall6(syscall.SYS_GETATTRLIST, uintptr(unsafe.Pointer(path)), uintptr(unsafe.Pointer(&list)),
		uintptr(unsafe.Pointer(&buf)), unsafe.Sizeof(buf), 0, 0)
	if errno != 0 {
		return true
	}
	return buf.size > 0
}
