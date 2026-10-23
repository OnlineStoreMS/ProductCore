package service

import (
	"strings"
	"testing"

	"productcore/internal/dto"
)

func TestCleanGeneratedTitle(t *testing.T) {
	got := cleanGeneratedTitle("「SHIMANO禧玛诺105链条 副本」\n不要带上这一行")
	if got != "105链条" {
		t.Fatalf("got %q", got)
	}
	for _, raw := range []string{"shimano链条", "ShiMaNo链条", "SHIMANO链条", "禧马诺链条", "喜玛诺链条"} {
		cleaned := cleanGeneratedTitle(raw)
		if strings.Contains(strings.ToLower(cleaned), "shimano") || strings.Contains(cleaned, "禧") || strings.Contains(cleaned, "喜玛") || strings.Contains(cleaned, "喜马") {
			t.Fatalf("%q -> %q", raw, cleaned)
		}
	}
	long := cleanGeneratedTitle("禧玛诺山地车链条变速器保护器后拨公路车配件十二速十一速十速九速八速")
	if douyinTitleWeight(long) > 60 {
		t.Fatalf("weight %d title %q", douyinTitleWeight(long), long)
	}
	if douyinTitleWeight("禧玛诺链条") != 10 {
		t.Fatal(douyinTitleWeight("禧玛诺链条"))
	}
}

func TestListingScenes(t *testing.T) {
	road := listingScenes(&dto.ProductDTO{Name: "105 UT 公路车链条"}, nil)
	if len(road) != 1 || road[0] != "公路自行车" {
		t.Fatalf("road %#v", road)
	}
	both := listingScenes(&dto.ProductDTO{Name: "11速公路车 12速山地车链条"}, nil)
	if strings.Join(both, ",") != "公路自行车,山地自行车" {
		t.Fatalf("both %#v", both)
	}
	plain := listingScenes(&dto.ProductDTO{Name: "后变速器", CategoryName: "其他维修工具"}, []string{"维修工具", "其他维修工具"})
	if len(plain) != 1 || plain[0] != "自行车" {
		t.Fatalf("plain %#v", plain)
	}
}
