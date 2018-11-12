package controllers

import (
	"dailyfresh/constants"
	"dailyfresh/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/gomodule/redigo/redis"
	"strconv"
)

type CartController struct {
	BaseController
}

type CartItem struct {
	GoodsSKU *models.GoodsSKU
	Count    int
	Amount   int
}

func (this *CartController) HandleAddCart() {
	skuid, err1 := this.GetInt("skuid")
	count, err2 := this.GetInt("count")
	defer this.ServeJSON()
	if err1 != nil || err2 != nil {
		this.Data["json"] = &ResponseJSON{401, nil, "传递参数有误"}
		return
	}
	userName := this.GetUserName()
	if userName == "" {
		this.Data["json"] = &ResponseJSON{401, nil, "用户尚未登录"}
		return
	}
	o := orm.NewOrm()
	user := models.User{UserName: userName}
	o.Read(&user, "UserName")
	dbhost := beego.AppConfig.String("dbhost")
	conn, e := redis.Dial("tcp", dbhost+":6379")
	if e != nil {
		this.Data["json"] = &ResponseJSON{401, nil, "redis数据库连接错误"}
		return
	}
	defer conn.Close()
	//先后去该商品之前的数量,然后累计后再存储
	preCount, _ := redis.Int(conn.Do("hget", constants.PRE_CART+strconv.Itoa(user.Id), skuid))
	conn.Do("hset", constants.PRE_CART+strconv.Itoa(user.Id), skuid, count+preCount)
	//获取商品条目数
	reply, err := conn.Do("hlen", constants.PRE_CART+strconv.Itoa(user.Id))
	cartCount, _ := redis.Int(reply, err)
	datamap := make(map[string]interface{})
	datamap["cartCount"] = cartCount
	datamap["username"] = userName
	this.Data["json"] = &ResponseJSON{200, datamap, "商品添加成功"}
}

func (this *CartController) ShowCart() {
	username := this.GetUserName()
	dbhost := beego.AppConfig.String("dbhost")
	conn, _ := redis.Dial("tcp", dbhost+":6379")
	defer conn.Close()
	o := orm.NewOrm()
	user := models.User{UserName: username}
	o.Read(&user, "UserName")
	//定义条目切片
	cartItemSlice := make([]CartItem, 0)
	//定义总计和总数据
	totalCount := 0
	totalAmount := 0
	//获取所有的商品id及对应的数量
	reply, err := conn.Do("hgetall", constants.PRE_CART+strconv.Itoa(user.Id))
	goodsMap, _ := redis.IntMap(reply, err)
	for skuid, count := range goodsMap {
		goodsId, _ := strconv.Atoi(skuid)
		goodsSKU := models.GoodsSKU{Id: goodsId}
		o.Read(&goodsSKU)
		amount := goodsSKU.Price * count
		cartItem := CartItem{&goodsSKU, count, amount}
		cartItemSlice = append(cartItemSlice, cartItem)
		totalCount += count
		totalAmount += amount
	}
	this.Data["totalCount"] = totalCount
	this.Data["totalAmount"] = totalAmount
	this.Data["cartItemSlice"] = cartItemSlice
	this.TplName = "cart.html"
}

func (this *CartController) HandleUpdateCart() {
	skuid, err1 := this.GetInt("skuid")
	count, err2 := this.GetInt("count")
	defer this.ServeJSON()
	if err1 != nil || err2 != nil {
		this.Data["json"] = &ResponseJSON{401, nil, "传递参数有误"}
		return
	}
	userName := this.GetUserName()
	if userName == "" {
		this.Data["json"] = &ResponseJSON{401, nil, "用户尚未登录"}
		return
	}
	o := orm.NewOrm()
	user := models.User{UserName: userName}
	o.Read(&user, "UserName")
	dbhost := beego.AppConfig.String("dbhost")
	conn, e := redis.Dial("tcp", dbhost+":6379")
	if e != nil {
		this.Data["json"] = &ResponseJSON{401, nil, "redis数据库连接错误"}
		return
	}
	defer conn.Close()
	//修改商品数量
	_, err := conn.Do("hset", constants.PRE_CART+strconv.Itoa(user.Id), skuid, count)
	if err != nil {
		this.Data["json"] = &ResponseJSON{401, nil, "商品更新失败"}
		return
	}
	this.Data["json"] = &ResponseJSON{200, nil, "商品修改成功"}
}

func (this *CartController) HandleDeleteCartItem() {
	defer this.ServeJSON()
	skuid, _ := this.GetInt("skuid")
	username := this.GetUserName()
	user := models.User{UserName: username}
	o := orm.NewOrm()
	o.Read(&user, "UserName")
	dbhost := beego.AppConfig.String("dbhost")
	conn, _ := redis.Dial("tcp", dbhost+":6379")
	defer conn.Close()
	_, err := conn.Do("hdel", constants.PRE_CART+strconv.Itoa(user.Id), skuid)
	if err != nil {
		this.Data["json"] = &ResponseJSON{401, nil, "商品删除失败"}
		return
	}
	this.Data["json"] = &ResponseJSON{200, nil, "商品删除成功"}
}
