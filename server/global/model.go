package global

import "time"

// OPTIMUS_MODEL shares existing primary key and timestamps without adding
// deletion behavior. Models retain their own explicit deletion fields.
type OPTIMUS_MODEL struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
