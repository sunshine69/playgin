package smb

import (
	"fmt"
	"regexp"
	"testing"
)

func checkErr(err error) {
	panic(fmt.Sprintf("%s\n", err.Error()))
}
func TestCreatingSmbVFS(t *testing.T) {
	// password := os.Getenv("SMB_PASSWORD")
	// smbVFS := SmbVFS_Connect("note-local.kaykraft.org", "stevek", password, "doc", "", "")
	// defer smbVFS.Close()
	//path := "ubuntu-22.04.1-desktop-amd64.iso"

	// matches, err := smbVFS.FS.Glob("*/***.iso")
	// if err != nil {
	// 	panic(err)
	// }
	// for _, match := range matches {
	// 	fmt.Println(match)
	// }
	namePtn := regexp.MustCompile(`^.*\.iso$`)
	if namePtn.MatchString("thhh.iso.fgfgf") {
		println("Matched")
	}

	// err := iofs.WalkDir(smbVFS.FS.DirFS("."), ".", func(path string, d iofs.DirEntry, err error) error {
	// 	if namePtn.MatchString(d.Name()) {
	// 		fmt.Println(path, d.Name(), err)
	// 		return iofs.SkipAll
	// 	}
	// 	return nil
	// })
	// if err != nil {
	// 	panic(err)
	// }
}
