package repo

import "gorm.io/gorm"

func scopeTenant(tenantID uint64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if tenantID == 0 {
			tenantID = 1
		}
		return db.Where("tenant_id = ?", tenantID)
	}
}

func normalizeTenantID(tenantID uint64) uint64 {
	if tenantID == 0 {
		return 1
	}
	return tenantID
}

// NormalizeTenantID exported for services.
func NormalizeTenantID(tenantID uint64) uint64 { return normalizeTenantID(tenantID) }
