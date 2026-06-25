package productimport

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"productcore/internal/dto"
	"productcore/internal/storage"
)

const SkuDetailFileName = "sku明细.txt"

const exportRootLabel = "商品中心"

// BuildExportFolderName ProductCore 导出包根目录/zip 文件名（不含 .zip）
func BuildExportFolderName(productID uint64) string {
	return fmt.Sprintf("%s_商品ID_%d", exportRootLabel, productID)
}

// BuildSkuFileName SKU 图片文件名，格式 SKU01_规格名(规格值).jpg
func BuildSkuFileName(index int, specName, specValue, ext string) string {
	specName = sanitizePathPart(specName)
	specValue = sanitizePathPart(specValue)
	if specName == "" {
		specName = "商品规格"
	}
	if specValue == "" {
		specValue = fmt.Sprintf("规格%d", index)
	}
	ext = normalizeImageExt(ext)
	return fmt.Sprintf("SKU%02d_%s(%s)%s", index, specName, specValue, ext)
}

// BuildSkuPlaceholderFileName 无规格图时的占位文件名（无后缀）
func BuildSkuPlaceholderFileName(index int, specName, specValue string) string {
	specName = sanitizePathPart(specName)
	specValue = sanitizePathPart(specValue)
	if specName == "" {
		specName = "商品规格"
	}
	if specValue == "" {
		specValue = fmt.Sprintf("规格%d", index)
	}
	return fmt.Sprintf("SKU%02d_%s(%s)", index, specName, specValue)
}

// ExportProduct 将商品导出为 ProductCore 标准 zip 包，返回 zip 路径与 cleanup
func ExportProduct(product *dto.ProductDTO, store storage.Storage) (zipPath string, cleanup func(), err error) {
	cleanup = func() {}
	if product == nil {
		return "", cleanup, fmt.Errorf("product is nil")
	}
	if product.ID == 0 {
		return "", cleanup, fmt.Errorf("商品 ID 无效")
	}
	if len(product.Skus) == 0 {
		return "", cleanup, fmt.Errorf("商品无 SKU，无法导出")
	}

	folderName := BuildExportFolderName(product.ID)

	tmpDir, err := os.MkdirTemp("", "product-export-*")
	if err != nil {
		return "", cleanup, err
	}
	cleanup = func() { _ = os.RemoveAll(tmpDir) }

	rootDir := filepath.Join(tmpDir, folderName)
	dirs := []string{
		filepath.Join(rootDir, DirMain),
		filepath.Join(rootDir, DirSKU),
		filepath.Join(rootDir, DirDetail),
		filepath.Join(rootDir, DirVideo),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return "", cleanup, err
		}
	}

	mainURLs := collectMainURLs(product)
	if len(mainURLs) == 0 {
		return "", cleanup, fmt.Errorf("商品无主图，无法导出")
	}
	for i, url := range mainURLs {
		dest := filepath.Join(rootDir, DirMain, fmt.Sprintf("%d%s", i+1, extFromURL(url)))
		if err := store.CopyStoredToPath(url, dest); err != nil {
			return "", cleanup, fmt.Errorf("复制主图失败: %w", err)
		}
	}

	for i, url := range detailPicURLs(product) {
		dest := filepath.Join(rootDir, DirDetail, fmt.Sprintf("%d%s", i+1, extFromURL(url)))
		if err := store.CopyStoredToPath(url, dest); err != nil {
			return "", cleanup, fmt.Errorf("复制详情图失败: %w", err)
		}
	}

	videoIdx := 0
	for _, url := range videoURLs(product) {
		videoIdx++
		dest := filepath.Join(rootDir, DirVideo, fmt.Sprintf("%d%s", videoIdx, extFromURL(url)))
		if err := store.CopyStoredToPath(url, dest); err != nil {
			return "", cleanup, fmt.Errorf("复制视频失败: %w", err)
		}
	}

	specName := exportSpecName(product.SkuSpecs, product.Skus)
	for i, sku := range product.Skus {
		idx := i + 1
		specValue := exportSpecValue(product.SkuSpecs, sku)
		picURL := resolveSkuPicURL(product, sku, specValue)
		if picURL == "" {
			dest := filepath.Join(rootDir, DirSKU, BuildSkuPlaceholderFileName(idx, specName, specValue))
			if err := os.WriteFile(dest, nil, 0o644); err != nil {
				return "", cleanup, fmt.Errorf("创建 SKU 占位文件失败: %w", err)
			}
			continue
		}
		dest := filepath.Join(rootDir, DirSKU, BuildSkuFileName(idx, specName, specValue, extFromURL(picURL)))
		if err := store.CopyStoredToPath(picURL, dest); err != nil {
			return "", cleanup, fmt.Errorf("复制 SKU 图片失败: %w", err)
		}
	}

	detailText := BuildSkuDetailText(product, specName)
	detailPath := filepath.Join(rootDir, SkuDetailFileName)
	if err := os.WriteFile(detailPath, []byte(detailText), 0o644); err != nil {
		return "", cleanup, err
	}

	zipPath = filepath.Join(tmpDir, folderName+".zip")
	if err := zipDir(rootDir, zipPath); err != nil {
		return "", cleanup, err
	}
	return zipPath, cleanup, nil
}

func collectMainURLs(product *dto.ProductDTO) []string {
	var urls []string
	if u := strings.TrimSpace(product.Pic); u != "" {
		urls = append(urls, u)
	}
	for _, u := range product.AlbumPics {
		if u = strings.TrimSpace(u); u != "" {
			urls = append(urls, u)
		}
	}
	return urls
}

func detailPicURLs(product *dto.ProductDTO) []string {
	if product.Media == nil {
		return nil
	}
	var urls []string
	for _, u := range product.Media.DetailPics {
		if u = strings.TrimSpace(u); u != "" {
			urls = append(urls, u)
		}
	}
	return urls
}

func videoURLs(product *dto.ProductDTO) []string {
	if product.Media == nil || product.Media.Videos == nil {
		return nil
	}
	v := product.Media.Videos
	var urls []string
	for _, u := range []string{v.Ratio11, v.Ratio34, v.Ratio169, v.Ratio916} {
		if u = strings.TrimSpace(u); u != "" {
			urls = append(urls, u)
		}
	}
	return urls
}

func exportSpecName(specs []dto.SkuSpecDTO, skus []dto.SkuDTO) string {
	for _, spec := range specs {
		if name := strings.TrimSpace(spec.Name); name != "" {
			return name
		}
	}
	if name := dominantSpecKeyFromSkus(skus); name != "" {
		return name
	}
	return "商品规格"
}

func exportSpecValue(specs []dto.SkuSpecDTO, sku dto.SkuDTO) string {
	var parts []string
	for _, spec := range specs {
		name := strings.TrimSpace(spec.Name)
		if name == "" {
			continue
		}
		if v := lookupSpecValue(name, spec, sku); v != "" {
			parts = append(parts, v)
		}
	}
	if len(parts) > 0 {
		return strings.Join(parts, "/")
	}
	keys := sortedSpecKeys(sku.Specs)
	for _, k := range keys {
		if v := strings.TrimSpace(sku.Specs[k]); v != "" {
			parts = append(parts, v)
		}
	}
	if len(parts) == 0 {
		return "默认"
	}
	return strings.Join(parts, "/")
}

// lookupSpecValue 按 sku_specs 中的规格名取值；兼容 Specs 键名与规格名不一致的情况。
func lookupSpecValue(specName string, spec dto.SkuSpecDTO, sku dto.SkuDTO) string {
	if v := strings.TrimSpace(sku.Specs[specName]); v != "" {
		return v
	}
	if len(sku.Specs) == 1 {
		for _, v := range sku.Specs {
			if v = strings.TrimSpace(v); v != "" {
				return v
			}
		}
	}
	for _, v := range sku.Specs {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		for _, sv := range spec.Values {
			if sv.Value == v {
				return v
			}
		}
	}
	return ""
}

func dominantSpecKeyFromSkus(skus []dto.SkuDTO) string {
	counts := map[string]int{}
	for _, sku := range skus {
		for k := range sku.Specs {
			if k = strings.TrimSpace(k); k != "" {
				counts[k]++
			}
		}
	}
	if len(counts) == 0 {
		return ""
	}
	best := ""
	bestCount := 0
	for k, n := range counts {
		if n > bestCount || (n == bestCount && (best == "" || k < best)) {
			best, bestCount = k, n
		}
	}
	return best
}

func sortedSpecKeys(specs map[string]string) []string {
	keys := make([]string, 0, len(specs))
	for k := range specs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func resolveSkuPicURL(product *dto.ProductDTO, sku dto.SkuDTO, specValue string) string {
	if u := strings.TrimSpace(sku.Pic); u != "" {
		return u
	}
	if len(product.SkuSpecs) > 0 {
		spec := product.SkuSpecs[0]
		valKey := specValue
		if idx := strings.Index(specValue, "/"); idx >= 0 {
			valKey = specValue[:idx]
		}
		for _, v := range spec.Values {
			if v.Value == valKey && strings.TrimSpace(v.Pic) != "" {
				return v.Pic
			}
		}
	}
	return ""
}

// BuildSkuDetailText 生成 sku 明细文本（手动铺货参考）
func BuildSkuDetailText(product *dto.ProductDTO, specName string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "ProductCore 商品ID: %d\r\n", product.ID)
	fmt.Fprintf(&b, "商品名称: %s\r\n", product.Name)
	materialCode := strings.TrimSpace(product.MaterialCode)
	if materialCode == "" {
		materialCode = "-"
	}
	fmt.Fprintf(&b, "资料编码: %s\r\n", materialCode)
	fmt.Fprintf(&b, "商品来源: %s\r\n", strings.TrimSpace(product.Source))
	fmt.Fprintf(&b, "品牌: %s\r\n", product.BrandName)
	fmt.Fprintf(&b, "分类: %s\r\n", product.CategoryName)
	fmt.Fprintf(&b, "规格维度: %s\r\n", specName)
	fmt.Fprintf(&b, "SKU 数量: %d\r\n", len(product.Skus))
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "--- SKU 明细 ---")
	fmt.Fprintf(&b, "%-4s | %-24s | %-20s | %8s | %6s | %6s\r\n",
		"序号", "规格", "SKU编码", "销售价", "库存", "重量")
	fmt.Fprintf(&b, "%s\r\n", strings.Repeat("-", 80))
	for i, sku := range product.Skus {
		specDisplay := exportSpecValue(product.SkuSpecs, sku)
		fmt.Fprintf(&b, "%-4s | %-24s | %-20s | %8.2f | %6d | %6.2f\r\n",
			fmt.Sprintf("%02d", i+1),
			truncateRunes(specDisplay, 24),
			truncateRunes(sku.SkuCode, 20),
			sku.Price,
			sku.Stock,
			sku.Weight,
		)
	}
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "说明: 本文件仅供手动铺货参考。")
	return b.String()
}

func truncateRunes(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-1]) + "…"
}

func sanitizePathPart(s string) string {
	s = strings.TrimSpace(s)
	replacer := strings.NewReplacer(
		"/", "_", "\\", "_", ":", "_", "*", "_", "?", "_",
		"\"", "_", "<", "_", ">", "_", "|", "_",
	)
	return replacer.Replace(s)
}

func normalizeImageExt(ext string) string {
	ext = strings.ToLower(ext)
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
		return ext
	default:
		return ".jpg"
	}
}

func extFromURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ".jpg"
	}
	path := raw
	if idx := strings.Index(raw, "://"); idx >= 0 {
		if slash := strings.Index(raw[idx+3:], "/"); slash >= 0 {
			path = raw[idx+3+slash:]
		}
	}
	ext := filepath.Ext(path)
	if ext == "" {
		return ".jpg"
	}
	ext = strings.ToLower(ext)
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".mp4", ".mov", ".webm", ".avi":
		return ext
	default:
		return ".jpg"
	}
}

func zipDir(srcDir, zipPath string) error {
	out, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer out.Close()

	w := zip.NewWriter(out)
	defer w.Close()

	baseDir := filepath.Dir(srcDir)
	rootName := filepath.Base(srcDir)

	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(baseDir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if info.IsDir() {
			if rel == rootName {
				return nil
			}
			_, err := w.Create(rel + "/")
			return err
		}
		hdr, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		hdr.Name = rel
		hdr.Method = zip.Deflate
		writer, err := w.CreateHeader(hdr)
		if err != nil {
			return err
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(writer, f)
		return err
	})
}
