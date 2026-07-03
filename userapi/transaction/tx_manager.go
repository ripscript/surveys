package transaction

import "gorm.io/gorm"

type TxManager interface {
	Begin() *gorm.DB
}

type txManager struct {
	db *gorm.DB
}

func NewTxManager(db *gorm.DB) TxManager {
	return &txManager{db: db}
}

func (t *txManager) Begin() *gorm.DB {
	return t.db.Begin()
}
