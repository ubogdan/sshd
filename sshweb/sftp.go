package sshweb

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/bytegang/pb"
	"github.com/bytegang/sshd/pkg/util"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"google.golang.org/grpc"
	"io"
	"io/ioutil"
	"mime/multipart"
	"path"
	"path/filepath"
	"time"
)

type sftpCli struct {
	account   *pb.ResSshConnCfg
	sshClient *ssh.Client
	client    *sftp.Client
	//todo:: 增加审批action记录器
}

func NewSftpClient(token, rpcAddr string) (*sftpCli, error) {
	var opts = []grpc.DialOption{grpc.WithInsecure()}
	conn, err := grpc.Dial(rpcAddr, opts...)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	acc, err := pb.NewByteGangsterClient(conn).WebXtermSsh(context.Background(), &pb.ReqToken{Token: token})
	if err != nil {
		return nil, err
	}

	sshClient, err := util.GetSshClient(acc.AssetConn, acc.ProxyConn, time.Second*30)
	if err != nil {
		return nil, err
	}
	client, err := sftp.NewClient(sshClient, sftp.MaxPacket(32<<10))
	if err != nil {
		return nil, err
	}
	return &sftpCli{account: acc, client: client, sshClient: sshClient}, nil
}

type DirItem struct {
	Name  string    `json:"name"`
	Path  string    `json:"path"` // including Name
	Size  int64     `json:"size"`
	Time  time.Time `json:"time"`
	Mod   string    `json:"mod"`
	IsDir bool      `json:"is_dir"`
}

type ResReadDir struct {
	List []DirItem `json:"list"`
	Dir  string    `json:"dir"`
}

func (c *sftpCli) Close() {
	if c.client != nil {
		c.client.Close()
	}
	if c.sshClient != nil {
		c.sshClient.Close()
	}
}

func (c *sftpCli) ReadDir(dirPath string) (*ResReadDir, error) {
	if dirPath == "" {
		dir, err := c.homeDir()
		if err != nil {
			return nil, err
		}
		dirPath = dir
	}

	files, err := c.client.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}
	fileList := make([]DirItem, 0) // this will not be converted to null if slice is empty.
	for _, file := range files {
		tt := DirItem{
			Name:  file.Name(),
			Size:  file.Size(),
			Path:  path.Join(dirPath, file.Name()),
			Time:  file.ModTime(),
			Mod:   file.Mode().String(),
			IsDir: file.IsDir(),
		}
		fileList = append(fileList, tt)
	}
	return &ResReadDir{
		List: fileList,
		Dir:  dirPath,
	}, nil
}

func (c *sftpCli) Mkdir(dirPath string) error {
	err := c.client.MkdirAll(dirPath)
	if err != nil {
		return err
	}
	return nil
}

func (c *sftpCli) Rename(o, n string) error {
	err := c.client.Rename(o, n)
	if err != nil {
		return err
	}
	return nil
}

func (c *sftpCli) DownloadDir(fullPath string) ([]byte, error) {
	sftpClient := c.client
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	err := zipAddFiles(w, sftpClient, fullPath, "/")
	if err != nil {
		return nil, err
	}
	// Make sure to check the error on Close.
	err = w.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
	//dName := time.Now().Format("2006_01_02T15_04_05Z07.zip")
	//extraHeaders := map[string]string{
	//	"Content-Disposition": fmt.Sprintf(`attachment; filename="%s"`, dName),
	//}
	//c.DataFromReader(http.StatusOK, int64(buf.Len()), "application/zip", buf, extraHeaders)

}

func (c *sftpCli) DownloadFile(fullPath string) ([]byte, error) {
	f, err := c.client.Open(fullPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	bs, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return bs, nil
}

func (c *sftpCli) homeDir() (string, error) {
	return c.client.Getwd()
}
func (c *sftpCli) Upload(desDir string, fileName string, srcFile multipart.File) error {
	if desDir == "$HOME" {
		wd, err := c.homeDir()
		if err != nil {
			return err
		}
		desDir = wd
	}

	dstFile, err := c.client.Create(path.Join(desDir, fileName))
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = dstFile.ReadFrom(srcFile)
	if err != nil {
		return err
	}
	return nil
}

func (c *sftpCli) Remove(fullPath, dirOrFile string) error {
	if fullPath == "/" || fullPath == "$HOME" {
		return errors.New("can't delete / or $HOME dir")
	}
	switch dirOrFile {
	case "dir":
		return removeNonemptyDirectory(c.client, fullPath)
	case "file":
		return c.client.Remove(fullPath)
	default:
		return errors.New("dirOrFile 参数是必须的且是file/dir")
	}
}

// removeNonemptyDirectory removes the non empty directory in sftp server.
// sftp protocol does not allows removing non empty directory.
// we need to traverse over the file tree to remove files and directories post-orderly
func removeNonemptyDirectory(c *sftp.Client, path string) error {
	list, err := c.ReadDir(path)
	if err != nil {
		return err
	}
	// walk over the tree
	for i, cur := range list {
		newPath := filepath.Join(path, list[i].Name())
		if cur.IsDir() {
			if err := removeNonemptyDirectory(c, newPath); err != nil {
				return err
			}
		} else {
			if err := c.Remove(newPath); err != nil {
				return err
			}
		}
	}
	// remove current directory, which now is empty
	return c.RemoveDirectory(path)
}

func zipAddFiles(w *zip.Writer, sftpC *sftp.Client, basePath, baseInZip string) error {
	// Open the Directory
	files, err := sftpC.ReadDir(basePath)
	if err != nil {
		return fmt.Errorf("sftp 读取目录 %s 失败:%s", basePath, err)
	}

	for _, file := range files {
		thisFilePath := basePath + "/" + file.Name()
		if file.IsDir() {

			err := zipAddFiles(w, sftpC, thisFilePath, baseInZip+file.Name()+"/")
			if err != nil {
				return fmt.Errorf("递归目录%s 失败:%s", thisFilePath, err)
			}
		} else {

			dat, err := sftpC.Open(thisFilePath)
			if err != nil {
				return fmt.Errorf("sftp 读取文件失败 %s:%s", thisFilePath, err)
			}
			// Add some files to the archive.
			zipElePath := baseInZip + file.Name()
			f, err := w.Create(zipElePath)
			if err != nil {
				return fmt.Errorf("写入zip writer header失败 %s:%s", zipElePath, err)
			}
			b, err := ioutil.ReadAll(dat)
			if err != nil {
				return fmt.Errorf("ioutil read all failed %s", err)
			}
			_, err = f.Write(b)
			if err != nil {
				return fmt.Errorf("写入zip writer 内容 bytes失败:%s", err)
			}
		}
	}
	return nil
}
