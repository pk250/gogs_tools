package models

import "time"

type Commits struct {
	Id      string `json:"id"`
	Message string `json:"message"`
	Url     string `json:"url"`
	Author  struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Username string `json:"username"`
	} `json:"author"`
	Committer struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Username string `json:"username"`
	} `json:"committer"`
	Timestamp time.Time `json:"timestamp"`
}

type Repository struct {
	Id    int `json:"id"`
	Owner struct {
		Id         int    `json:"id"`
		Login      string `json:"login"`
		Full_Name  string `json:"full_name"`
		Email      string `json:"email"`
		Avatar_Url string `json:"avatar_url"`
		Username   string `json:"username"`
	} `json:"owner"`
	Name              string    `json:"name"`
	Full_Name         string    `json:"full_name"`
	Description       string    `json:"description"`
	Private           bool      `json:"private"`
	Fork              bool      `json:"fork"`
	Html_Url          string    `json:"html_url"`
	Ssh_Url           string    `json:"ssh_url"`
	Clone_Url         string    `json:"clone_url"`
	Website           string    `json:"website"`
	Stars_Count       int       `json:"stars_count"`
	Forks_Count       int       `json:"forks_count"`
	Watchers_Count    int       `json:"watchers_count"`
	Open_Issues_Count int       `json:"open_issues_count"`
	Default_Branch    string    `json:"default_branch"`
	Created_At        time.Time `json:"created_at"`
	Updated_At        time.Time `json:"updated_at"`
}

type Pusher struct {
	Id         int    `json:"id"`
	Login      string `json:"login"`
	Full_Name  string `json:"full_name"`
	Email      string `json:"email"`
	Avatar_Url string `json:"avatar_url"`
	Username   string `json:"username"`
}

type Sender struct {
	Id         int    `json:"id"`
	Login      string `json:"login"`
	Full_Name  string `json:"full_name"`
	Email      string `json:"email"`
	Avatar_Url string `json:"avatar_url"`
	Username   string `json:"username"`
}

type Gogs struct {
	Ref         string     `json:"ref"`
	Before      string     `json:"before"`
	After       string     `json:"after"`
	Compare_Url string     `json:"compare_url"`
	Commits     []Commits  `json:"commits"`
	Repository  Repository `json:"repository"`
	Pusher      Pusher     `json:"pusher"`
	Sender      Sender     `json:"sender"`
}

type GogsDB struct {
	Id                   int64
	Ref                  string `orm:"size(64);"`
	Before               string `orm:"size(64);"`
	After                string `orm:"size(64);"`
	Commits_Id           string `orm:"size(64);unique"`
	Commits_Message      string
	Commits_Author_Name  string `orm:"size(64);"`
	Commits_Username     string `orm:"size(64);"`
	Commits_Timestamp    time.Time
	Repository_Name      string `orm:"size(64);"`
	Repository_Full_Name string `orm:"size(64);"`
	Repository_Clone_Url string
	Push_Username        string `orm:"size(64);"`
	Push_Email           string `orm:"size(64);"`
	Sender_Username      string `orm:"size(64);"`
	Sender_Email         string `orm:"size(64);"`
}
