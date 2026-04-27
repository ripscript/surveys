package enums

type WilayahID int

const (
	KECAMATAN WilayahID = 5
	KELURAHAN WilayahID = 4
	RW        WilayahID = 3
	RT        WilayahID = 2
)

func (t WilayahID) IsWilayahValid() bool {
	switch t {
	case KECAMATAN, KELURAHAN, RW, RT:
		return true
	}
	return false
}

func GetWilayahFromRole(role *int64) (int64, bool) {
	wilayahID := WilayahID(*role)
	switch wilayahID {
	case KECAMATAN:
		return int64(KECAMATAN), true
	case KELURAHAN:
		return int64(KELURAHAN), true
	case RW:
		return int64(RW), true
	case RT:
		return int64(RT), true
	}
	return 0, false
}

func (t WilayahID) Label() string {
	switch t {
	case KECAMATAN:
		return "Kecamatan"
	case KELURAHAN:
		return "Kelurahan"
	case RW:
		return "RW"
	case RT:
		return "RT"
	}
	return ""
}
