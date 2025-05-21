package services

import (
	"fmt"
	"wishes/models"

	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		db: db,
	}
}

func (s *UserService) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) GetUsers(pageIndex, pageSize int, isAdmin bool) ([]models.User, int64, error) {
	query := s.db.Model(&models.User{})

	if isAdmin {
		query = query.Where("is_admin = true")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pageIndex - 1) * pageSize

	var users []models.User
	if err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (s *UserService) UpdateUserAdminStatus(userID uint, isAdmin bool) error {
	result := s.db.Model(&models.User{}).Where("id = ?", userID).Update("is_admin", isAdmin)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("未找到ID为%d的用户", userID)
	}
	return nil
}

func (s *UserService) GetAdminUsers(pageIndex, pageSize int) ([]models.User, int64, error) {
	return s.GetUsers(pageIndex, pageSize, true)
}

func (s *UserService) GetNonAdminUsers(pageIndex, pageSize int) ([]models.User, int64, error) {
	query := s.db.Model(&models.User{}).Where("is_admin = ?", false)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pageIndex - 1) * pageSize

	var users []models.User
	if err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
