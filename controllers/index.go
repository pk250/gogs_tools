package controllers

type IndexController struct {
	BaseController
}

func (this *IndexController) Get() {
	// this.Layout = "index.html"
	// this.TplName = "datainfo.html"
	this.Redirect("layout/datainfo", 302)
}
