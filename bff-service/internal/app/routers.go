package app

import (
	"bff-service/pkg/middleware"

	scalargo "github.com/bdpiprava/scalar-go"
	"github.com/gin-gonic/gin"
)

func (a *App) RegisterRoutes() {
	docs := a.Router.Group("/ui")
	{
		docs.StaticFile("/openapi.json", "./static/swagger.json")
		docs.GET("/scalar", func(c *gin.Context) {
			html, err := scalargo.NewV2(
				scalargo.WithSpecURL("http://localhost:8000/ui/openapi.json"),
			)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			c.Data(200, "text/html; charset=utf-8", []byte(html))
		})
	}

	v1 := a.Router.Group("/api/ui")
	v1.Use(middleware.ContextInterceptor())
	{
		companies := v1.Group("/companies")
		{
			companies.GET("/:company_uuid/members", a.Handler.GetCompanyMembers)
		}
	}
}
