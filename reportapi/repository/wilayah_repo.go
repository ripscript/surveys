package repository

import (
	"backend/reportapi/models"
	"context"

	"gorm.io/gorm"
)

type WilayahRepository interface {
	GetRTInfo(ctx context.Context, rtID int64) (*models.DataRtModel, error)
	GetRTIDsUnderRW(ctx context.Context, rwID int64) ([]int64, error)
	GetRTIDsUnderKelurahan(ctx context.Context, kelurahanID int64) ([]int64, int, error)
	GetRTIDsUnderKecamatan(ctx context.Context, kecamatanID int64) ([]int64, int, int, error)
	GetRTIDs(ctx context.Context) ([]int64, int, int, int, error)

	GetChildrenOfRW(ctx context.Context, rwID int64) ([]WilayahChild, error)
	GetChildrenOfKelurahan(ctx context.Context, kelurahanID int64) ([]WilayahChild, error)
	GetChildrenOfKecamatan(ctx context.Context, kecamatanID int64) ([]WilayahChild, error)

	GetRWIDsUnderKelurahan(ctx context.Context, kelurahanID int64) ([]int64, error)
	GetKelurahanIDsUnderKecamatan(ctx context.Context, kecamatanID int64) ([]int64, error)
	GetAllKecamatanIDs(ctx context.Context) ([]int64, error)

	GetRTIDsByRWIDs(ctx context.Context, rwIDs []int64) ([]int64, error)
	GetRWIDsByKelurahanIDs(ctx context.Context, kelurahanIDs []int64) ([]int64, error)
	GetKelurahanIDsByKecamatanIDs(ctx context.Context, kecamatanIDs []int64) ([]int64, error)
	GetChildrenOfCity(ctx context.Context) ([]WilayahChild, error)

	GetSurveyIDsBySurveyor(ctx context.Context, respondentID int64) ([]int64, error)
	GetRTIDsBySurveyIDs(ctx context.Context, surveyIDs []int64) ([]int64, error)

	GetRTIDsGroupedBySurvey(ctx context.Context, surveyIDs []int64) (map[int64][]int64, error)
}

type wilayahRepository struct {
	dbSlave *gorm.DB
}

func NewWilayahRepository(dbSlave *gorm.DB) WilayahRepository {
	return &wilayahRepository{dbSlave: dbSlave}
}

func (r *wilayahRepository) GetRTInfo(ctx context.Context, rtID int64) (*models.DataRtModel, error) {
	var rt models.DataRtModel
	err := r.dbSlave.WithContext(ctx).First(&rt, rtID).Error
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *wilayahRepository) GetRTIDsUnderRW(ctx context.Context, rwID int64) ([]int64, error) {
	var ids []int64
	err := r.dbSlave.WithContext(ctx).
		Model(&models.DataRtModel{}).
		Where("rw_id = ?", rwID).
		Pluck("id", &ids).Error
	return ids, err
}

func (r *wilayahRepository) GetRTIDsUnderKelurahan(ctx context.Context, kelurahanID int64) ([]int64, int, error) {
	var rwIDs []int64
	if err := r.dbSlave.WithContext(ctx).
		Model(&models.DataRwModel{}).
		Where("kelurahan_id = ?", kelurahanID).
		Pluck("id", &rwIDs).Error; err != nil {
		return nil, 0, err
	}

	if len(rwIDs) == 0 {
		return []int64{}, 0, nil
	}

	var rtIDs []int64
	err := r.dbSlave.WithContext(ctx).
		Model(&models.DataRtModel{}).
		Where("rw_id IN ?", rwIDs).
		Pluck("id", &rtIDs).Error
	return rtIDs, len(rwIDs), err
}

func (r *wilayahRepository) GetRTIDsUnderKecamatan(ctx context.Context, kecamatanID int64) ([]int64, int, int, error) {
	var kelurahanIDs []int64
	if err := r.dbSlave.WithContext(ctx).
		Model(&models.KelurahanModel{}).
		Where("sub_district_id = ?", kecamatanID).
		Pluck("id", &kelurahanIDs).Error; err != nil {
		return nil, 0, 0, err
	}

	if len(kelurahanIDs) == 0 {
		return []int64{}, 0, 0, nil
	}

	var rwIDs []int64
	if err := r.dbSlave.WithContext(ctx).
		Model(&models.DataRwModel{}).
		Where("kelurahan_id IN ?", kelurahanIDs).
		Pluck("id", &rwIDs).Error; err != nil {
		return nil, 0, 0, err
	}

	var rtIDs []int64
	if len(rwIDs) > 0 {
		if err := r.dbSlave.WithContext(ctx).
			Model(&models.DataRtModel{}).
			Where("rw_id IN ?", rwIDs).
			Pluck("id", &rtIDs).Error; err != nil {
			return nil, 0, 0, err
		}
	}

	return rtIDs, len(rwIDs), len(kelurahanIDs), nil
}

func (r *wilayahRepository) GetRTIDs(ctx context.Context) ([]int64, int, int, int, error) {
	var kecamatanIDs []int64
	if err := r.dbSlave.WithContext(ctx).
		Model(&models.KecamatanModel{}).
		Pluck("id", &kecamatanIDs).Error; err != nil {
		return nil, 0, 0, 0, err
	}

	if len(kecamatanIDs) == 0 {
		return []int64{}, 0, 0, 0, nil
	}

	var kelurahanIDs []int64
	if err := r.dbSlave.WithContext(ctx).
		Model(&models.KelurahanModel{}).
		Where("sub_district_id IN ?", kecamatanIDs).
		Pluck("id", &kelurahanIDs).Error; err != nil {
		return nil, 0, 0, 0, err
	}

	if len(kelurahanIDs) == 0 {
		return []int64{}, 0, 0, 0, nil
	}

	var rwIDs []int64
	if err := r.dbSlave.WithContext(ctx).
		Model(&models.DataRwModel{}).
		Where("kelurahan_id IN ?", kelurahanIDs).
		Pluck("id", &rwIDs).Error; err != nil {
		return nil, 0, 0, 0, err
	}

	var rtIDs []int64
	if len(rwIDs) > 0 {
		if err := r.dbSlave.WithContext(ctx).
			Model(&models.DataRtModel{}).
			Where("rw_id IN ?", rwIDs).
			Pluck("id", &rtIDs).Error; err != nil {
			return nil, 0, 0, 0, err
		}
	}

	return rtIDs, len(rwIDs), len(kelurahanIDs), len(kecamatanIDs), nil
}

type WilayahChild struct {
	ID      int64
	Label   string
	GeoName *string
	RTIDs   []int64
}

func (r *wilayahRepository) GetChildrenOfRW(ctx context.Context, rwID int64) ([]WilayahChild, error) {
	var rts []models.DataRtModel
	err := r.dbSlave.WithContext(ctx).Where("rw_id = ?", rwID).Find(&rts).Error
	if err != nil {
		return nil, err
	}
	children := make([]WilayahChild, 0, len(rts))
	for _, rt := range rts {
		children = append(children, WilayahChild{
			ID:    int64(rt.ID),
			Label: "RT " + rt.NamaRt,
			RTIDs: []int64{int64(rt.ID)},
		})
	}
	return children, nil
}

func (r *wilayahRepository) GetChildrenOfKelurahan(ctx context.Context, kelurahanID int64) ([]WilayahChild, error) {
	var rws []models.DataRwModel
	if err := r.dbSlave.WithContext(ctx).Where("kelurahan_id = ?", kelurahanID).Find(&rws).Error; err != nil {
		return nil, err
	}
	children := make([]WilayahChild, 0, len(rws))
	for _, rw := range rws {
		var rtIDs []int64
		if err := r.dbSlave.WithContext(ctx).Model(&models.DataRtModel{}).Where("rw_id = ?", rw.ID).Pluck("id", &rtIDs).Error; err != nil {
			return nil, err
		}
		children = append(children, WilayahChild{ID: int64(rw.ID), Label: "RW " + rw.NamaRw, RTIDs: rtIDs})
	}
	return children, nil
}

func (r *wilayahRepository) GetChildrenOfKecamatan(ctx context.Context, kecamatanID int64) ([]WilayahChild, error) {
	var kelurahans []models.KelurahanModel
	if err := r.dbSlave.WithContext(ctx).Where("sub_district_id = ?", kecamatanID).Find(&kelurahans).Error; err != nil {
		return nil, err
	}
	children := make([]WilayahChild, 0, len(kelurahans))
	for _, kel := range kelurahans {
		var rwIDs []int64
		if err := r.dbSlave.WithContext(ctx).Model(&models.DataRwModel{}).Where("kelurahan_id = ?", kel.ID).Pluck("id", &rwIDs).Error; err != nil {
			return nil, err
		}
		var rtIDs []int64
		if len(rwIDs) > 0 {
			if err := r.dbSlave.WithContext(ctx).Model(&models.DataRtModel{}).Where("rw_id IN ?", rwIDs).Pluck("id", &rtIDs).Error; err != nil {
				return nil, err
			}
		}
		children = append(children, WilayahChild{
			ID:      int64(kel.ID),
			Label:   kel.VillageName,
			GeoName: kel.GeoName,
			RTIDs:   rtIDs,
		})
	}
	return children, nil
}

func (r *wilayahRepository) GetRWIDsUnderKelurahan(ctx context.Context, kelurahanID int64) ([]int64, error) {
	var ids []int64
	err := r.dbSlave.WithContext(ctx).Model(&models.DataRwModel{}).Where("kelurahan_id = ?", kelurahanID).Pluck("id", &ids).Error
	return ids, err
}

func (r *wilayahRepository) GetKelurahanIDsUnderKecamatan(ctx context.Context, kecamatanID int64) ([]int64, error) {
	var ids []int64
	err := r.dbSlave.WithContext(ctx).Model(&models.Kelurahan{}).Where("sub_district_id = ?", kecamatanID).Pluck("id", &ids).Error
	return ids, err
}

func (r *wilayahRepository) GetAllKecamatanIDs(ctx context.Context) ([]int64, error) {
	var ids []int64
	err := r.dbSlave.WithContext(ctx).Model(&models.Kecamatan{}).Pluck("id", &ids).Error
	return ids, err
}

func (r *wilayahRepository) GetRTIDsByRWIDs(ctx context.Context, rwIDs []int64) ([]int64, error) {
	if len(rwIDs) == 0 {
		return []int64{}, nil
	}
	var ids []int64
	err := r.dbSlave.WithContext(ctx).Model(&models.DataRtModel{}).Where("rw_id IN ?", rwIDs).Pluck("id", &ids).Error
	return ids, err
}

func (r *wilayahRepository) GetRWIDsByKelurahanIDs(ctx context.Context, kelurahanIDs []int64) ([]int64, error) {
	if len(kelurahanIDs) == 0 {
		return []int64{}, nil
	}
	var ids []int64
	err := r.dbSlave.WithContext(ctx).Model(&models.DataRwModel{}).Where("kelurahan_id IN ?", kelurahanIDs).Pluck("id", &ids).Error
	return ids, err
}

func (r *wilayahRepository) GetKelurahanIDsByKecamatanIDs(ctx context.Context, kecamatanIDs []int64) ([]int64, error) {
	if len(kecamatanIDs) == 0 {
		return []int64{}, nil
	}
	var ids []int64
	err := r.dbSlave.WithContext(ctx).Model(&models.Kelurahan{}).Where("sub_district_id IN ?", kecamatanIDs).Pluck("id", &ids).Error
	return ids, err
}

func (r *wilayahRepository) GetChildrenOfCity(ctx context.Context) ([]WilayahChild, error) {
	var kecamatans []models.KecamatanModel
	if err := r.dbSlave.WithContext(ctx).Find(&kecamatans).Error; err != nil {
		return nil, err
	}
	children := make([]WilayahChild, 0, len(kecamatans))
	for _, kec := range kecamatans {
		var kelurahanIDs []int64
		if err := r.dbSlave.WithContext(ctx).Model(&models.KelurahanModel{}).
			Where("sub_district_id = ?", kec.ID).Pluck("id", &kelurahanIDs).Error; err != nil {
			return nil, err
		}
		var rwIDs []int64
		if len(kelurahanIDs) > 0 {
			if err := r.dbSlave.WithContext(ctx).Model(&models.DataRwModel{}).
				Where("kelurahan_id IN ?", kelurahanIDs).Pluck("id", &rwIDs).Error; err != nil {
				return nil, err
			}
		}
		var rtIDs []int64
		if len(rwIDs) > 0 {
			if err := r.dbSlave.WithContext(ctx).Model(&models.DataRtModel{}).
				Where("rw_id IN ?", rwIDs).Pluck("id", &rtIDs).Error; err != nil {
				return nil, err
			}
		}
		children = append(children, WilayahChild{
			ID:      kec.ID,
			Label:   kec.SubDistrictName,
			GeoName: kec.GeoName,
			RTIDs:   rtIDs,
		})
	}
	return children, nil
}

func (r *wilayahRepository) GetSurveyIDsBySurveyor(ctx context.Context, respondentID int64) ([]int64, error) {
	var ids []int64
	err := r.dbSlave.WithContext(ctx).
		Table("survey__surveyors").
		Where("respondent_id = ?", respondentID).
		Pluck("survey_id", &ids).Error
	return ids, err
}

func (r *wilayahRepository) GetRTIDsBySurveyIDs(ctx context.Context, surveyIDs []int64) ([]int64, error) {
	if len(surveyIDs) == 0 {
		return []int64{}, nil
	}
	var ids []int64
	err := r.dbSlave.WithContext(ctx).
		Table("survey_respondents sr").
		Joins("JOIN respondents r ON r.id = sr.respondent_id").
		Where("sr.survey_id IN ?", surveyIDs).
		Where("r.rt_id IS NOT NULL").
		Distinct().
		Pluck("r.rt_id", &ids).Error
	return ids, err
}

func (r *wilayahRepository) GetRTIDsGroupedBySurvey(ctx context.Context, surveyIDs []int64) (map[int64][]int64, error) {
	result := make(map[int64][]int64, len(surveyIDs))
	if len(surveyIDs) == 0 {
		return result, nil
	}

	type row struct {
		SurveyID int64 `gorm:"column:survey_id"`
		RTID     int64 `gorm:"column:rt_id"`
	}
	var rows []row
	err := r.dbSlave.WithContext(ctx).
		Table("survey_respondents sr").
		Joins("JOIN respondents resp ON resp.id = sr.respondent_id").
		Select("DISTINCT sr.survey_id AS survey_id, resp.rt_id AS rt_id").
		Where("sr.survey_id IN ?", surveyIDs).
		Where("resp.rt_id IS NOT NULL").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, rw := range rows {
		result[rw.SurveyID] = append(result[rw.SurveyID], rw.RTID)
	}
	return result, nil
}
