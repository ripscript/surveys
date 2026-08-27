package enums

type DashboardStatus string

const (
	StatusBaik           DashboardStatus = "baik"
	StatusCukup          DashboardStatus = "cukup"
	StatusPerluPerhatian DashboardStatus = "perlu_perhatian"
	StatusPrioritas      DashboardStatus = "prioritas"
)

type DashboardStatusInfo struct {
	Key      DashboardStatus `json:"key"`
	Label    string          `json:"label"`
	ColorHex string          `json:"color_hex"`
	Severity int             `json:"severity"`
}

var DashboardStatusCatalog = []DashboardStatusInfo{
	{StatusBaik, "Baik", "#22C55E", 1},
	{StatusCukup, "Cukup", "#EAB308", 2},
	{StatusPerluPerhatian, "Perlu Perhatian", "#F97316", 3},
	{StatusPrioritas, "Prioritas", "#EF4444", 4},
}

var dashboardStatusIndex = buildDashboardStatusIndex()

func buildDashboardStatusIndex() map[DashboardStatus]DashboardStatusInfo {
	m := make(map[DashboardStatus]DashboardStatusInfo, len(DashboardStatusCatalog))
	for _, s := range DashboardStatusCatalog {
		m[s.Key] = s
	}
	return m
}

func IsValidDashboardStatus(key string) bool {
	_, ok := dashboardStatusIndex[DashboardStatus(key)]
	return ok
}

func GetDashboardStatusInfo(key string) (DashboardStatusInfo, bool) {
	info, ok := dashboardStatusIndex[DashboardStatus(key)]
	return info, ok
}

func DashboardStatusSeverity(key string) int {
	if info, ok := dashboardStatusIndex[DashboardStatus(key)]; ok {
		return info.Severity
	}
	return 0
}

func DashboardStatusLabel(key string) string {
	if info, ok := dashboardStatusIndex[DashboardStatus(key)]; ok {
		return info.Label
	}
	return ""
}
