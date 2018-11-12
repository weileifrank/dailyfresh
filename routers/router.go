package routers

import (
	"dailyfresh/constants"
	"dailyfresh/controllers"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
)

func init() {
	beego.InsertFilter("/fresh/*", beego.BeforeExec, filteFunc)
	beego.Router("/register", &controllers.UserController{}, "get:ShowRegister;post:HandleRegister")
	beego.Router("/active", &controllers.UserController{}, "get:ShowActive")
	beego.Router("/login", &controllers.UserController{}, "get:ShowLogin;post:HandleLogin")

	beego.Router("/", &controllers.GoodsController{}, "get:ShowIndex")
	beego.Router("/fresh/logout", &controllers.UserController{}, "get:ShowLogout")

	//用户中心信息页
	beego.Router("/fresh/userCenterInfo", &controllers.UserController{}, "get:ShowUserCenterInfo")
	//用户中心订单页
	beego.Router("/fresh/userCenterOrder", &controllers.UserController{}, "get:ShowUserCenterOrder")
	//用户中心地址页    命名即注释
	beego.Router("/fresh/userCenterSite", &controllers.UserController{}, "get:ShowUserCenterSite;post:HandleUserCenterSite")
	//商品详情展示
	beego.Router("/goodsDetail", &controllers.GoodsController{}, "get:ShowGoodsDetail")
	//展示商品列表
	beego.Router("/goodsList", &controllers.GoodsController{}, "get:ShowList")
	//添加购物车
	beego.Router("/fresh/addCart", &controllers.CartController{}, "post:HandleAddCart")
	//商品搜索
	beego.Router("/goodsSearch", &controllers.GoodsController{}, "post:HandleSearch")

	//添加购物车
	beego.Router("/fresh/addCart", &controllers.CartController{}, "post:HandleAddCart")
	//展示购物车页面
	beego.Router("/fresh/cart", &controllers.CartController{}, "get:ShowCart")

	//更新数量
	beego.Router("/fresh/UpdateCart", &controllers.CartController{}, "post:HandleUpdateCart")
	//根据id删除商品
	beego.Router("/fresh/deteteCartItem", &controllers.CartController{}, "post:HandleDeleteCartItem")

	//post展示订单页面
	beego.Router("/fresh/showOrder", &controllers.OrderController{}, "post:ShowOrder")

	//提交订单
	beego.Router("/fresh/addOrder", &controllers.OrderController{}, "post:HandleAddOrder")

	//处理支付
	beego.Router("/fresh/pay", &controllers.OrderController{}, "get:HandlePay")
	//支付成功
	beego.Router("/fresh/payok", &controllers.OrderController{}, "get:PayOk")
}

var filteFunc = func(ctx *context.Context) {
	session := ctx.Input.Session(constants.SESSION_KEY_USERNAME)
	if session == nil {
		beego.Info("被拦截了,需要先登录")
		ctx.Redirect(302, "/login")
	}
}
