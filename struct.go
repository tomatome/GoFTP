// struct.go
package main

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type MyPage struct {
	local  location
	remote location
	page   TabPage
}

type location struct {
	Tv    *walk.TableView
	Tl    *walk.LineEdit
	Model *FileModel
}

func (l location) Refresh() {
	l.Model.SetDirPath(l.Tl.Text())
}

func (l location) Hidden() {
	l.SetHidden(true)
	l.Refresh()
}

func (l location) Show() {
	l.SetHidden(false)
	l.Refresh()
}

func (l location) SetHidden(hidden bool) {
	l.Model.hidden = hidden
}

func (p MyPage) Send() error {
	fmt.Println("Link:", p.remote.Model.remote.IP)

	p.remote.Model.remote.Link()
	info := p.local.Model.items[p.local.Tv.CurrentIndex()]
	var localPath = filepath.Join(p.local.Tl.Text(), info.Name)
	var remotePath = p.remote.Tl.Text()

	return p.remote.Model.remote.Upload(localPath, remotePath)
}

func (p MyPage) Recv() error {
	fmt.Println("Link:", p.remote.Model.remote.IP)

	p.remote.Model.remote.Link()

	info := p.remote.Model.items[p.remote.Tv.CurrentIndex()]
	var localPath = p.local.Tl.Text()
	var remotePath = p.remote.Tl.Text() + "/" + info.Name

	p.remote.Model.remote.Download(remotePath, localPath)

	fmt.Println("recv file from remote server finished!")
	return nil
}

func formatSize(size int64) string {
	if size < 1024 {
		return strconv.FormatInt(size, 10) + " B"
	}

	s := float64(size)
	d := "B"

	if s >= 1024 {
		s = s / 1024.0
		d = "K"
	}

	if s >= 1024 {
		s = s / 1024.0
		d = "M"
	}
	if s >= 1024 {
		s = s / 1024.0
		d = "G"
	}

	return fmt.Sprintf("%.1f %s", s, d)
}
