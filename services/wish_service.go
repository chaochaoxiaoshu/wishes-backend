package services

import (
	"strconv"

	"gorm.io/gorm"

	"wishes/models"
)

type WishService struct {
	db *gorm.DB
}

func NewWishService(db *gorm.DB) *WishService {
	return &WishService{
		db: db,
	}
}

// WishResponse 包含心愿信息和附加字段
type WishResponse struct {
	models.Wish
	IsDone bool `json:"isDone"`
}

func (s *WishService) GetWishes(filters map[string]any) ([]WishResponse, int64, error) {
	query := s.db.Model(&models.Wish{})

	if content, ok := filters["content"].(string); ok && content != "" {
		query = query.Where("content LIKE ?", "%"+content+"%")
	}

	// 处理完成状态过滤
	if isDoneStr, ok := filters["isDone"].(string); ok && isDoneStr != "" {
		if isBool, err := strconv.ParseBool(isDoneStr); err == nil {
			if isBool {
				// 已认领：active_record_id 不为空
				query = query.Where("active_record_id IS NOT NULL")
			} else {
				// 可认领：active_record_id 为空
				query = query.Where("active_record_id IS NULL")
			}
		}
	}

	// 处理公开状态过滤
	if isPublishedStr, ok := filters["isPublished"].(string); ok && isPublishedStr != "" {
		if isBool, err := strconv.ParseBool(isPublishedStr); err == nil {
			query = query.Where("is_published = ?", isBool)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	pageIndex := filters["pageIndex"].(int)
	pageSize := filters["pageSize"].(int)
	offset := (pageIndex - 1) * pageSize

	var wishes []models.Wish
	if err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&wishes).Error; err != nil {
		return nil, 0, err
	}

	// 转换为带有 isDone 字段的响应结构体
	wishResponses := make([]WishResponse, len(wishes))
	for i, wish := range wishes {
		wishResponses[i] = WishResponse{
			Wish:   wish,
			IsDone: wish.ActiveRecordID != nil,
		}
	}

	return wishResponses, total, nil
}

func (s *WishService) CreateWish(wish *models.Wish) error {
	return s.db.Create(wish).Error
}

func (s *WishService) GetWishByID(id uint) (*models.Wish, error) {
	var wish models.Wish
	if err := s.db.First(&wish, id).Error; err != nil {
		return nil, err
	}
	return &wish, nil
}

func (s *WishService) UpdateWish(wish *models.Wish) error {
	return s.db.Save(wish).Error
}

func (s *WishService) DeleteWish(id uint) error {
	return s.db.Delete(&models.Wish{}, id).Error
}

func (s *WishService) GetWishesByDonorID(donorID uint, pageIndex, pageSize int) ([]models.Wish, int64, error) {
	query := s.db.Model(&models.Wish{}).Where("donor_id = ?", donorID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pageIndex - 1) * pageSize

	var wishes []models.Wish
	if err := query.Limit(pageSize).Offset(offset).Find(&wishes).Error; err != nil {
		return nil, 0, err
	}

	return wishes, total, nil
}

func (s *WishService) BatchCreateWishes(wishes []*models.Wish) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		for _, wish := range wishes {
			if err := tx.Create(wish).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
