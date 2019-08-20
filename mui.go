package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

const (
	Title = "客户端"
)

type MyMainWindow struct {
	main    *walk.MainWindow
	curUser *user.User
	tab     *walk.TabWidget
	sbi     *walk.StatusBarItem
	pages   []*MyPage
}

func initWindows() *MyMainWindow {
	mw := new(MyMainWindow)
	mw.curUser, _ = user.Current()
	mw.pages = make([]*MyPage, 0, 10)

	return mw
}

var mw *MyMainWindow

func main() {
	mw = initWindows()
	MWindow := MainWindow{
		AssignTo: &mw.main,
		Title:    Title,
		Icon:     "images/client.ico",
		MinSize:  Size{900, 600},
		//Size:      Size{800, 600},
		MenuItems: mw.initMenus(),
		ToolBar: ToolBar{
			ButtonStyle: ToolBarButtonImageBeforeText,
			Items: []MenuItem{
				Action{
					Text:  "上传",
					Image: "images/upload.ico",
					OnTriggered: func() {
						p := mw.pages[mw.tab.CurrentIndex()]
						now := time.Now()
						err := p.Send()
						if err != nil {
							mw.sbi.SetText(err.Error())
						} else {
							mw.sbi.SetText("Send successfully:" + time.Now().Sub(now).String())
						}
						p.remote.Model.SetDirPath(p.remote.Tl.Text())
					},
				},
				Action{
					Text:  "下载",
					Image: "images/down.ico",
					OnTriggered: func() {
						p := mw.pages[mw.tab.CurrentIndex()]
						now := time.Now()
						err := p.Recv()
						if err != nil {
							mw.sbi.SetText(err.Error())
						} else {
							mw.sbi.SetText("Recv successfully:" + time.Now().Sub(now).String())
						}
						p.local.Model.SetDirPath(p.local.Tl.Text())
					},
				},
			},
		},
		Layout: HBox{MarginsZero: true},
		Children: []Widget{
			mw.initTabWidget(),
		},
		StatusBarItems: []StatusBarItem{
			StatusBarItem{
				AssignTo: &mw.sbi,
				Text:     "正在加载...",
			},
		},
	}

	if _, err := MWindow.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func initMenu(text string, c walk.EventHandler, items []MenuItem) Menu {
	var m Menu
	m.Text = text
	m.OnTriggered = c
	m.Items = items
	return m
}

func initAction(text string, c walk.EventHandler, key walk.Key, img string) Action {
	var act Action
	act.Text = text
	act.OnTriggered = c
	act.Shortcut = Shortcut{walk.ModControl, key}
	if img != "" {
		act.Image = "images/" + img
	}

	return act
}

func (mw *MyMainWindow) initMenus() []MenuItem {
	items := make([]MenuItem, 0, 1)
	a := initAction("新建会话", mw.NewSession, walk.KeyO, "session.ico")
	items = append(items, a)
	m := initMenu("会话", nil, items)
	return []MenuItem{m}

}

func (mw *MyMainWindow) NewSession() {
	var wp *walk.TabPage
	tp := initTabPage("hello", "world")
	tp.page.AssignTo = &wp
	tp.page.Create(NewBuilder(mw.tab.Parent()))
	mw.tab.Pages().Add(*tp.page.AssignTo)
	mw.tab.CurrentIndexChanged()
}

func (mw *MyMainWindow) RmSession() {
	mw.tab.Pages().RemoveAt(1)
	mw.tab.CurrentIndexChanged()
}

func TableViewColumns() []TableViewColumn {
	return []TableViewColumn{
		TableViewColumn{
			DataMember: "Name",
			Width:      192,
		},
		TableViewColumn{
			DataMember: "Size",
			FormatFunc: func(value interface{}) string {
				return formatSize(value.(int64))
			},
			Alignment: AlignFar,
			Width:     64,
		},
		TableViewColumn{
			DataMember: "Modified",
			Format:     "2006-01-02 15:04:05",
			Width:      120,
		},
	}
}

func initTabPage(title, a string) *MyPage {
	p := new(MyPage)

	p.local.Model = NewFileModel(nil)
	p.local.Model.SetDirPath("d:\\")
	p.remote.Model = NewFileModel(newClient())
	p.remote.Model.SetDirPath("/")

	p.page.Title = title
	p.page.Image = "images/session.ico"
	p.page.Layout = VBox{}
	p.page.Children = []Widget{
		HSplitter{
			Children: []Widget{
				Composite{
					Layout: VBox{},
					Children: []Widget{
						LineEdit{AssignTo: &(p.local.Tl), Text: "d:\\", TextAlignment: AlignNear, OnKeyDown: func(key walk.Key) {
							if key == walk.KeyReturn {
								p.local.Model.SetDirPath(p.local.Tl.Text())
								p.local.Tl.SetText(p.local.Tl.Text())
							}
						}},
						TableView{
							AssignTo: &(p.local.Tv),
							Columns:  TableViewColumns(),
							Model:    p.local.Model,
							OnItemActivated: func() {
								idx := p.local.Tv.CurrentIndex()
								if idx < 0 {
									return
								}
								fs := filepath.Join(p.local.Model.dirPath, p.local.Model.items[idx].Name)
								p.local.Model.SetDirPath(fs)
								p.local.Tl.SetText(fs)

							},
							OnCurrentIndexChanged: func() {
								idx := p.local.Tv.CurrentIndex()
								if idx < 0 {
									return
								}
								fs := filepath.Join(p.local.Model.dirPath, p.local.Model.items[idx].Name)
								mw.sbi.SetText(fs)
							},
						},
					},
				},

				Composite{
					Layout: VBox{},
					Children: []Widget{
						LineEdit{AssignTo: &p.remote.Tl, Text: "/", TextAlignment: AlignNear, OnKeyDown: func(key walk.Key) {
							if key == walk.KeyReturn {
								p.remote.Model.SetDirPath(p.remote.Tl.Text())
								p.remote.Tl.SetText(p.remote.Tl.Text())
							}
						}},
						TableView{
							AssignTo: &p.remote.Tv,
							Columns:  TableViewColumns(),
							Model:    p.remote.Model,
							OnItemActivated: func() {
								idx := p.remote.Tv.CurrentIndex()
								fs := p.remote.Model.dirPath
								if p.remote.Model.items[idx].Name == ".." {
									if p.remote.Model.dirPath == "/" {
										return
									} else {
										idx := strings.LastIndex(p.remote.Model.dirPath, "/")
										if idx == -1 {
											return
										}
										if idx == 0 {
											idx = 1
										}
										fs = p.remote.Model.dirPath[:idx]
									}

								} else {
									if p.remote.Model.dirPath == "/" {
										fs = p.remote.Model.dirPath + p.remote.Model.items[idx].Name
									} else {
										fs = strings.Join([]string{p.remote.Model.dirPath, p.remote.Model.items[idx].Name}, "/")
									}
								}
								p.remote.Model.SetDirPath(fs)
								p.remote.Tl.SetText(fs)
							},
						},
					},
				},
			},
		},
	}

	return p
}
func (mw *MyMainWindow) initTabWidget() TabWidget {
	pages := make([]TabPage, 0, 2)
	p := initTabPage("新建会话", "test")
	mw.pages = append(mw.pages, p)
	pages = append(pages, p.page)
	return TabWidget{
		AssignTo: &mw.tab,
		Pages:    pages,
	}
}
