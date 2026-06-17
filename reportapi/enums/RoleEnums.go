package enums

type RoleID int

const (
	ROLE_RT              RoleID = 2
	ROLE_RW              RoleID = 3
	ROLE_KELURAHAN       RoleID = 4
	ROLE_KECAMATAN       RoleID = 5
	ROLE_PEMERINTAH_KOTA RoleID = 6
	ROLE_ADMIN           RoleID = 7
	ROLE_SURVEYOR        RoleID = 8
	ROLE_WALIKOTA        RoleID = 9
)

// Label mengembalikan nama tampilan dari RoleID.
// Mengembalikan string kosong jika ID tidak dikenali.
func (r RoleID) Label() string {
	switch r {
	case ROLE_RT:
		return "RT"
	case ROLE_RW:
		return "RW"
	case ROLE_KELURAHAN:
		return "Kelurahan"
	case ROLE_KECAMATAN:
		return "Kecamatan"
	case ROLE_PEMERINTAH_KOTA:
		return "Pemerintah Kota"
	case ROLE_ADMIN:
		return "Admin"
	case ROLE_SURVEYOR:
		return "Surveyor"
	case ROLE_WALIKOTA:
		return "Walikota"
	default:
		return ""
	}
}

// Contoh penggunaan:
//   id := RoleID(4)
//   fmt.Println(id.Label()) // "Kelurahan"
//
//   // Dari database:
//   var roleID RoleID = RoleID(dbRow.RoleID)
//   label := roleID.Label()
