package system

import (
	"context"

	"optimus/server/global"
	systemModel "optimus/server/model/system"

	"go.uber.org/zap"
)

type JwtService struct{}

var JwtServiceApp = new(JwtService)

// JsonInBlacklist revokes a token: it is persisted and cached until it expires.
func (jwtService *JwtService) JsonInBlacklist(ctx context.Context, jwtList systemModel.JwtBlacklist) error {
	if err := global.OPTIMUS_DB.WithContext(ctx).Create(&jwtList).Error; err != nil {
		return err
	}
	global.BlackCache.SetDefault(jwtList.Jwt, struct{}{})
	return nil
}

// IsBlacklist reports whether a token has been revoked.
func (jwtService *JwtService) IsBlacklist(jwt string) bool {
	_, ok := global.BlackCache.Get(jwt)
	return ok
}

// LoadAll loads persisted revoked tokens into BlackCache at startup, as GVA.
func LoadAll() {
	var data []string
	if err := global.OPTIMUS_DB.Model(&systemModel.JwtBlacklist{}).Select("jwt").Find(&data).Error; err != nil {
		zap.L().Error("load jwt blacklist failed", zap.Error(err))
		return
	}
	for i := 0; i < len(data); i++ {
		global.BlackCache.SetDefault(data[i], struct{}{})
	}
}
