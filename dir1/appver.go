package controllers

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"zhugopub/app/entity"
	"zhugopub/app/libs"
	"zhugopub/app/service"

	"github.com/astaxie/beego"
)

type AppverController struct {
	BaseController
}

// // 项目列表 用户写入app_name
// func (this *AppverController) List() {
// 	var list []entity.App

// 	appName := this.GetString("app_name")

// 	//模糊查询
// 	service.AppService.Query().Filter("AppName__icontains", appName).All(&list)

// 	this.Data["app_name"] = appName
// 	this.Data["pageTitle"] = "选择微服务"
// 	this.Data["applist"] = list
// 	this.Data["count"] = len(list)

// 	this.display()
// }

// 微服务版本列表
//传入后端list
func (this *AppverController) List() {
	page, _ := strconv.Atoi(this.GetString("page"))
	if page < 1 {
		page = 1
	}

	count, _ := service.AppverService.GetTotal()
	list, _ := service.AppverService.GetList(page, this.pageSize)

	this.Data["count"] = count
	this.Data["list"] = list

	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("ProjectController.List"), true).ToString()
	this.Data["pageTitle"] = "项目列表"

	//this.curlDisplay()

	this.display()
}

// 添加app_ver
//后端传入前端 appver_name,描述,config,源文件上传方式，源文件类型
func (this *AppverController) Add() {
	var pid int
	id, _ := this.GetInt("id") //id是APP-id
	beego.Trace("pppppid is", id)
	pid = id
	app, err := service.AppService.GetAppById(pid, false)
	this.checkError(err)

	if this.isPost() {
		p := &entity.Appver{}
		p.AppId, _ = this.GetInt("app_id")
		beego.Trace("p appid is", p.AppId)
		p.AppverName = this.GetString("appver_name")
		p.Description = this.GetString("appver_description")
		// p.Attachment = this.GetString("appver_file")
		// beego.Trace("文件是", p.Attachment)
		p.SourceUrl = this.GetString("appver_url")
		beego.Trace("远程URL", p.SourceUrl)
		// p.Config = this.GetString("appver_config")

		p.ConfigUrl = this.GetString("config_up_url")

		p.UserName = this.auth.GetUser().UserName

		//add到数据库
		//源文件上传-goroutine
		//url download->本地-goroutine

		//通过app_id获取appName
		app, err := service.AppService.GetAppById(p.AppId, false)
		this.checkError(err)

		//建立appver-bin目录和appver-conf目录
		appverBinPath := service.GetAppverBinPath(app.AppName, p.AppverName)
		appverConfPath := service.GetAppverConfigPath(app.AppName, p.AppverName)
		os.MkdirAll(appverBinPath, 0755)
		os.MkdirAll(appverConfPath, 0755)

		//源文件上传
		_, fh, err := this.GetFile("appver_file")
		this.checkError(err)
		if fh != nil {
			p.Attachment = fh.Filename
			beego.Trace("attache is", p.Attachment)
			destPath := path.Join(appverBinPath, fh.Filename)
			err = this.SaveToFile("appver_file", destPath)
			this.checkError(err)
		}

		//配置文件上传
		_, configfh, err := this.GetFile("config_up_file")
		this.checkError(err)
		if configfh != nil {
			p.ConfigFile = configfh.Filename
			beego.Trace("attache is", p.ConfigFile)
			destPath := path.Join(appverConfPath, configfh.Filename)
			err = this.SaveToFile("config_up_file", destPath)
			this.checkError(err)
		}

		//判断是否输入的是否有效
		if err := this.validAppver(p); err != nil {
			this.showMsg(err.Error(), MSG_ERR)
		}

		//ADD到APP数据库
		err = service.AppverService.AddAppver(p)
		this.checkError(err)

		//记录用户操作
		//service.ActionService.Add("add_project", this.auth.GetUserName(), "project", p.Id, "")

		this.redirect(beego.URLFor("AppController.Detail", "id", p.AppId))
	}

	this.Data["appId"] = pid
	this.Data["pageTitle"] = "添加" + app.AppName + "版本"
	this.display()

}

// 微服务版本详情 需要传递给后端 app_name/Description/本地文件or remote-url/配置/创建人/更新时间
func (this *AppverController) Detail() {
	id, _ := this.GetInt("id") //appverId
	p, err := service.AppverService.GetAppverById(id)
	this.checkError(err)

	//根据appverId找app_name
	app, err := service.AppService.GetAppById(p.AppId, false)
	this.checkError(err)

	this.Data["appver"] = p
	this.Data["appname"] = app.AppName
	this.Data["pageTitle"] = p.AppverName + "详情"

	this.display()

	//this.curlDisplay()
}

// 编辑appver
func (this *AppverController) Edit() {
	id, _ := this.GetInt("id") //appverId
	p, err := service.AppverService.GetAppverById(id)
	this.checkError(err)

	//根据appverId找app_name
	app, err := service.AppService.GetAppById(p.AppId, false)
	this.checkError(err)

	if this.isPost() {

		p.AppverName = this.GetString("appver_name")
		p.Description = this.GetString("appver_description")
		p.SourceUrl = this.GetString("appver_url")
		p.ConfigUrl = this.GetString("config_up_url")
		p.UserName = this.auth.GetUser().UserName

		//上传新文件+删除老文件
		//获取目录
		appverBinPath := service.GetAppverBinPath(app.AppName, p.AppverName)
		appverConfPath := service.GetAppverConfigPath(app.AppName, p.AppverName)

		//源文件上传
		_, fh, err := this.GetFile("appver_file")
		if err != nil {
			if err.Error() != "http: no such file" {
				this.checkError(err)
			}
		} else {
			if fh != nil {
				p.Attachment = fh.Filename
				beego.Trace("attachesource is", p.Attachment)
				destPath := path.Join(appverBinPath, fh.Filename)
				beego.Trace("destPath is", destPath)
				err = this.SaveToFile("appver_file", destPath)
				this.checkError(err)

				//删除老文件
			}
		}

		//配置文件上传
		_, configfh, err := this.GetFile("config_up_file")
		if err != nil {
			if err.Error() != "http: no such file" {
				this.checkError(err)
			}
		} else {
			if configfh != nil {
				p.ConfigFile = configfh.Filename
				beego.Trace("attache is", p.ConfigFile)
				destPath := path.Join(appverConfPath, configfh.Filename)
				err = this.SaveToFile("config_up_file", destPath)
				this.checkError(err)

				//删除老文件
			}
		}

		//判断是否输入的是否有效
		if err := this.validAppver(p); err != nil {
			this.showMsg(err.Error(), MSG_ERR)
		}
		//appver表更新
		err = service.AppverService.UpdateAppver(p, "AppverName", "Description", "Attachment", "SourceUrl", "ConfigFile", "ConfigUrl", "UserName")
		this.checkError(err)

		//记录用户行为
		//service.ActionService.Add("edit_project", this.auth.GetUserName(), "project", p.Id, "")

		this.redirect(beego.URLFor("AppController.Detail", "id", app.Id))

	}

	this.Data["appver"] = p
	this.Data["appname"] = app.AppName
	this.Data["pageTitle"] = "编辑微服务"
	this.display()
}

// 验证提交
func (this *AppverController) validAppver(p *entity.Appver) error {
	errorMsg := ""
	if p.AppverName == "" {
		errorMsg = "请输入微服务版本名称"
	} else if p.Attachment == "" && p.SourceUrl == "" {
		errorMsg = "请上传本地文件或者输入远端文件URL"
	} else if p.ConfigFile == "" && p.ConfigUrl == "" {
		errorMsg = "请上传配置文件或者输入远端文件URL"
	}
	if errorMsg != "" {
		return fmt.Errorf(errorMsg)
	}
	return nil
}

