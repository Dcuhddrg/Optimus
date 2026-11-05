// Package task holds scheduled jobs, as GVA's server/task.
package task

import (
	"errors"
	"fmt"
	"time"

	"optimus/server/global"
	"optimus/server/model/common"
	"optimus/server/utils"

	"gorm.io/gorm"
)

const defaultAuditLogRetentionDays = 90

// ClearTable deletes expired audit logs and revoked tokens that have outlived
// their lifetime, as GVA's task.ClearTable. Rows are deleted in batches.
func ClearTable(db *gorm.DB) error {
	if db == nil {
		return errors.New("db Cannot be empty")
	}
	retentionDays := global.OPTIMUS_CONFIG.System.AuditLogRetentionDays
	if retentionDays <= 0 {
		retentionDays = defaultAuditLogRetentionDays
	}
	tokenLifetime, err := utils.ParseDuration(global.OPTIMUS_CONFIG.JWT.ExpiresTime)
	if err != nil {
		return err
	}
	clearTableDetail := []common.ClearDB{
		{TableName: "sys_audit_logs", CompareField: "created_at", Interval: fmt.Sprintf("%dh", retentionDays*24)},
		{TableName: "jwt_blacklists", CompareField: "created_at", Interval: tokenLifetime.String()},
	}
	for _, detail := range clearTableDetail {
		duration, err := time.ParseDuration(detail.Interval)
		if err != nil {
			return err
		}
		if duration < 0 {
			return errors.New("parse duration < 0")
		}
		before := time.Now().Add(-duration)
		for {
			result := db.Exec(fmt.Sprintf("DELETE FROM %s WHERE %s < ? LIMIT 1000", detail.TableName, detail.CompareField), before)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				break
			}
		}
	}
	return nil
}
