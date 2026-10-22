package service

import "testing"

func TestCleanGeneratedTitle(t *testing.T) {
	got := cleanGeneratedTitle("「SHIMANO禧玛诺105链条 副本」\n不要带上这一行")
	if got != "SHIMANO禧玛诺105链条" {
		t.Fatalf("got %q", got)
	}
	long := cleanGeneratedTitle("禧玛诺山地车链条变速器保护器后拨公路车配件十二速十一速十速九速八速")
	if douyinTitleWeight(long) > 60 {
		t.Fatalf("weight %d title %q", douyinTitleWeight(long), long)
	}
	if douyinTitleWeight("禧玛诺链条") != 10 {
		t.Fatal(douyinTitleWeight("禧玛诺链条"))
	}
}
