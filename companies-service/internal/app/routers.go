package app

import (
	"companies-service/internal/permissions"
	"companies-service/pkg/middleware"

	scalargo "github.com/bdpiprava/scalar-go"
	"github.com/gin-gonic/gin"
)

func (a *App) RegisterRoutes() {

	a.Router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/health"},
	}))
	a.Router.Use(gin.Recovery())

	docs := a.Router.Group("/companies")
	{
		docs.StaticFile("/openapi.json", "./static/swagger.json")
		docs.GET("/scalar", func(c *gin.Context) {
			html, err := scalargo.NewV2(
				scalargo.WithSpecURL("http://localhost:8000/companies/openapi.json"),
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
		companies := v1.Group("/companies")
		{
			companies.POST("", a.CompanyHandler.CreateCompany)
			companies.GET("", a.CompanyHandler.GetCompanies)

			companies.GET(
				"/:company_uuid",
				middleware.PermissionMiddleware(
					a.Store,
					a.Rdb,
					permissions.CompanyPermissionDomain,
					permissions.GetCompanyByUuidPermission,
				),
				a.CompanyHandler.GetCompanyByUuid,
			)
			companies.DELETE(
				"/:company_uuid",
				middleware.PermissionMiddleware(
					a.Store,
					a.Rdb,
					permissions.CompanyPermissionDomain,
					permissions.DeleteCompanyPermission,
				),
				a.CompanyHandler.DeleteCompany,
			)
			companies.PATCH("/:company_uuid",
				middleware.PermissionMiddleware(
					a.Store,
					a.Rdb,
					permissions.CompanyPermissionDomain,
					permissions.UpdateCompanyPermission,
				),
				a.CompanyHandler.UpdateCompany,
			)
		}

		roles := v1.Group("/companies/:company_uuid/roles")
		{
			roles.POST(
				"",
				middleware.PermissionMiddleware(
					a.Store,
					a.Rdb,
					permissions.RolePermissionDomain,
					permissions.CreateRolePermission,
				),
				a.RoleHandler.CreateRole,
			)
			roles.GET(
				"",
				middleware.PermissionMiddleware(
					a.Store,
					a.Rdb,
					permissions.RolePermissionDomain,
					permissions.GetRolesPermission,
				),
				a.RoleHandler.GetRoles,
			)
			roles.GET(
				"/:role_id",
				middleware.PermissionMiddleware(
					a.Store,
					a.Rdb,
					permissions.RolePermissionDomain,
					permissions.GetRoleByUuidPermission,
				),
				a.RoleHandler.GetRoleByUuid,
			)
			roles.PATCH(
				"/:role_id",
				middleware.PermissionMiddleware(
					a.Store,
					a.Rdb,
					permissions.RolePermissionDomain,
					permissions.UpdateRolePermission,
				),
				a.RoleHandler.UpdateRole,
			)
			roles.DELETE(
				"/:role_id",
				middleware.PermissionMiddleware(
					a.Store,
					a.Rdb,
					permissions.RolePermissionDomain,
					permissions.DeleteRolePermission,
				),
				a.RoleHandler.DeleteRole,
			)
		}

		invitation := v1.Group("/companies/:company_uuid/invitations")
		{
			invitation.POST(
				"",
				middleware.PermissionMiddleware(
					a.Store,
					a.Rdb,
					permissions.InvitationPermissionDomain,
					permissions.CreateInvitationPermission,
				),
				a.InvitationHandler.CreateInvitation,
			)
		}

		acceptInvitation := v1.Group("/invitations/:tokenUrl/accept")
		{
			acceptInvitation.POST("", a.InvitationHandler.AcceptInvitation)
		}

		members := v1.Group("/companies/:company_uuid/members")
		{
			members.POST(
				"/:account_uuid/role",
				middleware.PermissionMiddleware(
					a.Store,
					a.Rdb,
					permissions.MembersPermissionDomain,
					permissions.UpdateRoleMemberPermission,
				),
				a.MemberHandler.UpdateRoleMember,
			)
			members.POST(
				"/:account_uuid/set-owner",
				middleware.PermissionMiddleware(
					a.Store,
					a.Rdb,
					permissions.MembersPermissionDomain,
					permissions.SetOwnerPermission,
				),
				a.MemberHandler.SetOwner,
			)
			members.DELETE(
				"/:account_uuid",
				middleware.PermissionMiddleware(
					a.Store,
					a.Rdb,
					permissions.MembersPermissionDomain,
					permissions.DeleteMemberPermission,
				),
				a.MemberHandler.DeleteMember,
			)
			members.GET(
				"",
				middleware.PermissionMiddleware(
					a.Store,
					a.Rdb,
					permissions.MembersPermissionDomain,
					permissions.GetMembersPermission,
				),
				a.MemberHandler.GetMembers)
		}

		permissionsAccess := v1.Group("/companies/permissions")
		{
			permissionsAccess.GET("", a.PermissionHandler.GetPermission)
		}

		permissionsUpdate := v1.Group("/companies/:company_uuid/members/:account_uuid/permissions")
		{
			permissionsUpdate.PUT(
				"",
				middleware.PermissionMiddleware(
					a.Store,
					a.Rdb,
					permissions.PermissionDomain,
					permissions.UpdateAccessPermission,
				),
				a.PermissionHandler.UpdatePermissions,
			)
			permissionsUpdate.GET(
				"",
				middleware.PermissionMiddleware(
					a.Store,
					a.Rdb,
					permissions.PermissionDomain,
					permissions.GetAccountPermission,
				),
				a.PermissionHandler.GetAccountPermission,
			)
		}
	}
}
