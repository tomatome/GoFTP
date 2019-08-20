package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/lxn/walk"
	"github.com/pkg/sftp"
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

type FileInfo struct {
	Name     string
	Size     int64
	Modified time.Time
	Dir      bool
}

type FileModel struct {
	walk.SortedReflectTableModelBase
	dirPath string
	remote  *Client
	items   []*FileInfo
}

var _ walk.ReflectTableModel = new(FileModel)

func NewFileModel(remote *Client) *FileModel {
	m := new(FileModel)
	m.remote = remote
	m.items = make([]*FileInfo, 0, 30)
	return m
}

func (m *FileModel) Items() interface{} {
	return m.items
}

func (m *FileModel) Len() int {
	return len(m.items)
}

func (m *FileModel) SetDirPath(dirPath string) error {
	var w []os.FileInfo
	var err error
	if m.remote != nil {
		if m.remote.client == nil {
			sftpClient, err := connect(m.remote.Username, m.remote.Password, m.remote.IP, m.remote.Port)
			if err != nil {
				log.Fatal("Connect:", err)
			}
			m.remote.client = sftpClient
		}
		w, err = m.remote.client.ReadDir(dirPath)
	} else {
		w, err = ioutil.ReadDir(dirPath)
	}
	if err != nil {
		fmt.Println(dirPath, ":", err)
		return err
	}

	m.dirPath = dirPath
	m.items = m.items[0:0]
	item := &FileInfo{
		Name:     "..",
		Size:     0,
		Modified: time.Now(),
		Dir:      true,
	}

	m.items = append(m.items, item)
	for _, info := range w {
		name := info.Name()
		if shouldExclude(name) {
			continue
		}

		item := &FileInfo{
			Name:     name,
			Size:     info.Size(),
			Modified: info.ModTime(),
			Dir:      info.IsDir(),
		}

		m.items = append(m.items, item)
	}

	m.PublishRowsReset()

	return nil
}
func (m *FileModel) Image(row int) interface{} {
	if m.items[row].Dir {
		return "images/dir.ico"
	}
	return "images/file.ico"
}
func shouldExclude(name string) bool {
	switch name {
	case "System Volume Information", "pagefile.sys", "swapfile.sys":
		return true
	}

	return false
}
