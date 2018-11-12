package controllers

import (
	"dailyfresh/constants"
	"dailyfresh/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/gomodule/redigo/redis"
	"strconv"
	"strings"
	"time"
)

type OrderController struct {
	BaseController
}

func (this *OrderController) ShowOrder() {
	skuids := this.GetStrings("skuid")
	if len(skuids) == 0 {
		this.Redirect("/fresh/cart", 302)
		return
	}
	o := orm.NewOrm()
	dbhost := beego.AppConfig.String("dbhost")
	conn, _ := redis.Dial("tcp", dbhost+":6379")
	defer conn.Close()
	username := this.GetUserName()
	user := models.User{UserName: username}
	o.Read(&user, "UserName")

	//定义条目切片
	cartItemSlice := make([]CartItem, 0)
	//定义总计和总数据
	totalCount := 0
	totalAmount := 0
	for _, skuid := range skuids {
		goodsId, _ := strconv.Atoi(skuid)
		goodsSKU := models.GoodsSKU{Id: goodsId}
		o.Read(&goodsSKU)
		//获取商品的数量
		count, _ := redis.Int(conn.Do("hget", constants.PRE_CART+strconv.Itoa(user.Id), skuid))
		amount := goodsSKU.Price * count
		cartItem := CartItem{&goodsSKU, count, amount}
		cartItemSlice = append(cartItemSlice, cartItem)
		totalCount += count
		totalAmount += amount
	}
	var addresses []models.Address
	o.QueryTable("Address").RelatedSel("User").Filter("User__Id", user.Id).All(&addresses)
	beego.Info("addresses====", addresses)
	this.Data["addresses"] = addresses
	this.Data["totalCount"] = totalCount
	this.Data["totalAmount"] = totalAmount
	transferPrice := 10
	this.Data["transferPrice"] = transferPrice
	this.Data["total"] = transferPrice + totalAmount
	this.Data["cartItemSlice"] = cartItemSlice
	this.Data["skuids"] = skuids
	this.TplName = "place_order.html"

}

func (this *OrderController) HandleAddOrder() {
	//{addrId, payId, skuids, totalAmount, totalCount, transferPrice, total};
	addrId, _ := this.GetInt("addrId")
	payId, _ := this.GetInt("payId")
	//totalAmount, _ := this.GetInt("totalAmount")
	totalCount, _ := this.GetInt("totalCount")
	transferPrice, _ := this.GetInt("transferPrice")
	total, _ := this.GetInt("total")
	skuidsString := this.GetString("skuids")
	skuids := strings.Split(skuidsString[1:len(skuidsString)-1], " ")
	defer this.ServeJSON()
	if len(skuids) == 0 {
		this.Data["json"] = &ResponseJSON{401, nil, "传参有误"}
		return
	}
	o := orm.NewOrm()
	o.Begin()
	username := this.GetUserName()
	user := models.User{UserName: username}
	o.Read(&user, "UserName")
	address := models.Address{Id: addrId}
	o.Read(&address)
	var orderInfo models.OrderInfo
	orderInfo.OrderId = time.Now().Format("20060102150405") + strconv.Itoa(user.Id)
	orderInfo.User = &user
	orderInfo.Orderstatus = 1
	orderInfo.PayMethod = payId
	orderInfo.TotalCount = totalCount
	orderInfo.TotalPrice = total
	orderInfo.TransitPrice = transferPrice
	orderInfo.Address = &address
	o.Insert(&orderInfo)

	dbhost := beego.AppConfig.String("dbhost")
	conn, _ := redis.Dial("tcp", dbhost+":6379")
	defer conn.Close()
	for _, skuid := range skuids {
		id, _ := strconv.Atoi(skuid)
		goodsSKU := models.GoodsSKU{Id: id}
		queryIndex := 3
		for queryIndex > 0 {
			o.Read(&goodsSKU)
			var orderGoods models.OrderGoods
			orderGoods.OrderInfo = &orderInfo
			orderGoods.GoodsSKU = &goodsSKU
			count, _ := redis.Int(conn.Do("hget", constants.PRE_CART+strconv.Itoa(user.Id), id))
			if count > goodsSKU.Stock {
				this.Data["json"] = &ResponseJSON{401, nil, "库存不足"}
				o.Rollback()
				return
			}
			preCount := goodsSKU.Stock
			orderGoods.Count = count
			orderGoods.Price = count * goodsSKU.Price
			_, err := o.Insert(&orderGoods)
			if err != nil {
				if queryIndex > 0 {
					queryIndex -= 1
					continue
				}
				this.Data["json"] = &ResponseJSON{401, nil, "订单提交失败"}
				o.Rollback()
				return
			}

			goodsSKU.Stock -= count
			goodsSKU.Sales += count
			updateCount, err := o.QueryTable("GoodsSKU").Filter("Id", goodsSKU.Id).Filter("Stock", preCount).Update(orm.Params{"Stock": goodsSKU.Stock, "Sales": goodsSKU.Sales})
			if updateCount == 0 || err != nil {
				if queryIndex > 0 {
					queryIndex -= 1
					continue
				}
				this.Data["json"] = &ResponseJSON{401, nil, "订单提交失败"}
				o.Rollback()
				return
			} else {
				conn.Do("hdel", constants.PRE_CART+strconv.Itoa(user.Id), id)
				break
			}
		}
	}
	o.Commit()
	this.Data["json"] = &ResponseJSON{200, nil, "订单处理成功"}
}

func (this *OrderController) HandlePay() {
	orderId := this.GetString("orderId")
	totalPrice := this.GetString("totalPrice")
	beego.Info("orderId========", orderId)
	beego.Info("totalPrice========", totalPrice)
}

func (this *OrderController) PayOk() {

}
