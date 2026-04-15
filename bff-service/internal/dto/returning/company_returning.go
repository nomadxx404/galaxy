package returning

import "time"

type GetMembersResponse struct {
	AccountUuid string    `json:"account_uuid"`
	RoleID      int32     `json:"role_id"`
	CreatedAt   time.Time `json:"created_at"`
	RoleName    string    `json:"role_name"`
	RoleColor   string    `json:"role_color"`
}
