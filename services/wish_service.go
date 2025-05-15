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

func (s *WishService) GetWishes(filters map[string]any) ([]models.Wish, int64, error) {
	query := s.db.Model(&models.Wish{})

	if content, ok := filters["content"].(string); ok && content != "" {
		query = query.Where("content LIKE ?", "%"+content+"%")
	}

	// 处理完成状态过滤
	if isDoneStr, ok := filters["isDone"].(string); ok && isDoneStr != "" {
		if isBool, err := strconv.ParseBool(isDoneStr); err == nil {
			if isBool {
				// 已完成：active_record_id 不为空
				query = query.Where("active_record_id IS NOT NULL")
			} else {
				// 未完成：active_record_id 为空
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
	if err := query.Limit(pageSize).Offset(offset).Find(&wishes).Error; err != nil {
		return nil, 0, err
	}

	return wishes, total, nil
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
