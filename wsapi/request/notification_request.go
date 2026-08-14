package request

type ReportStatusNotification struct {
	UserID       int64   `json:"user_id" validate:"required"`
	LaporanID    int64   `json:"laporan_id" validate:"required"`
	Success      *bool   `json:"success"`
	DocumentID   *string `json:"document_id"`
	ErrorMessage *string `json:"error_message,omitempty"`
}
