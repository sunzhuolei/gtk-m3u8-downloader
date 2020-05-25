package main

import (
	//"fmt"
	//"testing"
	//"time"
	//"time"
	//"time"
	"github.com/mattn/go-gtk/gtk"
	"os"
	//"strconv"
	"github.com/mattn/go-gtk/glib"
	//"fmt"
	"strconv"
	"time"
	"github.com/mattn/go-gtk/gdk"
)
func ddd(){
	gtk.Init(&os.Args)
	//gtk.MainIterationDo(true)
	win := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
	win.SetTitle("test")
	win.SetSizeRequest(600,400)
	fix := gtk.NewFixed()
	label:= gtk.NewLabel("sddssdds")
	btn := gtk.NewButtonWithLabel("点我")
	fix.Put(label,100,60)
	fix.Put(btn,100,100)
	gdk.ThreadsInit()
	btn.Connect("clicked", func(context *glib.CallbackContext) {
		for i:=1;i<=100;i++{
			gdk.ThreadsEnter()
			label.SetText("文字"+strconv.Itoa(i))
			time.Sleep(300*time.Millisecond)
			gdk.ThreadsLeave()
		}
	})


	fix.Add(label)
	win.Add(fix)
	win.ShowAll()
	gtk.Main()
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})
}