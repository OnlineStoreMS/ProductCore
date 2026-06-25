package dto

import (
	"encoding/json"
	"fmt"
	"strings"
)

// legacySkuSpec 兼容旧版 values 为 string 数组
type legacySkuSpec struct {
	Name   string          `json:"name"`
	Values json.RawMessage `json:"values"`
}

// CompactSkuSpecs 去掉空白规格值；无有效内容时返回空切片（不入库占位项）
func CompactSkuSpecs(specs []SkuSpecDTO) []SkuSpecDTO {
	out := make([]SkuSpecDTO, 0, len(specs))
	for _, spec := range specs {
		filled := make([]SkuSpecValueDTO, 0, len(spec.Values))
		for _, v := range spec.Values {
			if isBlankSpecValue(v) {
				continue
			}
			filled = append(filled, v)
		}
		name := strings.TrimSpace(spec.Name)
		if name == "" || len(filled) == 0 {
			continue
		}
		out = append(out, SkuSpecDTO{Name: name, Values: filled})
	}
	return out
}

func isBlankSpecValue(v SkuSpecValueDTO) bool {
	return strings.TrimSpace(v.Value) == "" &&
		strings.TrimSpace(v.Remark) == "" &&
		strings.TrimSpace(v.Pic) == ""
}

// NormalizeSkuSpecs 将 sku_specs_json 归一化为 SkuSpecValueDTO 结构
func NormalizeSkuSpecs(raw string) []SkuSpecDTO {
	if raw == "" || raw == "null" {
		return nil
	}
	var legacy []legacySkuSpec
	if err := json.Unmarshal([]byte(raw), &legacy); err != nil {
		return nil
	}
	out := make([]SkuSpecDTO, 0, len(legacy))
	for _, item := range legacy {
		spec := SkuSpecDTO{Name: item.Name}
		if len(item.Values) == 0 {
			out = append(out, spec)
			continue
		}
		// 尝试新格式 [{value,remark,pic}]
		var values []SkuSpecValueDTO
		if err := json.Unmarshal(item.Values, &values); err == nil && len(values) > 0 {
			spec.Values = values
			out = append(out, spec)
			continue
		}
		// 旧格式 ["红","蓝"]
		var strs []string
		if err := json.Unmarshal(item.Values, &strs); err == nil {
			for _, s := range strs {
				spec.Values = append(spec.Values, SkuSpecValueDTO{Value: s})
			}
		}
		out = append(out, spec)
	}
	return CompactSkuSpecs(out)
}

// ValidateSkuSpecsNoDuplicateValues 同一规格项下规格值不可重复（忽略空白值）
func ValidateSkuSpecsNoDuplicateValues(specs []SkuSpecDTO) error {
	for _, spec := range specs {
		name := strings.TrimSpace(spec.Name)
		if name == "" {
			continue
		}
		seen := make(map[string]struct{})
		for _, v := range spec.Values {
			val := strings.TrimSpace(v.Value)
			if val == "" {
				continue
			}
			if _, dup := seen[val]; dup {
				return fmt.Errorf("spec %q has duplicate value %q", name, val)
			}
			seen[val] = struct{}{}
		}
	}
	return nil
}

// SpecValueTexts 提取规格值文本列表（用于 SKU 组合）
func SpecValueTexts(spec SkuSpecDTO) []string {
	out := make([]string, 0, len(spec.Values))
	for _, v := range spec.Values {
		if v.Value != "" {
			out = append(out, v.Value)
		}
	}
	return out
}

// ProductMediaWithoutPics34 写入 media_json 时排除 pics34（3:4 主图走独立列）
func ProductMediaWithoutPics34(m *ProductMediaDTO) *ProductMediaDTO {
	if m == nil {
		return nil
	}
	return CompactProductMedia(&ProductMediaDTO{
		Videos:     m.Videos,
		Materials:  m.Materials,
		DetailPics: m.DetailPics,
	})
}

// Pics34FromMedia 从 API media 提取 3:4 主图列表（含空数组表示清空）
func Pics34FromMedia(m *ProductMediaDTO) []string {
	if m == nil || m.Pics34 == nil {
		return []string{}
	}
	out := make([]string, len(m.Pics34))
	copy(out, m.Pics34)
	return out
}

// MergePics34IntoMedia 将 3:4 主图合并进 API 响应 media
func MergePics34IntoMedia(media *ProductMediaDTO, pics34 []string) *ProductMediaDTO {
	if len(pics34) == 0 {
		return media
	}
	if media == nil {
		return &ProductMediaDTO{Pics34: pics34}
	}
	media.Pics34 = pics34
	return media
}
func ParseProductMedia(raw string) *ProductMediaDTO {
	if raw == "" || raw == "null" {
		return nil
	}
	var media ProductMediaDTO
	if err := json.Unmarshal([]byte(raw), &media); err != nil {
		return nil
	}
	return CompactProductMedia(&media)
}

// CompactProductMedia 去掉空字段，全空则返回 nil
func CompactProductMedia(m *ProductMediaDTO) *ProductMediaDTO {
	if m == nil {
		return nil
	}
	out := &ProductMediaDTO{}
	if len(m.Pics34) > 0 {
		out.Pics34 = m.Pics34
	}
	if len(m.DetailPics) > 0 {
		out.DetailPics = m.DetailPics
	}
	if v := compactVideos(m.Videos); v != nil {
		out.Videos = v
	}
	if mat := compactMaterials(m.Materials); mat != nil {
		out.Materials = mat
	}
	if out.Pics34 == nil && out.DetailPics == nil && out.Videos == nil && out.Materials == nil {
		return nil
	}
	return out
}

func compactVideos(v *ProductVideosDTO) *ProductVideosDTO {
	if v == nil {
		return nil
	}
	out := &ProductVideosDTO{}
	if v.Ratio11 != "" {
		out.Ratio11 = v.Ratio11
	}
	if v.Ratio34 != "" {
		out.Ratio34 = v.Ratio34
	}
	if v.Ratio169 != "" {
		out.Ratio169 = v.Ratio169
	}
	if v.Ratio916 != "" {
		out.Ratio916 = v.Ratio916
	}
	if out.Ratio11 == "" && out.Ratio34 == "" && out.Ratio169 == "" && out.Ratio916 == "" {
		return nil
	}
	return out
}

func compactMaterials(m *ProductMaterialsDTO) *ProductMaterialsDTO {
	if m == nil {
		return nil
	}
	out := &ProductMaterialsDTO{}
	if m.White != "" {
		out.White = m.White
	}
	if m.Transparent != "" {
		out.Transparent = m.Transparent
	}
	if m.Guide34 != "" {
		out.Guide34 = m.Guide34
	}
	if m.Long != "" {
		out.Long = m.Long
	}
	if out.White == "" && out.Transparent == "" && out.Guide34 == "" && out.Long == "" {
		return nil
	}
	return out
}
