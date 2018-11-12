package main

import (
	_ "dailyfresh/models"
	_ "dailyfresh/routers"
	"github.com/astaxie/beego"
)

func main() {
	beego.Run()
}
