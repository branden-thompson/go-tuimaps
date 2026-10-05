package tiles

import (
	"os"
	"reflect"
)

// ownedHere reports whether a directory belongs to the user running the
// program: a root another user made could hold tiles they planted, and
// could read the names of the tiles kept in it (D-134). The owner is the
// Uid of the platform's own file record; a platform whose record has none
// (Windows) gives no owner to check, and the directory is taken as owned.
func ownedHere(info os.FileInfo) bool {
	if info == nil {
		return false
	}
	record := reflect.ValueOf(info.Sys())
	if record.Kind() == reflect.Pointer {
		record = record.Elem()
	}
	if record.Kind() != reflect.Struct {
		return true
	}
	uid := record.FieldByName("Uid")
	if !uid.IsValid() || !uid.CanUint() {
		return true
	}
	me := os.Geteuid()
	return me >= 0 && uid.Uint() == uint64(me)
}
