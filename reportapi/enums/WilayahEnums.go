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

func IsWilayahExist(w WilayahID) bool {
	switch w {
	case KECAMATAN, KELURAHAN, RW, RT:
		return true
	}
	return false
}

func WilayahIDToInt64(w WilayahID) int64 {
	switch w {
	case KECAMATAN:
		return 5
	case KELURAHAN:
		return 4
	case RW:
		return 3
	case RT:
		return 2
	default:
		return 0
	}
}
