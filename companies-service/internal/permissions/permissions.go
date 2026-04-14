package permissions

type PermissionInfo struct {
	Mask        int64  `json:"mask"`
	Domain      string `json:"domain"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

const (
	UpdateCompanyPermission    int64 = 1 << 0
	GetCompanyByUuidPermission int64 = 1 << 1
	DeleteCompanyPermission    int64 = 1 << 2

	CreateInvitationPermission int64 = 1 << 0

	UpdateRoleMemberPermission int64 = 1 << 0
	SetOwnerPermission         int64 = 1 << 1
	DeleteMemberPermission     int64 = 1 << 2
	GetMembersPermission       int64 = 1 << 3

	CreateRolePermission    int64 = 1 << 0
	GetRolesPermission      int64 = 1 << 1
	GetRoleByUuidPermission int64 = 1 << 2
	UpdateRolePermission    int64 = 1 << 3
	DeleteRolePermission    int64 = 1 << 4

	UpdateAccessPermission int64 = 1 << 0
	GetAccountPermission   int64 = 1 << 1
)

const (
	CompanyPermissionDomain    string = "company"
	InvitationPermissionDomain string = "invitation"
	MembersPermissionDomain    string = "member"
	RolePermissionDomain       string = "role"
	PermissionDomain           string = "permission"
)

var AllDomainPermission = []string{
	CompanyPermissionDomain,
	InvitationPermissionDomain,
	MembersPermissionDomain,
	RolePermissionDomain,
	PermissionDomain,
}

var AllPermissions = []PermissionInfo{
	{
		Mask:        UpdateCompanyPermission,
		Domain:      CompanyPermissionDomain,
		Name:        "Управление компанией",
		Description: "Позволяет изменять название, описание и основные настройки компании",
	},
	{
		Mask:        GetCompanyByUuidPermission,
		Domain:      CompanyPermissionDomain,
		Name:        "Просмотр информации",
		Description: "Дает доступ к просмотру профиля компании",
	},
	{
		Mask:        DeleteCompanyPermission,
		Domain:      CompanyPermissionDomain,
		Name:        "Удаление компании",
		Description: "Критическое право: позволяет полностью удалить компанию и все связанные данные",
	},

	{
		Mask:        CreateInvitationPermission,
		Domain:      InvitationPermissionDomain,
		Name:        "Создание приглашений",
		Description: "Позволяет приглашать новых участников в компанию",
	},

	{
		Mask:        UpdateRoleMemberPermission,
		Domain:      MembersPermissionDomain,
		Name:        "Управление ролями участников",
		Description: "Позволяет изменять роль участника",
	},
	{
		Mask:        SetOwnerPermission,
		Domain:      MembersPermissionDomain,
		Name:        "Назначение владельца",
		Description: "Позволяет передавать права владельца компании другому участнику",
	},
	{
		Mask:        DeleteMemberPermission,
		Domain:      MembersPermissionDomain,
		Name:        "Исключение участников",
		Description: "Позволяет удалять участников из состава компании",
	},
	{
		Mask:        GetMembersPermission,
		Domain:      MembersPermissionDomain,
		Name:        "Просмотр списка участников",
		Description: "Дает доступ к списку всех участниеов",
	},

	{
		Mask:        CreateRolePermission,
		Domain:      RolePermissionDomain,
		Name:        "Создание ролей",
		Description: "Позволяет создавать новые шаблоны роли внутри компании",
	},
	{
		Mask:        GetRolesPermission,
		Domain:      RolePermissionDomain,
		Name:        "Просмотр ролей",
		Description: "Позволяет видеть список доступных ролей",
	},
	{
		Mask:        GetRoleByUuidPermission,
		Domain:      RolePermissionDomain,
		Name:        "Детали роли",
		Description: "Позволяет смотреть подробную информацию о роли",
	},
	{
		Mask:        UpdateRolePermission,
		Domain:      RolePermissionDomain,
		Name:        "Редактирование ролей",
		Description: "Позволяет изменять данные в существующих ролях",
	},
	{
		Mask:        DeleteRolePermission,
		Domain:      RolePermissionDomain,
		Name:        "Удаление ролей",
		Description: "Позволяет удалять роль",
	},

	{
		Mask:        UpdateAccessPermission,
		Domain:      PermissionDomain,
		Name:        "Выдать доступ",
		Description: "Позволяет выдать доступ к действиям на уровне компании",
	},
	{
		Mask:        GetAccountPermission,
		Domain:      PermissionDomain,
		Name:        "Получить список доступов",
		Description: "Позволяет получить список доступов к действиям на уровне компании у участника компании",
	},
}
