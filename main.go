package main

import (
	"github.com/astaxie/beego"
	"fmt"
	"path/filepath"
	"os"
	"github.com/astaxie/beego/logs"
	"strings"
	"path"
	"io/ioutil"
	"time"
	"encoding/json"
	"net/http"
	"net/url"
)

// 方法：获取当前程序运行的路径(例如：/home/gopath/src/ClientManagement/ClientManagement)
/*
*  传入参数：
*  @Param: Type:
*  @Param: Type:
*  @Param: Type:
*  返回参数：
*  @Param: Type:string
*  @Param: Type:
*/
func GetCurrentDirectory() string {
	dir,err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil{
		logs.Error(err)
	}
	return strings.Replace(dir,"\\","/",-1)
}

type MainController struct {
	beego.Controller
}

type FileInfo struct {
	Name string `json:"name"`
	Size string `json:"size"`
	Modtime string `json:"modtime"`
	Isdir bool `json:"isdir"`
	Path string `json:"path"`
	Type string `json:"type"`
}

type Message struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (c *MainController) ListDirFor(dir string) ([]FileInfo,error) {
	var files []FileInfo
	CurrentRunPath := GetCurrentDirectory()
	fullPath := path.Join(CurrentRunPath,dir)
	dirs,err := ioutil.ReadDir(fullPath)
	if err != nil{
		return files,err
	}else {
		for _,v := range dirs {
			newSize,newType := FileSizeToHumanRead(float64(v.Size()),"Byte")
			tmpFileSize := fmt.Sprintf("%.2f %s",newSize,newType)
			tmp := FileInfo{
				Name:v.Name(),
				Size:tmpFileSize,
				Modtime:c.TimeFormatTo(v.ModTime()),
				Isdir:v.IsDir(),
				Path:path.Join(dir,v.Name()),
				Type:CheckFileType(v.Name()),
			}
			if !c.InBlockList(dir,tmp.Name){
				files = append(files,tmp)
			}

		}
		return files,nil
	}
}
// Byte,KB,MB,GB,TB,PB
func FileSizeToHumanRead(filesize float64,thetype string) (float64,string) {
	if filesize <= 1024 {
		return filesize,thetype
	}else {
		newSize := float64(filesize) / 1024
		if thetype == "Byte"{
			thetype = "KB"
		}else if thetype == "KB" {
			thetype = "MB"
		}else if thetype == "MB" {
			thetype = "GB"
		}else if thetype == "GB" {
			thetype = "TB"
		}else if thetype == "PB" {
			thetype = "PB"
		}
		if newSize > 1024 {
			newSize,thetype = FileSizeToHumanRead(newSize,thetype)
		}
		return newSize,thetype
	}
}

//获取相关目录的文件列表
func (c *MainController) GetFileList() {

	var files []FileInfo
	dir := c.GetString("dir")
	fmt.Println(dir)
	if dir == ""{
		dir = "/"
	}
	files,err := c.ListDirFor(dir)
	if err  != nil{
		c.Ctx.WriteString(err.Error())
	}
	c.Data["json"] = files
	c.ServeJSON()
}

func (c *MainController) TimeFormatTo(t time.Time) (string) {
	return t.Format("2006-01-02 15:04:05")
}

func (c *MainController) InBlockList(path, filename string) (bool) {
	type BlockFile struct {
		Path string
		Filename string
	}
	var lists []BlockFile
	files := beego.AppConfig.String("block")
	json.Unmarshal([]byte(files),&lists)
	for _,v := range lists{
		if path == v.Path && filename == v.Filename {
			return true
		}
	}
	return false
}

func (c *MainController) GetFileMd5()  {

}

func (c *MainController) ServeFile(path string) {
	//c.Ctx.ResponseWriter.Status = 200
	http.ServeFile(c.Ctx.ResponseWriter,c.Ctx.Request,path)
}


func (c *MainController) ToServeFile()  {
	fullurl := c.Ctx.Request.RequestURI
	parsedurl,err := url.QueryUnescape(fullurl)
	if err != nil{
		fmt.Println(err.Error())
	}
	fmt.Println(parsedurl)
	thepath := parsedurl
	if thepath == "" {
		thepath = "/"
	}
	currentPath := GetCurrentDirectory()
	fullPath := path.Join(currentPath,thepath)
	fmt.Println(fullPath)
	pathinfo,err := os.Stat(fullPath)
	if err != nil{
		c.Ctx.WriteString("找不到文件或者文件夹")
		c.Finish()
	}
	if pathinfo.IsDir() {
		files,err := c.ListDirFor(thepath)
		if err != nil{
			c.Ctx.WriteString("找不到文件或者文件夹")
			c.Finish()
		}
		// 下面这段，当只有一个/的时候，他都会split出两边为空的情况
		var thePaths []string
		if thepath == "/"{
			thePaths = append(thePaths,"")
		}else {
			thePaths = strings.Split(thepath,"/")
		}

		var pathlinks []struct{
			Name string
			Path string
		}
		var count int
		for i := 0 ;i <len(thePaths) ;i ++ {
			fmt.Println(len(thePaths))
			fmt.Println(count)
			var tmp struct{
				Name string
				Path string
			}
			for j := 0 ;j <=i ;j++{
				if tmp.Path == "" {
					tmp.Path = "/"
				}else if tmp.Path == "/" {
					tmp.Path = tmp.Path + thePaths[j]
				}else {
					tmp.Path = tmp.Path + "/" + thePaths[j]
				}
			}
			tmp.Name = thePaths[i]
			pathlinks = append(pathlinks,tmp)
			count += 1
		}
		fmt.Println(thePaths)
		fmt.Println(pathlinks)
		c.Data["paths"] = pathlinks

		c.Data["files"] = files
		c.TplName = "index.html"
		//c.ServeJSON()
		c.Finish()
	}else {
		fmt.Println("这是文件")
		c.ServeFile(fullPath)
	}
}

func GetFileSuffix(filepath string) (suffix,body string) {
	// 由于split出来的不可能小于1 所以就不判断小1的情况了。
	strs := strings.Split(filepath,".")
	if len(strs) == 1 {
		return "",filepath
	}else {
		tmpstr := ""
		for i := 0;i <len(strs) -1 ;i ++ {
			tmpstr = tmpstr + strs[i]
		}
		return strs[1],tmpstr
	}
}

// 检查文件类型
func CheckFileType(filename string) (string) {
	filesuffix,_ := GetFileSuffix(filename)
	switch filesuffix {
	case  "7z"    :
		return "fa-file-archive-o"
	case  "bz"    :
		return "fa-file-archive-o"
	case  "gz"    :
		return "fa-file-archive-o"
	case  "rar"   :
		return "fa-file-archive-o"
	case  "tar"   :
		return "fa-file-archive-o"
	case  "zip"   :
		return "fa-file-archive-o"
	case  "aac"   :
		return "fa-music"
	case  "flac"  :
		return "fa-music"
	case  "mid"   :
		return "fa-music"
	case  "midi"  :
		return "fa-music"
	case  "mp3"   :
		return "fa-music"
	case  "ogg"   :
		return "fa-music"
	case  "wma"   :
		return "fa-music"
	case  "wav"   :
		return "fa-music"
	case  "c"     :
		return "fa-code"
	case  "class" :
		return "fa-code"
	case  "cpp"   :
		return "fa-code"
	case  "css"   :
		return "fa-code"
	case  "erb"   :
		return "fa-code"
	case  "htm"   :
		return "fa-code"
	case  "html"  :
		return "fa-code"
	case  "java"  :
		return "fa-code"
	case  "js"    :
		return "fa-code"
	case  "php"   :
		return "fa-code"
	case  "pl"    :
		return "fa-code"
	case  "py"    :
		return "fa-code"
	case  "rb"    :
		return "fa-code"
	case  "xhtml" :
		return "fa-code"
	case  "xml"   :
		return "fa-code"
	case  "accdb" :
		return "fa-hdd-o"
	case  "db"    :
		return "fa-hdd-o"
	case  "dbf"   :
		return "fa-hdd-o"
	case  "mdb"   :
		return "fa-hdd-o"
	case  "pdb"   :
		return "fa-hdd-o"
	case  "sql"   :
		return "fa-hdd-o"
	case  "csv"   :
		return "fa-file-text"
	case  "doc"   :
		return "fa-file-text"
	case  "docx"  :
		return "fa-file-text"
	case  "odt"   :
		return "fa-file-text"
	case  "pdf"   :
		return "fa-file-text"
	case  "xls"   :
		return "fa-file-text"
	case  "xlsx"  :
		return "fa-file-text"
	case  "app"   :
		return "fa-list-alt"
	case  "bat"   :
		return "fa-list-alt"
	case  "com"   :
		return "fa-list-alt"
	case  "exe"   :
		return "fa-list-alt"
	case  "jar"   :
		return "fa-list-alt"
	case  "msi"   :
		return "fa-list-alt"
	case  "vb"    :
		return "fa-list-alt"
	case  "eot"   :
		return "fa-font"
	case  "otf"   :
		return "fa-font"
	case  "ttf"   :
		return "fa-font"
	case  "woff"  :
		return "fa-font"
	case  "gam"   :
		return "fa-gamepad"
	case  "nes"   :
		return "fa-gamepad"
	case  "rom"   :
		return "fa-gamepad"
	case  "sav"   :
		return "fa-floppy-o"
	case  "bmp"   :
		return "fa-picture-o"
	case  "gif"   :
		return "fa-picture-o"
	case  "jpg"   :
		return "fa-picture-o"
	case  "jpeg"  :
		return "fa-picture-o"
	case  "png"   :
		return "fa-picture-o"
	case  "psd"   :
		return "fa-picture-o"
	case  "tga"   :
		return "fa-picture-o"
	case  "tif"   :
		return "fa-picture-o"
	case  "box"   :
		return "fa-archive"
	case  "deb"   :
		return "fa-archive"
	case  "rpm"   :
		return "fa-archive"
	case  "cmd"   :
		return "fa-terminal"
	case  "sh"    :
		return "fa-terminal"
	case  "cfg"   :
		return "fa-file-text"
	case  "ini"   :
		return "fa-file-text"
	case  "log"   :
		return "fa-file-text"
	case  "md"    :
		return "fa-file-text"
	case  "rtf"   :
		return "fa-file-text"
	case  "txt"   :
		return "fa-file-text"
	case  "ai"    :
		return "fa-picture-o"
	case  "drw"   :
		return "fa-picture-o"
	case  "eps"   :
		return "fa-picture-o"
	case  "ps"    :
		return "fa-picture-o"
	case  "svg"   :
		return "fa-picture-o"
	case  "avi"   :
		return "fa-youtube-play"
	case  "flv"   :
		return "fa-youtube-play"
	case  "mkv"   :
		return "fa-youtube-play"
	case  "mov"   :
		return "fa-youtube-play"
	case  "mp4"   :
		return "fa-youtube-play"
	case  "mpg"   :
		return "fa-youtube-play"
	case  "ogv"   :
		return "fa-youtube-play"
	case  "webm"  :
		return "fa-youtube-play"
	case  "wmv"   :
		return "fa-youtube-play"
	case  "swf"   :
		return "fa-youtube-play"
	case  "bak"   :
		return "fa-floppy"
	case  "msg"   :
		return "fa-envelope"
	default:
		return "fa-file"
	}
}


func main() {
	beego.Router("/listdir", &MainController{},"get,post:GetFileList")
	beego.Router("/*",&MainController{},"*:ToServeFile")
	beego.Run()
}

