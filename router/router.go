package router

import "github.com/kataras/iris/v12"

type News struct {
	Name string
}

func (news *News) Init(app *iris.Application) {
	root := app.Party(news.Name)
	root.Use()
}
