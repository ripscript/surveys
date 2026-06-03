package payloads

type LogPayload struct {
	// ID          uint64 `gorm:"primaryKey" json:"id"`
	ServiceName string  `gorm:"size:100;not null" json:"serviceName"`
	Module      string  `gorm:"size:100;not null" json:"module"`
	Action      string  `gorm:"size:50;not null" json:"action"`
	UserID      float64 `gorm:"size:255" json:"userId"`
	UserName    string  `gorm:"size:255" json:"userName"`
	EntityID    string  `gorm:"size:255" json:"entityID"`
	Description string  `gorm:"type:text" json:"description"`
}
