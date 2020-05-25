package main

import (
	"github.com/mattn/go-gtk/gdk"
	"github.com/mattn/go-gtk/gdkpixbuf"
	"github.com/mattn/go-gtk/glib"
	"os"
	"strings"
)


/**
绘制图片到窗口上
 */
func DrawWindowFromFile(context *glib.CallbackContext){
	data := context.Data()
	obj,ok := data.(*SpiderWindow)
	if ok == false{
		return
	}
	//设置画板
	paint := obj.Win.GetWindow().GetDrawable()
	gc := gdk.NewGC(paint)
	GOPATH := os.Getenv("GOPATH")
	GOPATH = strings.ReplaceAll(GOPATH,"\\","/")
	pixbuf,_ := gdkpixbuf.NewPixbufFromFileAtScale(GOPATH+"/src/images/mmbg.jpg",640,400,false)
	paint.DrawPixbuf(gc,pixbuf,0,0,0,0,-1,-1,gdk.RGB_DITHER_NONE,0,0)
	pixbuf.Unref()
}


/**
处理下载
 */
func HandleDownLoad(context *glib.CallbackContext){
	data := context.Data()
	obj,ok := data.(*SpiderWindow)
	if ok== false{
		return
	}
	url := obj.Input.GetText()
	download(url,obj)
}

