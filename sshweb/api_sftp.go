package sshweb

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"path/filepath"
	"strings"
)

// apiSftpLs sftp show dir list
// @Summary show all items in a path
// @Description list all items of a directory by sftp
// @Tags SFTP
// @Accept  json
// @Produce  json
// @Param token query string true "TOKEN"
// @Param dir query string true "the path of a dir"
// @Success 200 {object} responseObj{data=bastion.ResReadDir}
// @Router /api/sftp/ls [get]
func (ap *app) apiSftpLs(c *gin.Context) {
	token, isBreak := checkToken(c)
	if isBreak {
		return
	}
	dirPath := c.Query("dir")

	client, err := NewSftpClient(token, ap.AddrRpc)
	if checkErr(c, err) {
		return
	}
	defer client.Close()
	items, err := client.ReadDir(dirPath)
	if checkErr(c, err) {
		return
	}
	returnData(c, items)
}

// apiSftpRm delete
// @Summary remove a file or a directory
// @Description delete a file or a directory by sftp
// @Tags SFTP
// @Accept  json
// @Produce  json
// @Param token query string true "TOKEN"
// @Param dir_or_file query string true "is a file or dir" Enums(file,dir)
// @Success 200 {object} responseObj
// @Router /api/sftp/rm [get]
func (ap *app) apiSftpRm(c *gin.Context) {

	token, isBreak := checkToken(c)
	if isBreak {
		return
	}
	dirPath := c.Query("path")
	dirOrFile := c.Query("dir_or_file")

	client, err := NewSftpClient(token, ap.AddrRpc)
	if checkErr(c, err) {
		return
	}
	defer client.Close()
	err = client.Remove(dirPath, dirOrFile)
	if checkErr(c, err) {
		return
	}
	returnOk(c)
}

// apiSftpUpload upload
// @Summary upload an item
// @Description upload an item by SFTP
// @Tags SFTP
// @Accept  multipart/form-data
// @Produce  json
// @Param token formData string true "TOKEN"
// @Param dir formData string true "the destination dir path"
// @Param file formData file true "the file will be uploaded"
// @Success 200 {object} responseObj
// @Router /api/sftp/upload [post]
func (ap *app) apiSftpUpload(c *gin.Context) {
	token, isBreak := checkToken(c)
	if isBreak {
		return
	}
	desDir := c.Query("dir")
	formFile, err := c.FormFile("file")
	if checkErr(c, err) {
		return
	}

	srcFile, err := formFile.Open()
	if checkErr(c, err) {
		return
	}
	defer srcFile.Close()

	client, err := NewSftpClient(token, ap.AddrRpc)
	if checkErr(c, err) {
		return
	}
	defer client.Close()

	err = client.Upload(desDir, formFile.Filename, srcFile)
	if checkErr(c, err) {
		return
	}
	returnOk(c)
}

// apiSftpDownloadFile download
// @Summary download a file
// @Description download a file
// @Tags SFTP
// @Accept  json
// @Produce  octet-stream
// @Param token query string true "TOKEN"
// @Param path query string true "a file's path"
// @Success 200 {object} responseObj
// @Router /api/sftp/download/file [get]
func (ap *app) apiSftpDownloadFile(c *gin.Context) {
	token, isBreak := checkToken(c)
	if isBreak {
		return
	}
	path := c.Query("path")

	client, err := NewSftpClient(token, ap.AddrRpc)
	if checkErr(c, err) {
		return
	}
	defer client.Close()

	bs, err := client.DownloadFile(path)
	if checkErr(c, err) {
		return
	}
	fn := filepath.Base(path)
	returnFile(c, bs, "octet-stream", fmt.Sprintf(`attachment; filename="%s"`, fn))

}

// apiSftpDownloadDir download a dir
// @Summary download a dir
// @Description download a dir then return a zip file
// @Tags SFTP
// @Accept  json
// @Produce  octet-stream
// @Param token query string true "TOKEN"
// @Param path query string true "the dir to be downloaded"
// @Success 200 {object} responseObj
// @Router /api/sftp/download/dir [get]
func (ap *app) apiSftpDownloadDir(c *gin.Context) {
	token, isBreak := checkToken(c)
	if isBreak {
		return
	}
	path := c.Query("path")

	client, err := NewSftpClient(token, ap.AddrRpc)
	if checkErr(c, err) {
		return
	}
	defer client.Close()

	bs, err := client.DownloadDir(path)
	if checkErr(c, err) {
		return
	}
	fn := strings.ReplaceAll(path, "/", "_") + ".zip"
	returnFile(c, bs, "application/zip", fmt.Sprintf(`attachment; filename="%s"`, fn))
}

// apiSftpMkdir make a dir
// @Summary make a dir
// @Description make a dir
// @Tags SFTP
// @Accept  json
// @Produce  octet-stream
// @Param token query string true "TOKEN"
// @Param path query string true "the path of dir which will be made"
// @Success 200 {object} responseObj
// @Router /api/sftp/mkdir [get]
func (ap *app) apiSftpMkdir(c *gin.Context) {
	token, isBreak := checkToken(c)
	if isBreak {
		return
	}
	path := c.Query("path")

	client, err := NewSftpClient(token, ap.AddrRpc)
	if checkErr(c, err) {
		return
	}
	defer client.Close()
	err = client.Mkdir(path)
	if checkErr(c, err) {
		return
	}
	returnOk(c)
}
