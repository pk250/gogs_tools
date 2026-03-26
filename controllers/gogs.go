package controllers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"gogs_tools/models"
	"gogs_tools/services"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

type GogsControllers struct {
	BaseController
}

func (this *GogsControllers) Post() {
	var gogs models.Gogs
	var info []string
	data := this.Ctx.Input.RequestBody
	if err := json.Unmarshal(data, &gogs); err != nil {
		this.Ctx.ResponseWriter.WriteHeader(400)
		this.Data["json"] = map[string]interface{}{"status": 0, "msg": "请求体解析失败"}
		this.ServeJSON()
		return
	}

	// Webhook 密钥验签（按仓库配置）
	o := orm.NewOrm()
	var repoCfgForSig models.RepoConfig
	if err := o.QueryTable("repo_config").Filter("RepoName", gogs.Repository.Name).One(&repoCfgForSig); err == nil && repoCfgForSig.WebhookSecret != "" {
		sig := this.Ctx.Request.Header.Get("X-Gogs-Signature")
		mac := hmac.New(sha256.New, []byte(repoCfgForSig.WebhookSecret))
		mac.Write(data)
		expected := hex.EncodeToString(mac.Sum(nil))
		if !hmac.Equal([]byte(sig), []byte(expected)) {
			this.Ctx.ResponseWriter.WriteHeader(401)
			this.Data["json"] = map[string]interface{}{"status": 0, "msg": "Webhook 签名校验失败"}
			this.ServeJSON()
			return
		}
	}
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
		_, err := o.Insert(&gogsDB)
		if err != nil {
			info = append(info, "gogs insert fail:"+commit.Id)
		} else {
			//!<插入待编译列表
			datainfo := models.Datainfos{
				StorageName: gogs.Repository.Full_Name,
				CommitValue: commit.Id,
				CommitTime:  commit.Timestamp,
				CommitAuth:  commit.Author.Name,
				CommitLogs:  commit.Message,
			}
			_, err2 := o.Insert(&datainfo)
			if err2 != nil {
				info = append(info, "datainfo insert fail:"+commit.Id)
				fmt.Print(err2)
			} else {
				// 判断触发模式，自动模式则入队
				repoConfig := models.RepoConfig{RepoName: gogs.Repository.Name}
				if err2 := o.Read(&repoConfig, "RepoName"); err2 == nil && repoConfig.TriggerMode == "auto" {
					if _, enqErr := services.Enqueue(gogs.Repository.Name, commit.Id, commit.Message, commit.Author.Name); enqErr != nil {
						logs.Warn("[Gogs] Enqueue failed repo=%s commit=%s err=%v", gogs.Repository.Name, commit.Id, enqErr)
					}
				}
			}
		}

	}

	this.Ctx.ResponseWriter.WriteHeader(200)
	this.Data["json"] = map[string]interface{}{"status": 1, "msg": "推送成功", "url": "/gogs", "fail_info": info}
	this.ServeJSON()
	return
}
