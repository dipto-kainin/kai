// write a test to check
package main

import (
	"github.com/dipto-kainin/kai"
	"github.com/dipto-kainin/kai/cmd/example"
)

func main() {
	app := kai.NewApp()

	app.Use(kai.Logger(),kai.DamageControl())

	app.UseRoutes(example.TEST_ROUTES, example.TEST_ROUTES1)

	app.GET("/hello", func(c *kai.Context) {
		c.JSON(200, map[string]string{
			"message": "Hello, World!",
		})
	})
	app.Play(8000)
}