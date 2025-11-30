package example

import (
	"github.com/dipto-kainin/kai-rest-server/kai"
)

func TEST_HANDELER_FUNC1() kai.HandlerFunc {
	return func(c *kai.Context) {
		c.JSON(200, map[string]string{
			"message": "This is a test handler 1",
		})
	}
}