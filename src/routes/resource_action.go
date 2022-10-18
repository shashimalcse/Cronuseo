package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/shashimalcse/Cronuseo/controllers"
)

func ResourceActionRoutes(router *gin.Engine) {

	resourceActionRouter := router.Group("/resource_actions")
	resourceActionRouter.GET("/:res_id", controllers.GetResourceActions)
	resourceActionRouter.GET("/:res_id/:id", controllers.GetResourceAction)
	resourceActionRouter.POST("/:res_id", controllers.CreateResourceAction)
	resourceActionRouter.DELETE("/:res_id/:id", controllers.DeleteResourceAction)
	resourceActionRouter.PUT("/:res_id/:id", controllers.UpdateResourceAction)
}