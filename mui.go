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
	main      *walk.MainWindow
	curUser   *user.User
	tab       *walk.TabWidget
	sbi       *walk.StatusBarItem
	pages     []*MyPage
	hidden    bool
	hlb       *walk.ListBox
	lb        *walk.ListBox
	nodeModel *NodeModel
}

func initWindows() *MyMainWindow {
	mw := new(MyMainWindow)
	mw.curUser, _ = user.Current()
	mw.pages = make([]*MyPage, 0, 10)
	mw.hidden = true
	mw.nodeModel = newNodeModel()

	return mw
}

var mw *MyMainWindow

func main() {
	mw = initWindows()
	MWindow := MainWindow{
		AssignTo:  &mw.main,
		Title:     Title,
		Icon:      "images/client.ico",
		MinSize:   Size{900, 600},
		MenuItems: mw.initMenus(),
		ToolBar: ToolBar{
			ButtonStyle: ToolBarButtonImageBeforeText,
			Items: []MenuItem{
				Action{
					Text:  "会话管理",
					Image: "images/list.ico",
					OnTriggered: func() {
						if mw.hlb.Visible() {
							mw.hlb.SetVisible(false)
						} else {
							mw.hlb.SetVisible(true)
						}
					},
				},
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
				Action{
					Text:  "关闭会话",
					Image: "images/close.ico",
					OnTriggered: func() {
						mw.tab.Pages().RemoveAt(mw.tab.CurrentIndex())
					},
				},
			},
		},
		Layout: HBox{SpacingZero: true},
		Children: []Widget{
			HSplitter{
				Children: []Widget{
					ListBox{
						AssignTo: &mw.hlb,
						Model:    mw.nodeModel,
						Visible:  false,
						OnItemActivated: func() {
							i := mw.hlb.CurrentIndex()
							node := mw.nodeModel.Node(i)
							fmt.Println(node)
							mw.NewSession(node)
						},
					},

					mw.initTabWidget(),
				},
			},
		},
		StatusBarItems: []StatusBarItem{
			StatusBarItem{
				AssignTo: &mw.sbi,
				Text:     "加载完成",
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
	a := initAction("新建会话", mw.RunNewDialog, walk.KeyO, "session.ico")
	items = append(items, a)
	m := initMenu("会话", nil, items)

	items1 := make([]MenuItem, 0, 1)
	a1 := initAction("隐藏文件", func() {
		hidden := false
		if mw.hidden { //strings.EqualFold(a1.Text, "显示隐藏文件") {
			//a1.Text = "关闭隐藏文件"
			hidden = false
		} else {
			//a1.Text = "显示隐藏文件"
			hidden = true
		}
		mw.hidden = hidden
		for _, p := range mw.pages {
			p.local.Model.hidden = hidden
			p.local.Model.SetDirPath(p.local.Tl.Text())
			p.remote.Model.hidden = hidden
			p.remote.Model.SetDirPath(p.remote.Tl.Text())
		}
	}, walk.KeyO, "session.ico")
	items1 = append(items1, a1)
	m1 := initMenu("设置", nil, items1)
	return []MenuItem{m, m1}
}
func (mw *MyMainWindow) RunNewDialog() {
	var dlg *walk.Dialog
	var ip, user, passwd *walk.LineEdit
	var port *walk.NumberEdit
	Dialog{
		AssignTo: &dlg,
		Title:    "会话列表",
		MinSize:  Size{500, 400},
		Layout:   VBox{},
		Children: []Widget{
			HSplitter{
				Children: []Widget{
					Composite{
						Layout: VBox{},
						Children: []Widget{
							ListBox{
								AssignTo: &mw.lb,
								Model:    mw.nodeModel,
								OnCurrentIndexChanged: func() {
									i := mw.lb.CurrentIndex()
									node := mw.nodeModel.Node(i)
									ip.SetText(node.IP)
									port.SetValue(float64(node.Port))
									user.SetText(node.Username)
									passwd.SetText(node.Password)
								},
								OnItemActivated: func() {
									i := mw.lb.CurrentIndex()
									node := mw.nodeModel.Node(i)
									fmt.Printf("OnItemActivated:%+v\n", node)
									dlg.Close(0)
									mw.NewSession(node)
								},
							},
							Composite{
								Layout: HBox{},
								Children: []Widget{
									PushButton{
										Text: "新建会话",
										OnClicked: func() {

										},
									},
									PushButton{
										Text: "删除会话",
										OnClicked: func() {

										},
									},
								},
							},
						},
					},
					GroupBox{
						Layout:        Grid{Columns: 2},
						StretchFactor: 2,
						Children: []Widget{
							Label{Text: "主机名:"},
							Label{Text: "端口:"},
							LineEdit{AssignTo: &ip},
							NumberEdit{AssignTo: &port, StretchFactor: 4, Value: 22},
							Label{Text: "用户名:"},
							Label{Text: "密码:"},
							LineEdit{AssignTo: &user},
							LineEdit{AssignTo: &passwd},
							PushButton{
								Text: "保存",
								OnClicked: func() {
									client := newClient()
									client.IP = ip.Text()
									client.Port = int(port.Value())
									client.Username = user.Text()
									client.Password = passwd.Text()
									mw.nodeModel.Add(client)
									mw.nodeModel.WriteSession(client)
									ip.SetText("")
									port.SetValue(22)
									user.SetText("")
									passwd.SetText("")
								},
							},
							PushButton{
								Text: "清空",
								OnClicked: func() {

								},
							},
						},
					},
				},
			},
		},
	}.Run(mw.main)
}

func (mw *MyMainWindow) NewSession(c *Client) {
	var wp *walk.TabPage
	tp := initTabPage(c)
	tp.page.AssignTo = &wp
	tp.page.Create(NewBuilder(mw.tab.Parent()))
	mw.tab.Pages().Add(wp)
	mw.tab.SetCurrentIndex(mw.tab.Pages().Len() - 1)
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

func initTabPage(c *Client) *MyPage {
	p := new(MyPage)

	p.local.Model = NewFileModel(nil)
	p.local.Model.SetDirPath("D:\\")
	p.remote.Model = NewFileModel(c)
	p.remote.Model.SetDirPath("/")

	p.page.Title = c.Title()
	p.page.Image = "images/session.ico"
	p.page.Layout = VBox{}
	p.page.Children = []Widget{
		HSplitter{
			Children: []Widget{
				Composite{
					Layout: VBox{},
					Children: []Widget{
						LineEdit{AssignTo: &(p.local.Tl), Text: "D:\\", TextAlignment: AlignNear, OnKeyDown: func(key walk.Key) {
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
	c := newClient()
	p := initTabPage(c)
	mw.pages = append(mw.pages, p)
	pages = append(pages, p.page)
	return TabWidget{
		AssignTo:      &mw.tab,
		Pages:         pages,
		StretchFactor: 5,
	}
}
