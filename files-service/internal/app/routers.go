package app

import (
	"files-service/pkg/middleware"

	scalargo "github.com/bdpiprava/scalar-go"
	"github.com/gin-gonic/gin"
)

func (a *App) RegisterRoutes() {

	a.Router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/health"},
	}))
	a.Router.Use(gin.Recovery())

	docs := a.Router.Group("/files")
	{
		docs.StaticFile("/openapi.json", "./static/swagger.json")
		docs.GET("/scalar", func(c *gin.Context) {
			html, err := scalargo.NewV2(
				scalargo.WithSpecURL("http://localhost:8000/files/openapi.json"),
			)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			c.Data(200, "text/html; charset=utf-8", []byte(html))
		})
	}

	v1 := a.Router.Group("/api")
	v1.Use(middleware.ContextMiddleware())
	{
		files := v1.Group("/files")
		{
			files.POST("", a.FileHandler.CreateFile)
			files.POST("/:entity_id/search", a.FileHandler.GetFiles)
			files.DELETE("/:entity_id/:file_id", a.FileHandler.DeleteFiles)
			files.GET("/download/:file_id", a.FileHandler.DownloadFile)
		}
	}
}
