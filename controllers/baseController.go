package controllers

import (
	"crypto/sha256"
	"dailyfresh/constants"
	"dailyfresh/models"
	"encoding/hex"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
)

type BaseController struct {
	beego.Controller
}

//定义json数据格式

type ResponseJSON struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}

func (this *BaseController) Sha256Str(s string) string {
	bytes := sha256.Sum256([]byte(s))
	return hex.EncodeToString(bytes[:])
}

func (this *BaseController) GetUserName() string {
	session := this.GetSession(constants.SESSION_KEY_USERNAME)
	if session != nil {
		username := session.(string)
		this.Data["username"] = username
		return username
	} else {
		this.Data["username"] = ""
		return ""
	}
}
func (this *BaseController) GetUserId(o orm.Ormer, username string) int {
	user := models.User{UserName: username}
	o.Read(&user, "UserName")
	return user.Id
}

func (this *BaseController) SetLayout(layoutName string) {
	this.Layout = layoutName
}
