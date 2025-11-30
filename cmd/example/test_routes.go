package example

import (
	"github.com/dipto-kainin/kai-rest-server/internal/kai"
)

func TEST_ROUTES(app *kai.App) {
	app.GET("/test", TEST_HANDELER_FUNC())
}

func TEST_ROUTES1(app *kai.App) {
	app.GET("/test1", TEST_HANDELER_FUNC1())
}