package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/lxn/walk"
)

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
	hidden  bool
}

var _ walk.ReflectTableModel = new(FileModel)

func NewFileModel(remote *Client) *FileModel {
	m := new(FileModel)
	m.remote = remote
	m.items = make([]*FileInfo, 0, 30)
	m.hidden = true

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
		c := m.remote.Link()
		w, err = c.ReadDir(dirPath)
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
		if m.shouldExclude(name) {
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
func (m *FileModel) shouldExclude(name string) bool {
	switch name {
	case "System Volume Information", "pagefile.sys", "swapfile.sys":
		return true
	}

	if m.hidden &&
		(strings.HasPrefix(name, ".") || strings.HasPrefix(name, "$")) {
		return true
	}

	return false
}

type NodeModel struct {
	walk.ListModelBase
	nodes []*Client
}

func newNodeModel() *NodeModel {
	m := &NodeModel{nodes: make([]*Client, 0, 100)}
	m.ReadSession()

	return m
}

const SESSION_DATA = "sessions.json"

var CurDir = filepath.Dir(os.Args[0])

func (m *NodeModel) ReadSession() {
	file := path.Join(CurDir, SESSION_DATA)
	data, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Errorf("%v\n", err)
		return
	}

	content := string(data)
	for _, line := range strings.Split(content, "\n") {
		if len(line) <= 0 {
			continue
		}
		var c Client
		json.Unmarshal([]byte(line), &c)
		m.Add(&c, true)
	}
}

func (m *NodeModel) WriteSession(c *Client) {
	file := path.Join(CurDir, SESSION_DATA)
	s, e := json.Marshal(c)
	f, e := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if e != nil {
		fmt.Errorf("%v\n", e)
		return
	}
	f.Write(s)
	f.WriteString("\n")
	f.Close()
}
func (m *NodeModel) RemoveSession(c *Client) {
	file := path.Join(CurDir, SESSION_DATA)
	f, e := os.OpenFile(file, os.O_RDWR|os.O_TRUNC, 0666)
	if e != nil {
		fmt.Errorf("%v\n", e)
		return
	}
	for _, v := range m.nodes {
		s, _ := json.Marshal(v)
		f.Write(s)
		f.WriteString("\n")
	}
	f.Close()
}
func (m *NodeModel) ItemCount() int {
	return len(m.nodes)
}

func (m *NodeModel) Value(index int) interface{} {
	return m.nodes[index].Title()
}

func (m *NodeModel) Add(c *Client, first bool) {
	for _, v := range m.nodes {
		if v.IP == c.IP {
			return
		}
	}
	m.nodes = append(m.nodes, c)
	m.PublishItemsInserted(len(m.nodes)-1, len(m.nodes)-1)
	if first {
		return
	}
	m.WriteSession(c)
}
func (m *NodeModel) Remove(c *Client) {
	idx := -1
	for i, v := range m.nodes {
		if v.IP == c.IP {
			idx = i
		}
	}
	if idx < 0 {
		return
	}
	m.nodes = append(m.nodes[:idx], m.nodes[idx+1:]...)
	m.RemoveSession(c)
	m.PublishItemsRemoved(idx, idx)
}
func (m *NodeModel) Node(index int) *Client {
	return m.nodes[index]
}
