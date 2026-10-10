package repo

import (
	"strings"
	"testing"

	"productcore/internal/dto"
)

func TestItemOrderSQLDefaultsToSalesDesc(t *testing.T) {
	sql := itemOrderSQL(dto.DistributionItemQuery{})
	if !strings.Contains(sql, "sales") || !strings.Contains(sql, "DESC") {
		t.Fatalf("got %s", sql)
	}
	if strings.Contains(sql, "?") {
		t.Fatalf("placeholder leaked into order sql: %s", sql)
	}
}

func TestItemOrderSQLWhitelist(t *testing.T) {
	sql := itemOrderSQL(dto.DistributionItemQuery{SortBy: "listedAt", SortOrder: "asc"})
	if !strings.Contains(sql, "listed_at") || !strings.Contains(sql, "ASC") {
		t.Fatalf("got %s", sql)
	}
	sql = itemOrderSQL(dto.DistributionItemQuery{SortBy: "name;drop", SortOrder: "desc"})
	if strings.Contains(sql, "drop") || !strings.Contains(sql, "sales") {
		t.Fatalf("unknown column was not ignored: %s", sql)
	}
}
