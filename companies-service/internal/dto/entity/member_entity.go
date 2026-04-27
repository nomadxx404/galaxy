package entity

type CreateMemberEntity struct {
	CompanyUuid string
	AccountUuid string
	RoleID      int32
	IsOwner     bool
}
