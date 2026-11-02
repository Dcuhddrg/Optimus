package request

// PageInfo holds pagination shared by endpoints that actually return pages.
// Normalize preserves Optimus defaults: page 1, page size 20, maximum 200.
type PageInfo struct {
	Page     int `json:"page" form:"page"`
	PageSize int `json:"pageSize" form:"pageSize"`
}

// Normalize resets invalid sizes to the default instead of clamping them.
func (p *PageInfo) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 || p.PageSize > 200 {
		p.PageSize = 20
	}
}
