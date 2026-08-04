package productimport

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	DirMain   = "主图"
	DirSKU    = "SKU"
	DirDetail = "详情图"
	DirVideo  = "视频"
)

var (
	folderNamePattern = regexp.MustCompile(`^(.+)_商品ID_(.+)$`)
	skuIndexPattern   = regexp.MustCompile(`(?i)^SKU(\d+)_(.+)$`)
	numSuffixPattern  = regexp.MustCompile(`(\d+)`)
)

var allowedSkuImageExt = map[string]struct{}{
	".jpg": {}, ".jpeg": {}, ".png": {}, ".webp": {},
}

var allowedVideoExt = map[string]struct{}{
	".mp4": {}, ".mov": {}, ".webm": {}, ".avi": {},
}

// ParsedPackage 解压并解析后的商品包结构（本地路径）
type ParsedPackage struct {
	Source       string
	MaterialCode string
	RootDir      string
	MainPics     []string
	DetailPics   []string
	Videos       []ParsedVideo
	Skus         []ParsedSku
}

type ParsedVideo struct {
	Path  string
	Ratio string // 1:1 / 3:4 / 16:9 / 9:16
}

type ParsedSku struct {
	Index     int
	Code      string
	SpecName  string
	SpecValue string
	PicPath   string
}

// ParseZip 解压 zip 并解析目录结构
func ParseZip(zipPath string) (*ParsedPackage, func(), error) {
	cleanup := func() {}
	tmpDir, err := os.MkdirTemp("", "product-import-*")
	if err != nil {
		return nil, cleanup, err
	}
	cleanup = func() { _ = os.RemoveAll(tmpDir) }

	if err := extractZip(zipPath, tmpDir); err != nil {
		cleanup()
		return nil, func() {}, err
	}

	rootDir, folderName, err := findRootDir(tmpDir)
	if err != nil {
		cleanup()
		return nil, func() {}, err
	}

	source, materialCode, err := parseFolderName(folderName)
	if err != nil {
		cleanup()
		return nil, func() {}, err
	}

	pkg := &ParsedPackage{
		Source:       source,
		MaterialCode: materialCode,
		RootDir:      rootDir,
	}

	mainDir := filepath.Join(rootDir, DirMain)
	skuDir := filepath.Join(rootDir, DirSKU)
	detailDir := filepath.Join(rootDir, DirDetail)
	videoDir := filepath.Join(rootDir, DirVideo)

	if !isDir(mainDir) {
		cleanup()
		return nil, func() {}, fmt.Errorf("缺少 %s 文件夹", DirMain)
	}
	if !isDir(skuDir) {
		cleanup()
		return nil, func() {}, fmt.Errorf("缺少 %s 文件夹", DirSKU)
	}

	mainPics, err := listImageFiles(mainDir)
	if err != nil {
		cleanup()
		return nil, func() {}, err
	}
	if len(mainPics) == 0 {
		cleanup()
		return nil, func() {}, fmt.Errorf("%s 文件夹中没有图片", DirMain)
	}
	pkg.MainPics = mainPics

	if isDir(detailDir) {
		detailPics, err := listImageFiles(detailDir)
		if err != nil {
			cleanup()
			return nil, func() {}, err
		}
		pkg.DetailPics = detailPics
	}

	if isDir(videoDir) {
		videos, err := listVideoFiles(videoDir)
		if err != nil {
			cleanup()
			return nil, func() {}, err
		}
		pkg.Videos = videos
	}

	skus, err := parseSkuDir(skuDir)
	if err != nil {
		cleanup()
		return nil, func() {}, err
	}
	if len(skus) == 0 {
		cleanup()
		return nil, func() {}, fmt.Errorf("%s 文件夹中没有有效的 SKU 图片", DirSKU)
	}
	pkg.Skus = skus

	return pkg, cleanup, nil
}

func extractZip(zipPath, dest string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("无法打开 zip 文件: %w", err)
	}
	defer r.Close()

	destAbs, err := filepath.Abs(dest)
	if err != nil {
		return err
	}

	for _, f := range r.File {
		name := filepath.Clean(strings.ReplaceAll(f.Name, "\\", "/"))
		if strings.HasPrefix(name, "..") {
			return errors.New("zip 包含非法路径")
		}
		target := filepath.Join(dest, name)
		targetAbs, err := filepath.Abs(target)
		if err != nil {
			return err
		}
		if !strings.HasPrefix(targetAbs, destAbs+string(os.PathSeparator)) && targetAbs != destAbs {
			return errors.New("zip 包含非法路径")
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := extractFile(f, target); err != nil {
			return err
		}
	}
	return nil
}

func extractFile(f *zip.File, target string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, rc)
	return err
}

func findRootDir(tmpDir string) (rootDir, folderName string, err error) {
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return "", "", err
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			dirs = append(dirs, e.Name())
		}
	}
	if len(dirs) == 0 {
		return "", "", errors.New("zip 中未找到商品文件夹")
	}
	if len(dirs) > 1 {
		return "", "", errors.New("zip 中只能包含一个商品文件夹")
	}
	folderName = dirs[0]
	if !folderNamePattern.MatchString(folderName) {
		return "", "", fmt.Errorf("文件夹名称格式应为「来源_商品ID_资料编码」，当前为 %q", folderName)
	}
	return filepath.Join(tmpDir, folderName), folderName, nil
}

func parseFolderName(folderName string) (source, materialCode string, err error) {
	m := folderNamePattern.FindStringSubmatch(folderName)
	if len(m) != 3 {
		return "", "", fmt.Errorf("无法解析文件夹名称 %q", folderName)
	}
	source = strings.TrimSpace(m[1])
	materialCode = strings.TrimSpace(m[2])
	if source == "" {
		return "", "", errors.New("商品来源不能为空")
	}
	if materialCode == "" {
		return "", "", errors.New("资料编码不能为空")
	}
	return source, materialCode, nil
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func listImageFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		switch ext {
		case ".jpg", ".jpeg", ".png", ".webp":
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	sortByNumericSuffix(files)
	return files, nil
}

func listVideoFiles(dir string) ([]ParsedVideo, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if _, ok := allowedVideoExt[ext]; ok {
			paths = append(paths, filepath.Join(dir, e.Name()))
		}
	}
	sortByNumericSuffix(paths)
	out := make([]ParsedVideo, 0, len(paths))
	for _, p := range paths {
		out = append(out, ParsedVideo{Path: p})
	}
	return out, nil
}

func parseSkuDir(dir string) ([]ParsedSku, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var skus []ParsedSku
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext == "" {
			continue
		}
		if _, ok := allowedSkuImageExt[ext]; !ok {
			continue
		}
		m := skuIndexPattern.FindStringSubmatch(strings.TrimSuffix(name, ext))
		if m == nil {
			return nil, fmt.Errorf("SKU 文件名格式不正确: %s（应为 SKU01_规格名(规格值).jpg）", name)
		}
		idx, _ := strconv.Atoi(m[1])
		specName, specValue, err := parseSkuSpecParts(m[2])
		if err != nil {
			return nil, fmt.Errorf("SKU 文件名格式不正确: %s（%v）", name, err)
		}
		code := fmt.Sprintf("SKU%02d", idx)
		skus = append(skus, ParsedSku{
			Index:     idx,
			Code:      code,
			SpecName:  specName,
			SpecValue: specValue,
			PicPath:   filepath.Join(dir, name),
		})
	}
	sort.Slice(skus, func(i, j int) bool { return skus[i].Index < skus[j].Index })

	specName := skus[0].SpecName
	for _, s := range skus[1:] {
		if s.SpecName != specName {
			return nil, fmt.Errorf("SKU 规格名不一致: %q 与 %q", specName, s.SpecName)
		}
	}
	seen := make(map[string]struct{}, len(skus))
	for _, s := range skus {
		if _, dup := seen[s.SpecValue]; dup {
			return nil, fmt.Errorf("SKU 规格值重复: %s", s.SpecValue)
		}
		seen[s.SpecValue] = struct{}{}
	}
	return skus, nil
}

// parseSkuSpecParts 解析「规格名(规格值)」；规格值内可含括号，如 (5选1)。
func parseSkuSpecParts(body string) (specName, specValue string, err error) {
	body = strings.TrimSpace(body)
	if body == "" || !strings.HasSuffix(body, ")") {
		return "", "", errors.New("缺少规格名或规格值")
	}
	body = strings.TrimSuffix(body, ")")
	parenIdx := strings.Index(body, "(")
	if parenIdx <= 0 {
		return "", "", errors.New("缺少规格名或规格值")
	}
	specName = strings.TrimSpace(body[:parenIdx])
	specValue = strings.TrimSpace(body[parenIdx+1:])
	if specName == "" || specValue == "" {
		return "", "", errors.New("缺少规格名或规格值")
	}
	return specName, specValue, nil
}

func sortByNumericSuffix(paths []string) {
	sort.Slice(paths, func(i, j int) bool {
		return numericKey(paths[i]) < numericKey(paths[j])
	})
}

func numericKey(path string) int {
	base := filepath.Base(path)
	matches := numSuffixPattern.FindAllString(base, -1)
	if len(matches) == 0 {
		return 0
	}
	n, _ := strconv.Atoi(matches[len(matches)-1])
	return n
}
