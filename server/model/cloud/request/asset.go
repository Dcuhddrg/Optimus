package request

import commonRequest "optimus/server/model/common/request"

// AssetSearch contains transport filters only. Authorized project IDs are
// supplied separately by middleware, never accepted as request DTO fields.
type AssetSearch struct {
	commonRequest.PageInfo
	AccountID    *uint    `json:"accountId" form:"accountId"`
	Provider     *string  `json:"provider" form:"provider"`
	Region       *string  `json:"region" form:"region"`
	ResourceType *string  `json:"resourceType" form:"resourceType"`
	Status       *string  `json:"status" form:"status"`
	Keyword      *string  `json:"keyword" form:"keyword"`
	Tags         []string `json:"tag" form:"tag"`
}
