package models

import "time"

type Menu struct {
	ID        int    `json:"id" gorm:"primaryKey;autoIncrement"`
	MenuName  string `json:"menu_name" gorm:"type:varchar(255);not null"`
	Key       string `json:"key" gorm:"uniqueIndex"`
	Icon      string `json:"icon"`
	ParentID  *int   `json:"parent_id" gorm:"default:null"`
	Parent    *Menu  `gorm:"foreignKey:ParentID;constraint:OnDelete:CASCADE;"`
	Children  []Menu `gorm:"foreignKey:ParentID"`
	SortOrder int    `json:"sort_order" gorm:"default:0"`
	CreatedBy uint
	UpdatedBy uint
	DeletedBy *uint
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type MenuPermission struct {
	ID           int   `json:"id" gorm:"primaryKey;autoIncrement"`
	MenuID       int   `gorm:"uniqueIndex:idx_role_menu"`
	Menu         Menu  `gorm:"foreignKey:MenuID;constraint:OnDelete:CASCADE;"`
	RoleID       int   `gorm:"uniqueIndex:idx_role_menu"`
	ViewAction   bool  `json:"view_action"`
	CreateAction bool  `json:"create_action"`
	UpdateAction bool  `json:"update_action"`
	DeleteAction bool  `json:"delete_action"`
	ShowInMenu   *bool `json:"show_in_menu" gorm:"default:true"`
}

type MenuPermissionRole struct {
	ChildMenu []ChildMenu `json:"childMenu"`
	Icon      string      `json:"icon"`
	Key       string      `json:"key"`
	Title     string      `json:"title"`
}

type ChildMenu struct {
	ChildMenu []ChildMenu `json:"childMenu"`
	Icon      string      `json:"icon"`
	Key       string      `json:"key"`
	Title     string      `json:"title"`
}

type Role struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (Role) TableName() string {
	return "role"
}

type MenuIDs struct {
	Template           int
	ManajemenAlur      int
	Survey             int
	Beranda            int
	Pengaturan         int
	Laporan            int
	Monitoring         int
	Admin              int
	FormulirPertanyaan int
	Ucapan             int
	Surveys            int
	Hasil              int
	ManajemenPengguna  int
	ManajemenWilayah   int
	ManajemenCMS       int
	ManajemenArtikel   int
	Rating             int

	ManajemenMetrik int
	DaftarMetrik    int
	PemetaanMetrik  int

	Statistik             int
	AktifitasSurvey       int
	ProfilSaya            int
	Keluar                int
	ManajemenResponden    int
	ManajemenUser         int
	ManajemenBlokir       int
	ManajemenWilayahChild int
	ManajemenPejabat      int
	Artikel               int
	Promote               int
	Kategori              int
	MasterData            int
	ListSurvey            int
	Profile               int
	SurveyKewilayahan     int
	PengelolaSurvey       int
	MonitoringDanLaporan  int
	PengaturanAplikasi    int
	Dashboard             int
	DashboardUtama        int
	DashboardWilayah      int
	DashboardRT           int
	RiwayatDataDSS        int
}

type RoleIDs struct {
	Public   int
	Rt       int
	Rw       int
	Lurah    int
	Camat    int
	Pemkot   int
	Admin    int
	Surveyor int
	Walikota int
}
