package example

import (
	"github.com/dipto-kainin/kai-rest-server/internal/kai"
)

func TEST_HANDELER_FUNC() kai.HandlerFunc {
	return func(c *kai.Context) {
		c.JSON(200, map[string]string{
			"message": "This is a test handler",
		})
	}
}