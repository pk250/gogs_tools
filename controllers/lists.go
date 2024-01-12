package controllers

import (
	"archive/zip"
	"bufio"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/astaxie/beego"
	"github.com/gorilla/websocket"
)

type ListsController struct {
	BaseController
}

type CompileMessage struct {
	MdkType     string `json:"mdkVersion"`
	ProUrl      string `json:"urlPath"`
	Commit      string `json:"commitValue"`
	ProjectName string `json:"projectName"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func GetProjectFileName(path string, name string) string {
	var result string
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		//log.Println(path, "---", info.Name())
		if name == info.Name() {
			result = path
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return result
}

func GetFileListForType(fpath string, ty ...string) []string {
	var result, tys []string

	for _, v := range ty {
		tys = append(tys, filepath.Ext(v))
	}

	err := filepath.Walk(fpath, func(path string, info os.FileInfo, err error) error {
		//log.Println(path, "---", info.Name())
		for _, v := range tys {
			if v == filepath.Ext(path) {
				result = append(result, path)
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return result
}

func SaveBinaryFile(sourcePath []string, targetPath string) error {

	zipFile, err := os.Create(targetPath + ".zip")
	if err != nil {
		return err
	}
	defer zipFile.Close()
	zipWriter := zip.NewWriter(zipFile)

	for _, spath := range sourcePath {
		//!<获取文件名
		name := filepath.Base(spath)
		//!<在压缩包中创建文件
		fileWriter, err := zipWriter.Create(name)
		if err != nil {
			return err
		}
		//!<打开待压缩文件
		file, err := os.Open(spath)
		if err != nil {
			return err
		}
		defer file.Close()
		//!<拷贝文件到压缩包
		if _, err = io.Copy(fileWriter, file); err != nil {
			return err
		}
	}

	zipWriter.Close()

	return nil
}

func WriteCmdWs(ws *websocket.Conn, cmd *exec.Cmd) {
	stdout, err := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout

	if err != nil {
		log.Panic(err)
	}
	if err = cmd.Start(); err != nil {
		log.Panic(err)
	}
	for {
		tmp := make([]byte, 128)
		_, err := stdout.Read(tmp)
		if err != nil {
			break
		}
		err = ws.WriteMessage(websocket.BinaryMessage, tmp)
		if err != nil {
			log.Panic(err)
			ws.Close()
		}
	}

}

func WriteLogsWs(ws *websocket.Conn, logs string) {
	err := ws.WriteMessage(websocket.BinaryMessage, []byte(logs))
	if err != nil {
		log.Panic(err)
		ws.Close()
	}
}

func ReadLogs(ws *websocket.Conn, path string, name string) (string, error) {
	logsPath := GetProjectFileName(path, name)
	file, err := os.Open(logsPath)
	if err != nil {
		return logsPath, err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		err = ws.WriteMessage(websocket.BinaryMessage, line)
		if err != nil {
			ws.Close()
			return logsPath, err
		}
		err = ws.WriteMessage(websocket.BinaryMessage, []byte("\r\n"))
		if err != nil {
			ws.Close()
			return logsPath, err
		}
	}
	return logsPath, nil
}

func CreateCompileTask(this *ListsController) {
	ws, err := upgrader.Upgrade(this.Ctx.ResponseWriter, this.Ctx.Request, nil)
	if err != nil {
		log.Panic(err)
	}
	defer ws.Close()

	_, data, err := ws.ReadMessage()
	if err != nil {
		log.Panic(err)
		ws.Close()
	}
	log.Println(string(data))
	var msg CompileMessage
	err = json.Unmarshal(data, &msg)
	if err != nil {
		log.Panic(err)
	}
	log.Println(msg)
	WriteLogsWs(ws, "开始创建本地仓库\r\n")

	//!<创建文件夹

	if err := os.MkdirAll(beego.AppConfig.String("ClonePath")+"/"+msg.Commit, os.ModePerm); err == nil {
		log.Println("文件夹已经存在，先删除")
		os.RemoveAll(beego.AppConfig.String("ClonePath") + "/" + msg.Commit)
		os.MkdirAll(beego.AppConfig.String("ClonePath")+"/"+msg.Commit, os.ModePerm)
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = beego.AppConfig.String("ClonePath") + "/" + msg.Commit
	WriteCmdWs(ws, cmd)

	cmd = exec.Command("git", "remote", "add", "origin", msg.ProUrl)
	cmd.Dir = beego.AppConfig.String("ClonePath") + "/" + msg.Commit
	WriteCmdWs(ws, cmd)

	cmd = exec.Command("git", "fetch", "origin", msg.Commit)
	cmd.Dir = beego.AppConfig.String("ClonePath") + "/" + msg.Commit
	WriteCmdWs(ws, cmd)

	cmd = exec.Command("git", "merge", msg.Commit)
	cmd.Dir = beego.AppConfig.String("ClonePath") + "/" + msg.Commit
	WriteCmdWs(ws, cmd)
	//!<开始查找工程文件
	WriteLogsWs(ws, "开始查找工程文件\r\n")

	pjPath := GetProjectFileName(beego.AppConfig.String("ClonePath")+"/"+msg.Commit, msg.ProjectName)

	WriteLogsWs(ws, pjPath+"\r\n")
	WriteLogsWs(ws, "开始编译\r\n")
	cmd = exec.Command(beego.AppConfig.String("Mdk5"), "-b", "-j0", pjPath, "-o", "build_log.txt")
	//cmd.Dir = beego.AppConfig.String("ClonePath") + "/" + commit
	WriteCmdWs(ws, cmd)
	WriteLogsWs(ws, "获取编译日志\r\n")
	logsPath, err := ReadLogs(ws, beego.AppConfig.String("ClonePath")+"/"+msg.Commit, "build_log.txt")
	if err != nil {
		log.Panic(err)
	}

	//!<打包编译好的文件
	//!<step 1 查找编译出来的bin、axf、hex、map文件
	files := GetFileListForType(beego.AppConfig.String("ClonePath")+"/"+msg.Commit, ".bin", ".axf", ".hex", ".map")
	files = append(files, logsPath)
	SaveBinaryFile(files, beego.AppConfig.String("binout")+"/"+msg.Commit)
	//!<step 2 保存日志文件
	os.RemoveAll(beego.AppConfig.String("ClonePath") + "/" + msg.Commit)
	ws.Close()
}

func (this *ListsController) Compile() {
	this.Data["json"] = map[string]interface{}{"commits": "dasdasasds"}
	this.ServeJSON()

	return
}

func (this *ListsController) Commit() {
	this.Data["json"] = map[string]interface{}{"status": "1", "msg": "success"}
	this.ServeJSON()
	return
}

func (this *ListsController) Ws() {
	CreateCompileTask(this)
	this.EnableRender = false
}
