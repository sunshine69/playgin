package main

import (
	smbVFS "playgin/smb"
	"fmt"
	iofs "io/fs"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	smbServer      = os.Getenv("SMB_SERVER")
	smbUser        = os.Getenv("SMB_USER")
	smbPassword    = os.Getenv("SMB_PASSWORD")
	smbDomain      = os.Getenv("SMB_USER_DOMAIN")
	smbShare       = os.Getenv("SMB_SHARE")
	chunkSize      = 10 * 1024 * 1024
	slashPtnSuffix = regexp.MustCompile(`[/]+$`)
	slashPtnPrefix = regexp.MustCompile(`^[/]+`)
	httpPort       = os.Getenv("HTTP_PORT")
)

func init() {
	if smbDomain == "" {
		smbDomain = smbServer
	}
}

func main() {
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	router.GET("/smb/ls/*path", doList)
	router.GET("/smb/get/*path", doGet)
	router.POST("/smb/search", doSearch)

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home.html", gin.H{})
	})
	if httpPort == "" {
		httpPort = "8080"
	}
	router.Run("0.0.0.0:" + httpPort)
}

func doSearch(c *gin.Context) {
	keyword := c.PostForm("keyword")
	if keyword == "" {
		c.Writer.WriteString("Keyword search required")
		return
	}
	rootDir := c.PostForm("rootdir")
	if rootDir == "" {
		rootDir = "."
	}
	s := smbVFS.SmbVFS_Connect(smbServer, smbUser, smbPassword, smbShare, smbDomain, "")
	defer s.Close()
	namePtn, err := regexp.Compile(keyword)
	if err != nil {
		c.Writer.WriteString("<html><body>ERROR. Search pattern is in golang regexp format which is mostly PCRE compatible. See <a href=\"https://github.com/google/re2/wiki/Syntax\">https://github.com/google/re2/wiki/Syntax for details</a></body>")
		return
	}
	matches := []string{}
	err = iofs.WalkDir(s.FS.DirFS(rootDir), ".", func(path string, d iofs.DirEntry, err error) error {
		if !d.IsDir() && namePtn.MatchString(d.Name()) {
			// fmt.Println(path, d.Name(), err)
			path = strings.ReplaceAll(path, `\`, `/`)
			matches = append(matches, path)
			// return iofs.SkipAll
		}
		return nil
	})
	if err != nil {
		c.Writer.WriteString(err.Error())
	} else {
		if rootDir == "." {
			rootDir = ""
		} else {
			rootDir = slashPtnPrefix.ReplaceAllString(rootDir, "")
			rootDir = slashPtnSuffix.ReplaceAllString(rootDir, "")
		}
		c.HTML(http.StatusOK, "search_result.html", gin.H{
			"title":    "Explore smb share",
			"myheader": "search result",
			"matches":  matches,
			"rootdir":  rootDir,
		})
		// c.Writer.WriteString(fmt.Sprintf("%q", matches))
	}
}

func doGet(c *gin.Context) {
	s := smbVFS.SmbVFS_Connect(smbServer, smbUser, smbPassword, smbShare, smbDomain, "")
	defer s.Close()
	fullPath := c.Param("path")
	action := c.DefaultQuery("action", "download")
	fullPath = slashPtnPrefix.ReplaceAllString(fullPath, "")
	smbFile, err := s.FS.Open(fullPath)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}
	fileName := filepath.Base(strings.ReplaceAll(fullPath, `\`, `/`))
	fmt.Printf("Get File name: %s | '%s'\n", smbFile.Name(), fileName)

	w := c.Writer
	header := w.Header()

	header.Set("Content-Type", mime.TypeByExtension(filepath.Ext(fileName)))
	if action == "download" {
		header.Set("Transfer-Encoding", "chunked")
		header.Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")
	} else {
		header.Set("Content-Disposition", "inline; filename=\""+fileName+"\"")
	}
	w.WriteHeader(http.StatusOK)

	buff := make([]byte, chunkSize)

	for {
		readBytes, err := smbFile.Read(buff)
		if err != nil {
			fmt.Printf("%s\n", err.Error())
			w.WriteString("ERROR READ FILE")
		}
		if readBytes < chunkSize {
			buff = buff[:readBytes]
			w.Write(buff)
			w.(http.Flusher).Flush()
			break
		}
		w.Write(buff)
	}
	w.(http.Flusher).Flush()
}

// List directory content or files
func doList(c *gin.Context) {
	s := smbVFS.SmbVFS_Connect(smbServer, smbUser, smbPassword, smbShare, smbDomain, "")
	defer s.Close()
	fullPath := c.Param("path")

	fullPath = slashPtnSuffix.ReplaceAllString(fullPath, "")
	fullPath = slashPtnPrefix.ReplaceAllString(fullPath, "")
	updir := filepath.Dir(fullPath)
	if updir == "." || updir == "/" {
		updir = ""
	}

	c.HTML(http.StatusOK, "list.html", gin.H{
		"title":    "Explore smb share",
		"myheader": "List files",
		"finfo":    s.Ls(fullPath),
		"rootPath": fullPath,
		"updir":    updir,
	})
}
