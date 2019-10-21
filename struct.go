// struct.go
package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/pkg/sftp"
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
	var (
		err        error
		sftpClient *sftp.Client
	)

	fmt.Println("Link:", p.remote.Model.remote.IP)

	sftpClient = p.remote.Model.remote.Link()
	info := p.local.Model.items[p.local.Tv.CurrentIndex()]
	var localFilePath = filepath.Join(p.local.Tl.Text(), info.Name)
	var remoteFilePath = p.remote.Tl.Text() + "/" + info.Name

	if info.Dir {
		sftpClient.Mkdir(remoteFilePath)
	}

	srcFile, err := os.Open(localFilePath)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer srcFile.Close()

	dstFile, err := sftpClient.Create(remoteFilePath)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer dstFile.Close()

	if _, err = io.Copy(dstFile, srcFile); err != nil {
		log.Fatal(err)
		return err
	}

	fmt.Println("send file to remote server finished!")
	return nil
}

func (p MyPage) Recv() error {
	var (
		err        error
		sftpClient *sftp.Client
	)

	fmt.Println("Link:", p.remote.Model.remote.IP)

	sftpClient = p.remote.Model.remote.Link()

	info := p.remote.Model.items[p.remote.Tv.CurrentIndex()]
	var localFilePath = path.Join(p.local.Tl.Text(), info.Name)
	var remoteFilePath = p.remote.Tl.Text() + "/" + info.Name

	srcFile, err := sftpClient.Open(remoteFilePath)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(localFilePath)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer dstFile.Close()

	if _, err = io.Copy(dstFile, srcFile); err != nil {
		log.Fatal(err)
		return err
	}

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
