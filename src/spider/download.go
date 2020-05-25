package main

import (
	"fmt"
	"crypto/aes"
	"crypto/cipher"
	"bytes"
	"strconv"
	"time"
	"os"
	"strings"
	"io"
	"net/http"
	"regexp"
	"bufio"
	url2 "net/url"
	"github.com/mattn/go-gtk/gtk"
)


/**
下载逻辑
 */
func download(url string,obj *SpiderWindow){
	fmt.Println("开始下载的时间：",time.Now())
	sourcetop,completurl,err :=  GetTopInfo(url)
	if err != nil{
		obj.MsgLabel.SetText("获取m3u8顶级地址信息失败！")
		gtk.MainIteration()
		fmt.Println("获取m3u8顶级地址信息失败！")
		return
	}else{
		obj.MsgLabel.SetText("获取m3u8顶级地址信息成功！")
		gtk.MainIteration()
	}


	//获取视频列表
	obj.MsgLabel.SetText("正在获取视频播放列表，请稍等...！")
	sliceaddr,key,err := GetVideoListAndKey(sourcetop,completurl)
	if err != nil{
		obj.MsgLabel.SetText("获取的视频列表失败！")
		gtk.MainIteration()
		fmt.Println("获取的视频列表失败！")
		return
	}else{
		obj.MsgLabel.SetText("获取的视频列表成功！")
		gtk.MainIteration()
	}

	fmt.Println("视频列表下载完成的时间：",time.Now())
	//并发下载视频
	obj.MsgLabel.SetText("开始下载列表中的视频")
	gtk.MainIteration()
	page := make(chan int,20)
	for index,value := range sliceaddr{
		suburl := ""
		if strings.HasSuffix(completurl,"/"){
			suburl = completurl + value
		}else{
			suburl = completurl + "/"+ value
		}
		go DownLoadVideo(suburl,index,page,key)
	}


	//管道读取数据
	alreadyDownLoadNums := 0
	allneedDownLoadNums := len(sliceaddr)
	for _,_= range sliceaddr{
		select {
			case v := <-page:
				fmt.Printf("第%d个子视频下载完成！\n",v)
				obj.MsgLabel.SetText("视频片段"+strconv.Itoa(v)+"下载完成!")
				alreadyDownLoadNums ++
				completProgress := float64(alreadyDownLoadNums)/float64(allneedDownLoadNums)
				obj.Progress.SetFraction(completProgress)
				obj.Progress.SetText(strconv.Itoa(int(completProgress*100))+"%")
				gtk.MainIteration()
				//return
			case <-time.After(20*time.Second) :
				close(page)
				//删除文件
				for _,value := range sliceaddr{
					//打开文件
					sliceurl := strings.Split(value,"/")
					filename := sliceurl[len(sliceurl)-1]
					os.Remove(filename)
				}
				obj.MsgLabel.SetText("视频下载超时!")
				obj.Progress.SetFraction(0)
				obj.Progress.SetText("0%")
				gtk.MainIteration()
				return
				
		}
	}
	fmt.Println("所有视频下载完成的时间：",time.Now())
	//合并视频并删除视频片段
	mergeerr := MergeVideo(sliceaddr)
	if mergeerr != nil{
		obj.MsgLabel.SetText("合并视频片段失败！")
		gtk.MainIteration()
		fmt.Println("合并视频片段失败！失败原因：",mergeerr)
		//合并失败时需要删除源文件
		for _,value := range sliceaddr{
			//打开文件
			sliceurl := strings.Split(value,"/")
			filename := sliceurl[len(sliceurl)-1]
			os.Remove(filename)
		}
		return
	}
	obj.MsgLabel.SetText("视频合并完成！")
	gtk.MainIteration()
	fmt.Println("视频合并完成的时间：",time.Now())
	dialog := gtk.NewMessageDialog(
		obj.Win, //指定父窗口
		gtk.DIALOG_MODAL,              //模态对话框
		gtk.MESSAGE_QUESTION,          //指定对话框类型
		gtk.BUTTONS_YES_NO,            //默认按钮
		"恭喜下载合并完成！")                   //设置内容
	dialog.SetTitle("下载完成！") //对话框设置标题
	flag := dialog.Run() //运行对话框
	if flag == gtk.RESPONSE_YES {
		gtk.MainQuit()
	} else if flag == gtk.RESPONSE_NO {
		gtk.MainQuit()
	}
	dialog.Destroy() //销毁对
}


/**
获取网页中的m3u8地址
 */
func Getm3u8Addr(url string)(pregurl string,error error){
	res,err := http.Get(url)
	if err != nil{
		error = err
		return
	}
	defer res.Body.Close()
	buf := make([]byte,4096)
	html := ""
	for{
		n,err :=res.Body.Read(buf)
		if n==0{
			break
		}
		if err != nil && err != io.EOF{
			error = err
			return
		}
		html += string(buf)
	}
	reg := regexp.MustCompile(`(?i:^http|https):(.*?)*(\.m3u8)`)
	result := reg.FindStringSubmatch(html)
	if len(result) > 0{
		decodeurl,_ :=url2.QueryUnescape(result[0])
		pregurl = strings.ReplaceAll(decodeurl,"\\","")
	}
	return
}


/**
获取顶级m3u8地址，以及地址信息
 */
func GetTopInfo(url string)(topurl string,completurl string,error error){
	respon,err :=http.Get(url)
	if err != nil{
		error = err
		return
	}
	defer respon.Body.Close()
	rd := bufio.NewReader(respon.Body)
	//读取最后一行
	for{
		line,err := rd.ReadString('\n')
		//fmt.Println(line)
		if err != nil && err != io.EOF{
			error = err
			return
		}
		if strings.Contains(line,".m3u8"){
			line = strings.Replace(line, "\n", "", -1)
			topurl += line
		}
		if err == io.EOF{
			break
		}
	}
	//本来就是一级目录了
	if topurl == ""{
		topurl = url
		other := strings.Split(topurl,"/")
		completurl = strings.Join(other[:len(other)-1],"/")
	}else{
		if strings.HasPrefix(topurl,"/"){
			//在网站跟目录下
			topslice := strings.Split(url,"/")[:3]
			completurl = strings.Join(topslice,"/")
			topurl = strings.Join(topslice,"/") +topurl
		}else{
			//和顶级文件同级
			topslice := strings.Split(url,"/")
			newslice := topslice[:len(topslice)-1]
			topurl = strings.Join(newslice,"/") + "/"+topurl
			other := strings.Split(topurl,"/")
			completurl = strings.Join(other[:len(other)-1],"/")
		}
	}
	return
}


/**
获取视频列表以及视频加密的key
 */
func GetVideoListAndKey(url string,completurl string)(sliceaddr []string,key string,error error){
	subres,err := http.Get(url)
	fmt.Println(url)
	if err != nil{
		fmt.Println(err.Error())
		error = err
		return
	}
	defer subres.Body.Close()
	//按行读取,将片段内容地址存放到切片中
	reader := bufio.NewReader(subres.Body)
	reg := regexp.MustCompile(`URI="(.*?)"`)
	keyurl := ""
	for{
		line,err := reader.ReadString('\n')
		if err != nil && err != io.EOF{
			error = err
			return
		}
		//fmt.Println(line)
		if strings.Contains(line,".ts"){
			//去除后面的换行符
			line = strings.Replace(line, "\n", "", -1)
			sliceaddr = append(sliceaddr,line)
		}
		result := reg.FindStringSubmatch(line)
		if len(result) > 0{
			keyurl = result[1]
		}
		if io.EOF==err{
			break
		}
	}

	//同理对key进行处理
	keyrequesturl := ""
	if keyurl != ""{
		if strings.HasPrefix(keyurl,"/"){
			keyrequesturl = completurl  +keyurl
		}else{
			keyrequesturl = completurl + "/" +keyurl
		}
	}
	if keyrequesturl != ""{
		//发起请求，获取key
		keyresp,err :=http.Get(keyrequesturl)
		if err != nil{
			error = err
			return
		}
		defer keyresp.Body.Close()
		buf:= make([]byte,4096)
		for{
			n,err := keyresp.Body.Read(buf)
			if err !=nil && err != io.EOF{
				error =err
				return
			}
			if n==0{
				break
			}
			key += string(buf[:n])
		}
	}
	return
}


/**
视频文件合并
 */
func MergeVideo(slice []string)(error error){
	//创建一个文件，防止重名，用时间戳来定义文件名
	timeunix := strconv.Itoa(int(time.Now().Unix()))
	filename :=timeunix+".mp4"
	fp,err := os.Create(filename)
	if err != nil{
		error =err
		return
	}
	filecontent := ""
	for _,value := range slice{
		//打开文件
		sliceurl := strings.Split(value,"/")
		filename := sliceurl[len(sliceurl)-1]
		subfp,err := os.Open(filename)
		if err !=nil{
			error =err
			return
		}
		singlecontent := ""
		//读取文件内容
		buf:= make([]byte,4096)
		for{
			n,err := subfp.Read(buf)
			if err !=nil && err != io.EOF{
				error =err
				return
			}
			if n==0{
				break
			}
			singlecontent += string(buf[:n])
		}
		filecontent += singlecontent
		subfp.Close()
		removeerr := os.Remove(filename)
		if removeerr != nil{
			error = removeerr
			return
		}
	}
	//写入数据
	fp.WriteString(filecontent)
	fp.Close()
	return
}


/**
下载视频文件
 */
func DownLoadVideo(url string,index int,page chan<-int,key string){
	subrespon,err := http.Get(url)
	if err != nil{
		fmt.Println("获取视频失败！",err)
		return
	}

	//读取内容
	buf := make([]byte,4096)
	content := ""
	for{
		n,err :=subrespon.Body.Read(buf)
		if n == 0{
			break;
		}
		if err != nil && err != io.EOF{
			fmt.Println("读取视频文件失败！",err)
			return
		}
		//将内容写入视频文件中
		content += string(buf[:n])
	}
	if key != ""{
		decodeBytes, err := DecryptAes([]byte(content),[]byte(key))
		if err != nil{
			fmt.Println("解密视频失败！",err)
			return
		}
		content = string(decodeBytes)
	}
	if content != ""{
		//创建文件
		sliceurl := strings.Split(url,"/")
		filename := sliceurl[len(sliceurl)-1]
		fp,err := os.Create(filename)
		if err != nil{
			fmt.Println("创建文件失败！",err)
			return
		}
		fp.WriteString(content)
		fp.Close()
	}
	subrespon.Body.Close()
	page<-index+1
}


/**
视频文件解密
 */
func DecryptAes(crypted []byte,key []byte)(decodebytes []byte,error error){
	block, err := aes.NewCipher(key)
	if err != nil {
		error =err
		return
	}
	crypted = PKCS5Padding(crypted , len(key))
	blockMode := cipher.NewCBCDecrypter(block, key[:block.BlockSize()])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	decodebytes = origData
	return
}


/**
内容填充
 */
func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}
