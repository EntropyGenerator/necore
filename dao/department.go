package dao

import (
	"necore/database"
	"necore/model"

	"gorm.io/gorm"
)

func GetDepartmentList() ([]model.Department, error) {
	db := database.GetDepartmentDatabase()
	var departments []model.Department
	err := db.Order("sort_order asc, id asc").Find(&departments).Error
	return departments, err
}

func GetDepartmentByID(id string) (*model.Department, error) {
	db := database.GetDepartmentDatabase()
	var department model.Department
	if err := db.Where("id = ?", id).First(&department).Error; err != nil {
		return nil, err
	}
	return &department, nil
}

func CreateDepartment(department model.Department) error {
	return database.GetDepartmentDatabase().Create(&department).Error
}

func UpdateDepartment(id string, fields map[string]any) error {
	result := database.GetDepartmentDatabase().
		Model(&model.Department{}).
		Where("id = ?", id).
		Updates(fields)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func UpdateDepartmentOrders(orders []model.Department) error {
	db := database.GetDepartmentDatabase()
	return db.Transaction(func(tx *gorm.DB) error {
		for _, department := range orders {
			result := tx.Model(&model.Department{}).
				Where("id = ?", department.Id).
				Update("sort_order", department.SortOrder)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return gorm.ErrRecordNotFound
			}
		}
		return nil
	})
}

func DeleteDepartment(id string) error {
	db := database.GetDepartmentDatabase()
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("department_id = ?", id).Delete(&model.DepartmentMember{}).Error; err != nil {
			return err
		}
		result := tx.Unscoped().Where("id = ?", id).Delete(&model.Department{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func GetDepartmentMembers(departmentID string) ([]model.DepartmentMember, error) {
	db := database.GetDepartmentDatabase()
	var members []model.DepartmentMember
	err := db.Where("department_id = ?", departmentID).
		Order("sort_order asc, username asc").
		Find(&members).Error
	return members, err
}

func DepartmentMemberExists(departmentID, username string) (bool, error) {
	var count int64
	err := database.GetDepartmentDatabase().
		Model(&model.DepartmentMember{}).
		Where("department_id = ? AND username = ?", departmentID, username).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func AddDepartmentMember(member model.DepartmentMember) error {
	return database.GetDepartmentDatabase().Create(&member).Error
}

func RemoveDepartmentMember(departmentID, username string) error {
	result := database.GetDepartmentDatabase().
		Unscoped().
		Where("department_id = ? AND username = ?", departmentID, username).
		Delete(&model.DepartmentMember{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func UpdateDepartmentMemberOrders(departmentID string, members []model.DepartmentMember) error {
	db := database.GetDepartmentDatabase()
	return db.Transaction(func(tx *gorm.DB) error {
		for _, member := range members {
			result := tx.Model(&model.DepartmentMember{}).
				Where("department_id = ? AND username = ?", departmentID, member.Username).
				Updates(map[string]any{
					"sort_order": member.SortOrder,
					"is_leader":  member.IsLeader,
				})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return gorm.ErrRecordNotFound
			}
		}
		return nil
	})
}

func UpdateDepartmentMemberLeader(departmentID, username string, isLeader bool) error {
	result := database.GetDepartmentDatabase().
		Model(&model.DepartmentMember{}).
		Where("department_id = ? AND username = ?", departmentID, username).
		Update("is_leader", isLeader)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
