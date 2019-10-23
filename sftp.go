// ftpServer.go
package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type Client struct {
	IP       string //IP地址
	Username string //用户名
	InitDir  string
	Password string
	Port     int          //端口号
	client   *sftp.Client //ssh客户端
}

func newClient() *Client {
	return &Client{
		IP:       "192.168.0.76",
		Username: "root",
		Password: "jhadmin",
		Port:     22,
	}
}

func (c *Client) Title() string {
	return c.Username + "@" + c.IP
}
func (c *Client) isClose() bool {
	if c.client == nil {
		return true
	}

	if _, e := c.client.Getwd(); e != nil {
		c.client = nil
		return true
	}

	return false
}
func (c *Client) Link() *sftp.Client {
	if c.isClose() {
		client, err := connect(c.Username, c.Password, c.IP, c.Port)
		if err != nil {
			log.Fatal("Connect:", err)
		}
		c.client = client
	}

	return c.client
}
func (c *Client) IsDir(path string) bool {
	info, err := c.client.Stat(path)
	if err == nil && info.IsDir() {
		return true
	}
	return false
}
func (c *Client) IsFile(path string) bool {
	info, err := c.client.Stat(path)
	if err == nil && !info.IsDir() {
		return true
	}
	return false
}
func (c *Client) IsExist(path string) bool {
	_, err := c.client.Stat(path)
	return err == nil
}
func connect(user, password, host string, port int) (*sftp.Client, error) {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		sshClient    *ssh.Client
		client       *sftp.Client
		err          error
	)
	// get auth method
	auth = make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.Password(password))

	clientConfig = &ssh.ClientConfig{
		User:            user,
		Auth:            auth,
		Timeout:         60 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// connet to ssh
	addr = fmt.Sprintf("%s:%d", host, port)

	if sshClient, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}

	// create sftp client
	if client, err = sftp.NewClient(sshClient); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Client) Upload(local string, remote string) (err error) {
	info, err := os.Stat(local)
	if err != nil {
		return errors.New("上传(\"" + local + "\"):" + err.Error())
	}
	if info.IsDir() {
		return c.UploadDir(local, remote)
	}
	return c.UploadFile(local, remote)
}

func (c *Client) UploadFile(localFile, remote string) error {
	info, err := os.Stat(localFile)
	if err != nil || info.IsDir() {
		return errors.New("sftp: 本地文件不是文件 UploadFile(\"" + localFile + "\") 跳过上传")
	}

	l, err := os.Open(localFile)
	if err != nil {
		return errors.New("Upload Open " + localFile + ":" + err.Error())
	}
	defer l.Close()

	var remoteFile, remoteDir string
	if remote[len(remote)-1] == '/' {
		remoteFile = filepath.ToSlash(filepath.Join(remote, filepath.Base(localFile)))
		remoteDir = remote
	} else {
		remoteFile = remote
		remoteDir = filepath.ToSlash(filepath.Dir(remoteFile))
	}
	log.Println("Upload:", localFile, "-->", remoteFile)

	if _, err := c.client.Stat(remoteDir); err != nil {
		log.Println("Mkdir:", remoteDir)
		c.MkdirAll(remoteDir)
	}

	r, err := c.client.Create(remoteFile)
	if err != nil {
		c.client = nil
		c.Link()
		r, err = c.client.Create(remoteFile)
		if err != nil {
			return errors.New("Upload Create " + remoteFile + ":" + err.Error())
		}
	}

	_, err = io.Copy(r, l)

	return err
}

// UploadDir files without checking diff status
func (c *Client) UploadDir(localDir string, remoteDir string) (err error) {
	log.Println("sftp: UploadDir", localDir, "-->", remoteDir)

	rootLocal := filepath.Dir(localDir)
	if c.IsFile(remoteDir) {
		log.Println("sftp: Remove File:", remoteDir)
		c.client.Remove(remoteDir)
	}

	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println("err:", err)
			return err
		}

		relSrc, err := filepath.Rel(rootLocal, path)
		if err != nil {
			return err
		}
		finalDst := filepath.Join(remoteDir, relSrc)
		finalDst = filepath.ToSlash(finalDst)
		if info.IsDir() {
			if c.IsExist(finalDst) {
				return nil
			}
			err := c.MkdirAll(finalDst)
			if err != nil {
				log.Println("Mkdir failed:", err)
			}
		} else {
			return c.UploadFile(path, finalDst)
		}
		return nil

	}
	return filepath.Walk(localDir, walkFunc)
}
func (c *Client) MkdirAll(dirpath string) error {
	parentDir := filepath.ToSlash(filepath.Dir(dirpath))
	_, err := c.client.Stat(parentDir)
	if err != nil {
		if err.Error() == "file does not exist" {
			err := c.MkdirAll(parentDir)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	if c.isClose() {
		c.Link()
	}
	err = c.client.Mkdir(filepath.ToSlash(dirpath))
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) Download(remote string, local string) (err error) {
	if c.IsDir(remote) {
		return c.downloadDir(remote, local)

	}
	return c.downloadFile(remote, local)

}

// downloadFile a file from the remote server like cp
func (c *Client) downloadFile(remoteFile, local string) error {
	if !c.IsFile(remoteFile) {
		return errors.New("文件不存在或不是文件, 跳过目录下载 downloadFile(" + remoteFile + ")")
	}

	localFile := filepath.ToSlash(local)

	if err := os.MkdirAll(filepath.Dir(localFile), os.ModePerm); err != nil {
		log.Println(err)
		return err
	}

	r, err := c.client.Open(remoteFile)
	if err != nil {
		return err
	}
	defer r.Close()

	l, err := os.Create(localFile)
	if err != nil {
		return err
	}
	defer l.Close()

	_, err = io.Copy(l, r)
	return err
}

func (c *Client) downloadDir(remote, local string) error {
	var localDir, remoteDir string

	if !c.IsDir(remote) {
		return errors.New("目录不存在或不是目录, 跳过 downloadDir(" + remote + ")")
	}
	remoteDir = remote
	if remote[len(remote)-1] == '/' {
		localDir = local
	} else {
		localDir = path.Join(local, path.Base(remote))
	}

	walker := c.client.Walk(remoteDir)

	for walker.Step() {
		if err := walker.Err(); err != nil {
			log.Println(err)
			continue
		}

		info := walker.Stat()

		relPath, err := filepath.Rel(remoteDir, walker.Path())
		if err != nil {
			return err
		}

		localPath := filepath.ToSlash(filepath.Join(localDir, relPath))

		localInfo, err := os.Stat(localPath)
		if os.IsExist(err) {
			if localInfo.IsDir() {
				if info.IsDir() {
					continue
				}

				err = os.RemoveAll(localPath)
				if err != nil {
					return err
				}
			} else if info.IsDir() {
				err = os.Remove(localPath)
				if err != nil {
					return err
				}
			}
		}

		if info.IsDir() {
			err = os.MkdirAll(localPath, os.ModePerm)
			if err != nil {
				return err
			}

			continue
		}

		c.downloadFile(walker.Path(), localPath)

	}
	return nil
}
