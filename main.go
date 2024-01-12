package main

import (
	_ "gogs_tools/routers"
	"log"
	"os"

	"github.com/astaxie/beego"
)

func init() {
	if err := os.MkdirAll(beego.AppConfig.String("ClonePath"), os.ModePerm); err != nil {
		log.Panic(err)
	}
	if err := os.MkdirAll(beego.AppConfig.String("binout"), os.ModePerm); err != nil {
		log.Panic(err)
	}
}

func main() {
	beego.Run()
}
