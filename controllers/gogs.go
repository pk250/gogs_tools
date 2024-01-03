package controllers

import (
	"encoding/json"
	"fmt"
	"gogs_tools/models"

	"github.com/astaxie/beego/orm"
)

type GogsControllers struct {
	BaseController
}

func (this *GogsControllers) Post() {
	var gogs models.Gogs
	var info []string
	data := this.Ctx.Input.RequestBody
	err := json.Unmarshal(data, &gogs)
	if err != nil {
		panic(err)
	}

	o := orm.NewOrm()
	for _, commit := range gogs.Commits {
		gogsDB := models.GogsDB{Ref: gogs.Ref, Before: gogs.Before, After: gogs.After,
			Commits_Id:      commit.Id,
			Commits_Message: commit.Message, Commits_Author_Name: commit.Author.Name,
			Commits_Username: commit.Author.Username, Commits_Timestamp: commit.Timestamp,
			Repository_Name: gogs.Repository.Name, Repository_Full_Name: gogs.Repository.Full_Name,
			Repository_Clone_Url: gogs.Repository.Clone_Url, Push_Username: gogs.Pusher.Username,
			Push_Email: gogs.Pusher.Email, Sender_Username: gogs.Sender.Username,
			Sender_Email: gogs.Sender.Email,
		}
		_, err = o.Insert(&gogsDB)
		if err != nil {
			info = append(info, "gogs insert fail:"+commit.Id)
		} else {
			//!<插入待编译列表
			datainfo := models.Datainfos{
				StorageName: gogs.Repository.Full_Name,
				CommitValue: commit.Id,
				CommitTime:  commit.Timestamp,
				CommitAuth:  commit.Author.Name,
			}
			_, err = o.Insert(&datainfo)
			if err != nil {
				info = append(info, "datainfo insert fail:"+commit.Id)
				fmt.Print(err)
			}
		}

	}

	this.Ctx.ResponseWriter.WriteHeader(200)
	this.Data["json"] = map[string]interface{}{"status": 1, "msg": "推送成功", "url": "/gogs", "fail_info": info}
	this.ServeJSON()
	return
}
