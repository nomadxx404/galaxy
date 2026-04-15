package returning

import "time"

type AggregatedMemberResponse struct {
	AccountUuid  string    `json:"account_uuid"`
	Name         string    `json:"name"`
	Nickname     string    `json:"nickname"`
	AvatarFileID int       `json:"avatar_file_id"`
	RoleID       int32     `json:"role_id"`
	RoleName     string    `json:"role_name"`
	RoleColor    string    `json:"role_color"`
	CreatedAt    time.Time `json:"joined_at"`
}
