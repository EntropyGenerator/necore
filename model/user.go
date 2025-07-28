package model

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	Username   string `gorm:"uniqueIndex;not null" json:"username"`
	Password   string `gorm:"not null" json:"password"` // sha256 hashed
	Group      string `json:"group"`                    // json array: []string
	Department string `json:"department"`               // json array: []string
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
