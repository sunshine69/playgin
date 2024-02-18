package main

import (
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	smbVFS "playgin/smb"
	"regexp"

	"github.com/gin-gonic/gin"
)

var (
	smbServer      = os.Getenv("SMB_SERVER")
	smbUser        = os.Getenv("SMB_USER")
	smbPassword    = os.Getenv("SMB_PASSWORD")
	smbShare       = os.Getenv("SMB_SHARE")
	chunkSize      = 10 * 1024 * 1024
	slashPtnSuffix = regexp.MustCompile(`[/]+$`)
	slashPtnPrefix = regexp.MustCompile(`^[/]+`)
)

func main() {
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	router.GET("/smb/ls/*path", doList)
	router.GET("/smb/get/*path", doGet)
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home.html", gin.H{})
	})

	router.GET("/albums", getAlbums)
	router.GET("/albums/:id", getAlbumByID)
	router.POST("/albums", postAlbums)

	router.Run("localhost:8080")
}

func doGet(c *gin.Context) {
	s := smbVFS.SmbVFS_Connect(smbServer, smbUser, smbPassword, smbShare, "")
	defer s.Close()
	fullPath := c.Param("path")
	action := c.DefaultQuery("action", "download")
	fullPath = slashPtnPrefix.ReplaceAllString(fullPath, "")
	smbFile, err := s.FS.Open(fullPath)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}
	fmt.Printf("Get File name: %s\n", smbFile.Name())

	w := c.Writer
	header := w.Header()
	header.Set("Transfer-Encoding", "chunked")
	if action == "download" {
		header.Set("Content-Disposition", "attachment; filename="+filepath.Base(smbFile.Name()))
	}
	header.Set("Content-Type", mime.TypeByExtension(filepath.Ext(smbFile.Name())))
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
	s := smbVFS.SmbVFS_Connect("note-local.kaykraft.org", "stevek", smbPassword, "doc", "")
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

// album represents data about a record album.
type album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

// albums slice to seed record album data.
var albums = []album{
	{ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
	{ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
	{ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
}

// getAlbums responds with the list of all albums as JSON.
func getAlbums(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, albums)
}

// postAlbums adds an album from JSON received in the request body.
func postAlbums(c *gin.Context) {
	var newAlbum album

	// Call BindJSON to bind the received JSON to
	// newAlbum.
	if err := c.BindJSON(&newAlbum); err != nil {
		return
	}

	// Add the new album to the slice.
	albums = append(albums, newAlbum)
	c.IndentedJSON(http.StatusCreated, newAlbum)
}

// getAlbumByID locates the album whose ID value matches the id
// parameter sent by the client, then returns that album as a response.
func getAlbumByID(c *gin.Context) {
	id := c.Param("id")

	// Loop through the list of albums, looking for
	// an album whose ID value matches the parameter.
	for _, a := range albums {
		if a.ID == id {
			c.IndentedJSON(http.StatusOK, a)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
}
