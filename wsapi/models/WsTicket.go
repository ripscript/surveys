package models

import "time"

// WsTicket represents a WebSocket ticket for user authentication and session management.

type WsTicket struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Ticket    string    `gorm:"column:ticket;primaryKey;size:64" json:"ticket"`
	UserID    int32     `gorm:"column:user_id;not null;index" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID;references:ID" json:"-"`
	ExpiresAt time.Time `gorm:"column:expires_at;not null;index" json:"expires_at"`
	CreatedAt time.Time `gorm:"column:created_at;not null;autoCreateTime" json:"created_at"`
}

func (WsTicket) TableName() string {
	return "ws_tickets"
}
