package routes

import (
	"github.com/gin-gonic/gin"
)

func HandleList(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "list",
	})
}
