package smb

import (
	"fmt"
	"io/fs"
	"net"

	"github.com/hirochachacha/go-smb2"
)

type SmbVFS struct {
	Server     string
	Port       string
	Domain     string
	Username   string
	Password   string
	Sharename  string
	conn       net.Conn
	smbSession *smb2.Session
	FS         *smb2.Share
}

// List content of a path
func (smbVFS *SmbVFS) Ls(path string) []fs.FileInfo {
	listDir, err := smbVFS.FS.ReadDir(path)
	if err != nil {
		finfo, err := smbVFS.FS.Stat(path)
		if err != nil {
			fmt.Printf("%s\n", err.Error())
			return []fs.FileInfo{}
		} else {
			return []fs.FileInfo{finfo}
		}
	}

	return listDir
}

func (smbVFS *SmbVFS) Getfile(path string) {

}

func SmbVFS_Connect(server, username, password, sharename, domain, port string) *SmbVFS {
	if port == "" {
		port = "445"
	}
	if domain == "" {
		domain = server
	}
	this := SmbVFS{
		Server:    server,
		Port:      port,
		Username:  username,
		Password:  password,
		Sharename: sharename,
		Domain:    domain,
	}
	var err error
	this.conn, err = net.Dial("tcp", fmt.Sprintf("%s:%s", this.Server, this.Port))
	if err != nil {
		panic(err)
	}

	d := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     this.Username,
			Password: this.Password,
			Domain:   this.Domain,
		},
	}

	this.smbSession, err = d.Dial(this.conn)
	if err != nil {
		panic(err)
	}

	this.FS, err = this.smbSession.Mount(this.Sharename)
	if err != nil {
		panic(err)
	}

	return &this
}

func (smbVFS *SmbVFS) Close() {
	smbVFS.FS.Umount()
	smbVFS.smbSession.Logoff()
	smbVFS.conn.Close()
}
