package model

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	Username string `gorm:"uniqueIndex;not null" json:"username"`
	Password string `gorm:"not null" json:"password"` // sha256 hashed
	Group    string `json:"group"`                    // json array: []string
	Tags     string `json:"tags"`                     // json array: []string
}

type UserAvatar struct {
	gorm.Model

	Username string `gorm:"uniqueIndex;not null" json:"username"`
	Avatar   string `json:"avatar"`
}

// func AddUser(username string, password string) {
// 	db := database.GetInstance()
// 	ctx := context.Background()
// 	err := gorm.G[User](db.Database).Create(ctx, &User{
// 		Username: username,
// 		Password: password,
// 	})
// 	if err != nil {
// 		// TODO
// 		panic(err)
// 	}
// }
