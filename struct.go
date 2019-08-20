// struct.go
package main

import (
	"fmt"
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

func (p MyPage) Send() error {
	var (
		err        error
		sftpClient *sftp.Client
	)

	// 这里换成实际的 SSH 连接的 用户名，密码，主机名或IP，SSH端口
	sftpClient, err = connect("root", "jhadmin", "192.168.0.76", 22)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer sftpClient.Close()

	info := p.local.Model.items[p.local.Tv.CurrentIndex()]
	var localFilePath = filepath.Join(p.local.Tl.Text(), info.Name)
	var remoteDir = p.remote.Tl.Text()
	srcFile, err := os.Open(localFilePath)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer srcFile.Close()

	var remoteFileName = filepath.Base(localFilePath)
	fmt.Println(localFilePath, remoteFileName)
	dstFile, err := sftpClient.Create(remoteDir + "/" + remoteFileName)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer dstFile.Close()

	buf := make([]byte, 1024)
	for {
		n, _ := srcFile.Read(buf)
		if n == 0 {
			break
		}
		dstFile.Write(buf)
	}

	fmt.Println("copy file to remote server finished!")
	return nil
}

func SendFile(local, remote string) {

}

func (p MyPage) Recv() error {

	var (
		err        error
		sftpClient *sftp.Client
	)

	// 这里换成实际的 SSH 连接的 用户名，密码，主机名或IP，SSH端口
	sftpClient, err = connect("root", "jhadmin", "192.168.0.76", 22)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer sftpClient.Close()

	info := p.remote.Model.items[p.remote.Tv.CurrentIndex()]
	var remoteFilePath = p.remote.Tl.Text() + "/" + info.Name
	var localDir = p.local.Tl.Text()

	srcFile, err := sftpClient.Open(remoteFilePath)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer srcFile.Close()

	var localFileName = info.Name
	dstFile, err := os.Create(path.Join(localDir, localFileName))
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer dstFile.Close()

	if _, err = srcFile.WriteTo(dstFile); err != nil {
		log.Fatal(err)
		return err
	}

	fmt.Println("copy file from remote server finished!")
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
