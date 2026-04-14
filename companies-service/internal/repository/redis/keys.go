package rdb

import "fmt"

const (
	companyPrefix    = "company:info:"
	invitationPrefix = "invitation:"
	membersPrefix    = "members:info:"
	permissionPrefix = "permission:info:"
	userMaskPrefix   = "user:mask"
)

func GetCompanyKey(company_uuid string) string {
	return fmt.Sprintf("%s:%s", companyPrefix, company_uuid)
}

func GetInvitationKey(token string) string {
	return fmt.Sprintf("%s:%s", invitationPrefix, token)
}

func GetMembersKey(company_uuid string) string {
	return fmt.Sprintf("%s:%s", membersPrefix, company_uuid)
}

func GetPermissionKey(company_uuid, account_uuid string) string {
	return fmt.Sprintf("%s:%s:%s", permissionPrefix, company_uuid, account_uuid)
}

func GetUserMaskKey(companyUUID, accountUUID, domain string) string {
	return fmt.Sprintf("%s:%s:%s:%s", userMaskPrefix, companyUUID, accountUUID, domain)
}
