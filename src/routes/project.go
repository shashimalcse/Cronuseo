package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/shashimalcse/Cronuseo/controllers"
)

func ProjectRoutes(router *gin.Engine) {

	projectRouter := router.Group("/projects")

	projectRouter.GET("/:org_id", controllers.GetProjects)
	projectRouter.GET("/:org_id/:id", controllers.GetProject)
	projectRouter.POST("/:org_id", controllers.CreateProject)
	projectRouter.DELETE("/:org_id/:id", controllers.DeleteProject)
	projectRouter.PUT("/:org_id/:id", controllers.UpdateProject)
}