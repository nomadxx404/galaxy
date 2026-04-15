package request

type AccountBatchRequest struct {
	AccountUuids []string `json:"account_uuids"`
}
