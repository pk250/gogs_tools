package controllers

type ListsController struct {
	BaseController
}

func (this *ListsController) Compile() {
	this.Data["json"] = map[string]interface{}{"commits": "dasdasasds"}
	this.ServeJSON()
	return
}
