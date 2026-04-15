package returning

type AccountBatchResponse struct {
	AccountUuid  string `json:"account_uuid"`
	Name         string `json:"name"`
	Nickname     string `json:"nickname"`
	AvatarFileID int    `json:"avatar_file_id"`
}
