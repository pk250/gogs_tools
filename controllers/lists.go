package controllers

import (
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

func GetProjectFile(path string, name string) string {
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

func ReadLogs(ws *websocket.Conn, path string, name string) {
	file, err := os.Open(GetProjectFile(path, name))
	if err != nil {
		log.Panic(err)
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
			log.Panic(err)
			ws.Close()
		}
		err = ws.WriteMessage(websocket.BinaryMessage, []byte("\r\n"))
		if err != nil {
			log.Panic(err)
			ws.Close()
		}
	}
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

	pjPath := GetProjectFile(beego.AppConfig.String("ClonePath")+"/"+msg.Commit, msg.ProjectName)

	WriteLogsWs(ws, pjPath+"\r\n")
	WriteLogsWs(ws, "开始编译\r\n")
	cmd = exec.Command(beego.AppConfig.String("Mdk5"), "-b", "-j0", pjPath, "-o", "build_log.txt")
	//cmd.Dir = beego.AppConfig.String("ClonePath") + "/" + commit
	WriteCmdWs(ws, cmd)
	WriteLogsWs(ws, "获取编译日志\r\n")
	ReadLogs(ws, beego.AppConfig.String("ClonePath")+"/"+msg.Commit, "build_log.txt")

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
