package mysql

import (
	"bluebell/models"
	"go.uber.org/zap"
)

func InsertUser(u *models.User) {
	db.Create(u)
}

func QueryUserByUsername(name string) bool {
	var count int64 // 使用 int64 替代 int，因为数据库中的数量可能非常大
	result := db.Model(&models.User{}).Where("username = ?", name).Count(&count)
	if result.Error != nil {
		// 处理可能出现的错误
		zap.L().Error("QueryUserByUsername is err", zap.Error(result.Error))
		return false
	}
	return count > 0
}

func Login(user *models.User) bool {
	//var count int64 // 使用 int64 替代 int，因为数据库中的数量可能非常大
	result := db.Model(&models.User{}).Where("username = ? AND password = ?",
		user.Username,
		user.Password).First(user)
	if result.Error != nil {
		// 处理可能出现的错误
		zap.L().Error("QueryUserByUsername is err", zap.Error(result.Error))
		return false
	}
	return user == nil
}
