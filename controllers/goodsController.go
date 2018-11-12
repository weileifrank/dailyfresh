package controllers

import (
	"dailyfresh/constants"
	"dailyfresh/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/gomodule/redigo/redis"
	"math"
	"strconv"
)

type GoodsController struct {
	BaseController
}

type IndexTypeGoodsItem struct {
	//GoodsTypes []models.GoodsType
	//IndexGoodsBanners []models.IndexGoodsBanner
	//IndexPromotionBanners []models.IndexPromotionBanner
	GoodsType                  models.GoodsType
	TextIndexTypeGoodsBanners  []models.IndexTypeGoodsBanner
	ImageIndexTypeGoodsBanners []models.IndexTypeGoodsBanner
}

//展示首页
func (this *GoodsController) ShowIndex() {
	this.GetUserName()

	o := orm.NewOrm()
	//获取类型数据
	var goodsTypes []models.GoodsType
	o.QueryTable("GoodsType").All(&goodsTypes)
	//获取轮播图数据
	var indexGoodsBanners []models.IndexGoodsBanner
	o.QueryTable("IndexGoodsBanner").OrderBy("Index").All(&indexGoodsBanners)
	//获取促销商品数据
	var indexPromotionBanners []models.IndexPromotionBanner
	o.QueryTable("IndexPromotionBanner").OrderBy("Index").All(&indexPromotionBanners)

	//获取首页类型分条目数据
	indexTypeGoodsItems := make([]IndexTypeGoodsItem, 0)
	for _, goodsType := range goodsTypes {
		var textIndexTypeGoodsBanners []models.IndexTypeGoodsBanner
		var imageIndexTypeGoodsBanners []models.IndexTypeGoodsBanner
		o.QueryTable("IndexTypeGoodsBanner").RelatedSel("GoodsType", "GoodsSKU").OrderBy("Index").Filter("GoodsType", &goodsType).Filter("DisplayType", 0).All(&textIndexTypeGoodsBanners)
		o.QueryTable("IndexTypeGoodsBanner").RelatedSel("GoodsType", "GoodsSKU").OrderBy("Index").Filter("GoodsType", &goodsType).Filter("DisplayType", 1).All(&imageIndexTypeGoodsBanners)
		indexTypeGoodsItem := IndexTypeGoodsItem{GoodsType: goodsType, TextIndexTypeGoodsBanners: textIndexTypeGoodsBanners, ImageIndexTypeGoodsBanners: imageIndexTypeGoodsBanners}
		indexTypeGoodsItems = append(indexTypeGoodsItems, indexTypeGoodsItem)
	}
	this.Data["goodsTypes"] = goodsTypes
	this.Data["indexGoodsBanners"] = indexGoodsBanners
	this.Data["indexPromotionBanners"] = indexPromotionBanners
	this.Data["list"] = indexTypeGoodsItems
	this.SetLayout("layout.html")
	this.TplName = "index.html"
}

//展示商品详情
func (this *GoodsController) ShowGoodsDetail() {
	id, _ := this.GetInt("id")
	o := orm.NewOrm()
	var goodsSKU models.GoodsSKU
	o.QueryTable("GoodsSKU").RelatedSel("Goods", "GoodsType").Filter("Id", id).One(&goodsSKU)

	//获取推荐新品的数据
	var recommandGoods []models.GoodsSKU
	o.QueryTable("GoodsSKU").RelatedSel("Goods", "GoodsType").Filter("GoodsType", goodsSKU.GoodsType).OrderBy("Time").Limit(3, 0).All(&recommandGoods)

	username := this.GetUserName()
	if username != "" {
		dbhost := beego.AppConfig.String("dbhost")
		conn, e := redis.Dial("tcp", dbhost+":6379")
		if e != nil {
			beego.Info(e)
			this.Redirect("/", 302)
			return
		}
		defer conn.Close()
		_, err2 := conn.Do("lrem", constants.PRE_SCAN+strconv.Itoa(this.GetUserId(o, username)), 0, id)
		if err2 != nil {
			beego.Info(err2)
			this.Redirect("/", 302)
			return
		}
		_, err := conn.Do("lpush", constants.PRE_SCAN+strconv.Itoa(this.GetUserId(o, username)), id)
		if err != nil {
			beego.Info(err)
			this.Redirect("/", 302)
			return
		} else {
			beego.Info("插入浏览数据成功")
		}
	} else {
		beego.Info("username为空")
	}
	this.Data["goodsSKU"] = goodsSKU
	this.Data["recommandGoods"] = recommandGoods
	this.SetLayout("layout.html")
	this.TplName = "detail.html"
}

func PageTool(pageCount, pageIndex int) []int {
	var pages []int
	if pageCount <= 5 {
		for i := 1; i <= pageCount; i++ {
			pages = append(pages, i)
		}
	} else {
		if pageIndex <= 3 {
			pages = []int{1, 2, 3, 4, 5}
		} else if pageIndex >= pageCount-3 {
			pages = []int{pageCount - 4, pageCount - 3, pageCount - 2, pageCount - 1, pageCount}
		} else {
			pages = []int{pageCount - 2, pageCount - 1, pageCount, pageCount + 1, pageCount + 2}
		}
	}
	return pages
}

//展示商品列表
func (this *GoodsController) ShowList() {
	typeId, _ := this.GetInt("typeId")
	o := orm.NewOrm()
	//获取最早的商品
	var oldGoods []models.GoodsSKU
	o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id", typeId).OrderBy("-Time").Limit(2, 0).All(&oldGoods)
	this.Data["oldGoods"] = oldGoods

	//获取分页数据

	count, _ := o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id", typeId).Count()
	pageSize := 2
	pageCount := math.Ceil(float64(count) / float64(pageSize))
	pageIndex, erro := this.GetInt("pageIndex")
	if erro != nil {
		pageIndex = 1
	}
	pages := PageTool(int(pageCount), pageIndex)

	this.Data["pages"] = pages
	this.Data["typeId"] = typeId
	this.Data["pageIndex"] = pageIndex

	startIndex := (pageIndex - 1) * pageSize

	//判断首页和尾页
	prePageIndex := pageIndex - 1
	if prePageIndex < 1 {
		prePageIndex = 1
		this.Data["isFirstPage"] = true
	} else {
		this.Data["isFirstPage"] = false
	}

	this.Data["prePageIndex"] = prePageIndex
	nextPageIndex := pageIndex + 1
	if nextPageIndex > int(pageCount) {
		nextPageIndex = int(pageCount)
		this.Data["isLastPage"] = true
	} else {
		this.Data["isLastPage"] = false
	}
	this.Data["nextPageIndex"] = nextPageIndex
	//根据类型获取指定数据
	var goods []models.GoodsSKU
	sort := this.GetString("sort")
	if sort == "" {
		o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id", typeId).Limit(pageSize, startIndex).All(&goods)
		this.Data["sort"] = ""
	} else if sort == "price" {
		o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id", typeId).OrderBy("Price").Limit(pageSize, startIndex).All(&goods)
		this.Data["sort"] = "price"
	} else {
		o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id", typeId).OrderBy("Sales").Limit(pageSize, startIndex).All(&goods)
		this.Data["sort"] = "sale"
	}
	this.Data["goods"] = goods
	this.SetLayout("layout.html")
	this.TplName = "list.html"
}

func (this *GoodsController) HandleSearch() {
	goodsName := this.GetString("goodsName")
	o := orm.NewOrm()
	var goods []models.GoodsSKU
	if goodsName == "" {
		o.QueryTable("GoodsSKU").All(&goods)
	} else {
		o.QueryTable("GoodsSKU").Filter("Name__icontains", goodsName).All(&goods)
	}
	this.Data["goods"] = goods
	this.SetLayout("layout.html")
	this.TplName = "search.html"
}
