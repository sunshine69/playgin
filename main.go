package main

import (
	smbVFS "devops_app_go/smb"
	"fmt"
	"io"
	iofs "io/fs"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gomarkdown/markdown"
)

var (
	smbServer      = os.Getenv("SMB_SERVER")
	smbUser        = os.Getenv("SMB_USER")
	smbPassword    = os.Getenv("SMB_PASSWORD")
	smbDomain      = os.Getenv("SMB_USER_DOMAIN")
	smbShare       = os.Getenv("SMB_SHARE")
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
	router.GET("/smb/get", doGet)
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
	fullPath := c.Query("path")
	fmt.Printf("File path 1: '%s'\n", fullPath)
	action := c.DefaultQuery("action", "download")
	fullPath, err := url.QueryUnescape(fullPath)
	if err != nil {
		fmt.Printf("ERROR PathUnescape %s\n", err.Error())
	}
	fullPath = slashPtnPrefix.ReplaceAllString(fullPath, "")
	fmt.Printf("File path: '%s'\n", fullPath)
	smbFile, err := s.FS.Open(fullPath)
	if err != nil {
		fmt.Printf("ERROR FS.Open %s\n", err.Error())
	}
	fileName := filepath.Base(strings.ReplaceAll(fullPath, `\`, `/`))
	fmt.Printf("Get File name: %s | '%s'\n", smbFile.Name(), fileName)

	w := c.Writer
	header := w.Header()
	contentType := mime.TypeByExtension(filepath.Ext(fileName))
	println("Content-Type detected " + contentType)
	header.Set("Content-Type", contentType)
	if action == "download" {
		header.Set("Transfer-Encoding", "chunked")
		header.Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")
		io.Copy(w, smbFile)
	} else {
		header.Set("Content-Disposition", "inline; filename=\""+fileName+"\"")
		if strings.HasSuffix(fileName, ".md") || strings.HasSuffix(fileName, ".MD") {
			fileContent, _ := s.FS.ReadFile(fileName)
			htmlContent := markdown.ToHTML(fileContent, nil, nil)
			w.Write(htmlContent)
		} else {
			io.Copy(w, smbFile)
		}
	}
	w.Flush()
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
