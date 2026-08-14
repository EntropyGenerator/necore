package model

import "gorm.io/gorm"

type DepartmentMember struct {
	gorm.Model

	DepartmentId string `gorm:"index;not null;uniqueIndex:idx_dept_user" json:"departmentId"`
	Username     string `gorm:"index;not null;uniqueIndex:idx_dept_user" json:"username"`
	SortOrder    int    `gorm:"not null;default:0" json:"sortOrder"`
	IsLeader     bool   `gorm:"not null;default:false" json:"isLeader"`
}
