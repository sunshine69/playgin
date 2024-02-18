package smb

import (
	"fmt"
	"os"
	"testing"
)

func checkErr(err error) {
	panic(fmt.Sprintf("%s\n", err.Error()))
}
func TestCreatingSmbVFS(t *testing.T) {
	password := os.Getenv("SMB_PASSWORD")
	smbVFS := SmbVFS_Connect("note-local.kaykraft.org", "stevek", password, "doc", "")
	defer smbVFS.Close()
	// path := "ubuntu-22.04.1-desktop-amd64.iso"

	// matches, err := iofs.Glob(smbVFS.FS.DirFS("."), "*")
	// if err != nil {
	// 	panic(err)
	// }
	// for _, match := range matches {
	// 	fmt.Println(match)
	// }

}
