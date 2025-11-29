// write a test to check
package main

import (
	"github.com/dipto-kainin/kai-rest-server/internal/kai"
)

func main() {
	app := kai.NewApp()

	app.Use(kai.Logger(),kai.DamageControl())

	app.GET("/hello", func(c *kai.Context) {
		c.JSON(200, map[string]string{
			"message": "Hello, World!",
		})
	})
	app.Play(8000)
}