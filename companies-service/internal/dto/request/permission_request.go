package request

type UpdatePermissionsRequest struct {
	Permissions []DomainPermission `json:"permissions"`
}

type DomainPermission struct {
	Domain string `json:"domain"`
	Mask   int64  `json:"mask"`
}
