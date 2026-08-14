package request

type DownloadTempLaporanRequest struct {
	Path string `json:"path" validate:"required"`
}
