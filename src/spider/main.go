package main

import (
	"github.com/mattn/go-gtk/gtk"
	"os"
	//"github.com/mattn/go-gtk/gdk"
	//"fmt"
	"strings"
)


/**
窗口对象
 */
type SpiderWindow struct {
	Win       *gtk.Window
	Input     *gtk.Entry
	MsgLabel  *gtk.Label
	Progress  *gtk.ProgressBar
	CommitBtn *gtk.Button
}


/**
初始化窗口
 */
func (obj *SpiderWindow) init() {
	builder := gtk.NewBuilder()
	GOPATH := os.Getenv("GOPATH")
	GOPATH = strings.ReplaceAll(GOPATH,"\\","/")
	builder.AddFromFile(GOPATH +"/src/spider/window2.glade")
	obj.Win = gtk.WindowFromObject(builder.GetObject("window1"))
	obj.Win.SetSizeRequest(640, 400)
	obj.Win.SetTitle("m3u8视频下载器")
	obj.Win.SetResizable(false)
	obj.Win.SetAppPaintable(true)
	obj.Win.Connect("expose-event", DrawWindowFromFile, obj)
	obj.Input = gtk.EntryFromObject(builder.GetObject("entry1"))
	obj.MsgLabel = gtk.LabelFromObject(builder.GetObject("label5"))
	obj.Progress = gtk.ProgressBarFromObject(builder.GetObject("progressbar1"))
	obj.CommitBtn = gtk.ButtonFromObject(builder.GetObject("button1"))
	obj.CommitBtn.Connect("clicked",HandleDownLoad,obj)
}


/**
主入口
 */
func main() {
	gtk.Init(&os.Args)
	var spiderObj SpiderWindow
	spiderObj.init()
	spiderObj.Win.ShowAll()
	spiderObj.Win.Connect("destroy", func() {
		gtk.MainQuit()
	})
	spiderObj.Win.ShowAll()
	gtk.Main()
	return
}
