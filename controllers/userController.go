package controllers

import (
	"dailyfresh/constants"
	"dailyfresh/models"
	"encoding/base64"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/gomodule/redigo/redis"
	"strconv"
)

type UserController struct {
	BaseController
}

func (this *UserController) ShowRegister() {
	this.TplName = "register.html"
}

func (this *UserController) HandleRegister() {
	//var param = {username,pwd,cpwd,email};
	username := this.GetString("username")
	pwd := this.GetString("pwd")
	cpwd := this.GetString("cpwd")
	email := this.GetString("email")
	defer this.ServeJSON()
	if username == "" {
		this.Data["json"] = &ResponseJSON{401, nil, "用户名不合法"}
		return
	}
	if pwd == "" {
		this.Data["json"] = &ResponseJSON{401, nil, "密码不合法"}
		return
	}
	if pwd != cpwd {
		this.Data["json"] = &ResponseJSON{401, nil, "两次输入的密码不一致"}
		return
	}
	if email == "" {
		this.Data["json"] = &ResponseJSON{401, nil, "邮箱不能为空"}
		return
	}
	o := orm.NewOrm()
	user := models.User{UserName: username, PassWord: this.Sha256Str(pwd), Email: email, Active: false}
	read := o.Read(&user, "UserName")
	if read == nil {
		this.Data["json"] = &ResponseJSON{401, nil, "用户名已存在,请更换用户名注册"}
		return
	}
	_, err := o.Insert(&user)
	if err != nil {
		this.Data["json"] = &ResponseJSON{401, nil, "用户注册失败,请重新注册"}
		return
	}
	//dhhost:=beego.AppConfig.String("dbhost")
	this.Data["json"] = &ResponseJSON{200, "/active?id=" + strconv.Itoa(user.Id), "用户注册成功,请准备激活"}
}

func (this *UserController) ShowActive() {
	defer this.ServeJSON()
	id, e := this.GetInt("id")
	if e != nil {
		this.Data["json"] = &ResponseJSON{401, nil, "数据传入有误"}
		return
	}
	user := models.User{Id: id}
	o := orm.NewOrm()
	read := o.Read(&user)
	if read != nil {
		this.Data["json"] = &ResponseJSON{401, nil, "用户激活失败"}
		return
	}
	user.Active = true
	count, i := o.Update(&user)
	if count == 0 || i != nil {
		this.Data["json"] = &ResponseJSON{401, nil, "用户激活失败"}
		return
	}
	this.Data["json"] = &ResponseJSON{200, nil, "用户激活成功"}
}

func (this *UserController) ShowLogin() {
	cookie := this.Ctx.GetCookie(constants.COOKIE_KEY_USERNAME)
	if cookie != "" {
		bytes, _ := base64.StdEncoding.DecodeString(cookie)
		username := string(bytes)
		this.Data["username"] = username
		this.Data["checked"] = "checked"
	} else {
		this.Data["username"] = ""
		this.Data["checked"] = ""
	}
	this.TplName = "login.html"
}

func (this *UserController) HandleLogin() {
	defer this.ServeJSON()
	username := this.GetString("username")
	pwd := this.GetString("pwd")
	checked := this.GetString("checked")
	if username == "" {
		this.Data["json"] = &ResponseJSON{403, nil, "用户名不合法"}
		return
	}
	if pwd == "" {
		this.Data["json"] = &ResponseJSON{403, nil, "密码不合法"}
		return
	}
	user := models.User{}
	o := orm.NewOrm()
	one := o.QueryTable("User").Filter("UserName", username).Filter("PassWord", this.Sha256Str(pwd)).One(&user)
	if one != nil {
		this.Data["json"] = &ResponseJSON{403, nil, "用户名或密码有误"}
		return
	}
	if checked == "on" {
		this.Ctx.SetCookie(constants.COOKIE_KEY_USERNAME, base64.StdEncoding.EncodeToString([]byte(username)), constants.COOKIE_EXPIRE)
	} else {
		this.Ctx.SetCookie(constants.COOKIE_KEY_USERNAME, base64.StdEncoding.EncodeToString([]byte(username)), -1)
	}
	this.SetSession(constants.SESSION_KEY_USERNAME, username)
	this.Data["json"] = &ResponseJSON{200, nil, "登录成功"}

}

func (this *UserController) ShowLogout() {
	this.DelSession(constants.SESSION_KEY_USERNAME)
	this.Redirect("/login", 302)
}

//用户模块的
func (this *UserController) ShowUserCenterInfo() {
	this.Data["flag"] = 1
	username := this.GetUserName()
	o := orm.NewOrm()
	//查询地址
	var address models.Address
	o.QueryTable("Address").RelatedSel("User").Filter("User__UserName", username).Filter("Isdefault", true).One(&address)
	if address.Id == 0 {
		this.Data["address"] = ""
	} else {
		this.Data["address"] = address
	}

	//获取浏览的商品
	dbhost := beego.AppConfig.String("dbhost")
	conn, err := redis.Dial("tcp", dbhost+":6379")
	if err != nil {
		beego.Error(err)
		this.Redirect("/", 302)
		return
	}
	defer conn.Close()
	skuIds, e := redis.Ints(conn.Do("lrange", constants.PRE_SCAN+strconv.Itoa(this.GetUserId(o, username)), 0, 4))
	if e != nil {
		beego.Error(e)
		this.Redirect("/", 302)
		return
	}
	scanGoods := make([]models.GoodsSKU, 0)
	for _, skuid := range skuIds {
		goodsSku := models.GoodsSKU{Id: skuid}
		o.Read(&goodsSku)
		scanGoods = append(scanGoods, goodsSku)
	}
	this.Data["scanGoods"] = scanGoods
	this.SetLayout("userLayout.html")
	this.TplName = "user_center_info.html"
}

func (this *UserController) ShowUserCenterOrder() {
	userName := this.GetUserName()
	user := models.User{UserName: userName}
	o := orm.NewOrm()
	o.Read(&user, "UserName")
	var orderInfos []models.OrderInfo
	o.QueryTable("OrderInfo").RelatedSel("User").Filter("User__Id", user.Id).All(&orderInfos)
	goodsBuffer := make([]map[string]interface{}, len(orderInfos))
	for index, orderInfo := range orderInfos {
		var orderGoodsSlice []models.OrderGoods
		o.QueryTable("OrderGoods").RelatedSel("OrderInfo", "GoodsSKU").Filter("OrderInfo__Id", orderInfo.Id).All(&orderGoodsSlice)
		temp := make(map[string]interface{})
		temp["orderInfo"] = orderInfo
		temp["orderGoodsSlice"] = orderGoodsSlice
		goodsBuffer[index] = temp
	}
	this.Data["goodsBuffer"] = goodsBuffer
	//beego.Info("goodsBuffer==========",goodsBuffer)
	this.Data["flag"] = 2
	this.SetLayout("userLayout.html")
	this.TplName = "user_center_order.html"
}

func (this *UserController) ShowUserCenterSite() {
	this.Data["flag"] = 3
	username := this.GetUserName()
	o := orm.NewOrm()
	//查询地址
	var addresses []models.Address
	o.QueryTable("Address").RelatedSel("User").Filter("User__UserName", username).All(&addresses)
	if len(addresses) == 0 {
		this.Data["addresses"] = ""
	} else {
		this.Data["addresses"] = addresses
	}
	this.SetLayout("userLayout.html")
	this.TplName = "user_center_site.html"
}

func (this *UserController) HandleUserCenterSite() {
	beego.Info("收到处理请求了")
	receiver := this.GetString("receiver")
	addr := this.GetString("addr")
	zipCode := this.GetString("zipCode")
	phone := this.GetString("phone")
	beego.Info(receiver, addr, zipCode, phone)
	//校验数据
	if receiver == "" || addr == "" || zipCode == "" || phone == "" {
		beego.Info("数据不完整")
		this.Redirect("/fresh/userCenterSite", 302)
		return
	}
	o := orm.NewOrm()
	//添加默认地址之前需要把原来的默认地址修改为非默认地址
	var address models.Address
	address.Isdefault = true
	read := o.Read(&address, "Isdefault")
	if read == nil {
		address.Isdefault = false
		o.Update(&address)
	}
	user := models.User{UserName: this.GetUserName()}
	o.Read(&user, "UserName")
	newAddress := models.Address{Receiver: receiver, Zipcode: zipCode, Addr: addr, Phone: phone, Isdefault: true, User: &user}
	i, err := o.Insert(&newAddress)
	if err != nil {
		beego.Info("插入地址失败")
		this.Redirect("/fresh/userCenterSite", 302)
		return
	} else {
		beego.Info("i=============", i)
	}
	this.Redirect("/fresh/userCenterSite", 302)
}
