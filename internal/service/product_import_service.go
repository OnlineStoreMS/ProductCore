package service

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"productcore/internal/dto"
	"productcore/internal/pkg/mediainfo"
	"productcore/internal/pkg/productimport"
	"productcore/internal/pkg/util"
	"productcore/internal/storage"
)

const maxImportZipSize = 300 << 20 // 300MB

type ProductImportInput struct {
	Name       string
	SubTitle   string
	BrandID    uint64
	CategoryID uint64
}

type ProductImportService struct {
	products *ProductService
	store    storage.Storage
}

func NewProductImportService(products *ProductService, store storage.Storage) *ProductImportService {
	return &ProductImportService{products: products, store: store}
}

func (s *ProductImportService) ForTenant(tenantID uint64) *ProductImportService {
	cp := *s
	cp.products = s.products.ForTenant(tenantID)
	return &cp
}

func (s *ProductImportService) ImportFromZip(form ProductImportInput, zipFile *multipart.FileHeader) (*dto.ProductDTO, error) {
	name := strings.TrimSpace(form.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: 商品名称不能为空", ErrInvalidImport)
	}
	if form.BrandID == 0 {
		return nil, fmt.Errorf("%w: 请选择品牌", ErrInvalidImport)
	}
	if form.CategoryID == 0 {
		return nil, fmt.Errorf("%w: 请选择分类", ErrInvalidImport)
	}
	if zipFile == nil {
		return nil, fmt.Errorf("%w: 请上传 zip 文件", ErrInvalidImport)
	}
	if zipFile.Size > maxImportZipSize {
		return nil, fmt.Errorf("%w: zip 文件过大（最大 300MB）", ErrInvalidImport)
	}
	ext := strings.ToLower(filepath.Ext(zipFile.Filename))
	if ext != ".zip" {
		return nil, fmt.Errorf("%w: 仅支持 .zip 文件", ErrInvalidImport)
	}

	tmpZip, err := os.CreateTemp("", "import-*.zip")
	if err != nil {
		return nil, err
	}
	tmpZipPath := tmpZip.Name()
	defer func() {
		_ = tmpZip.Close()
		_ = os.Remove(tmpZipPath)
	}()

	src, err := zipFile.Open()
	if err != nil {
		return nil, err
	}
	if _, err := tmpZip.ReadFrom(src); err != nil {
		src.Close()
		return nil, err
	}
	src.Close()
	if err := tmpZip.Close(); err != nil {
		return nil, err
	}

	pkg, cleanup, err := productimport.ParseZip(tmpZipPath)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidImport, err.Error())
	}
	defer cleanup()

	if err := s.validateVideos(pkg); err != nil {
		return nil, err
	}

	productID, err := s.products.EnsureForImport(pkg.MaterialCode, name, form.BrandID, form.CategoryID)
	if err != nil {
		return nil, err
	}

	mainOpts := s.uploadOpts(productID, "import_main", 0)
	mainURLs, err := s.uploadFiles(pkg.MainPics, mainOpts)
	if err != nil {
		return nil, err
	}
	detailOpts := s.uploadOpts(productID, "import_detail", 0)
	detailURLs, err := s.uploadFiles(pkg.DetailPics, detailOpts)
	if err != nil {
		return nil, err
	}

	videosDTO := &dto.ProductVideosDTO{}
	for _, v := range pkg.Videos {
		res, err := importVideoResource(v.Ratio)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrInvalidImport, err.Error())
		}
		videoOpts := s.uploadOpts(productID, res, 0)
		url, err := s.store.UploadPath(v.Path, filepath.Base(v.Path), videoOpts)
		if err != nil {
			return nil, fmt.Errorf("上传视频失败: %w", err)
		}
		if err := assignImportVideo(videosDTO, v.Ratio, url); err != nil {
			return nil, fmt.Errorf("%w: %s", ErrInvalidImport, err.Error())
		}
	}

	skuDTOs := make([]dto.SkuDTO, 0, len(pkg.Skus))
	specValues := make([]dto.SkuSpecValueDTO, 0, len(pkg.Skus))
	specName := pkg.Skus[0].SpecName
	valuePics := make(map[string]string, len(pkg.Skus))

	for _, item := range pkg.Skus {
		skuOpts := s.uploadOpts(productID, "import_sku", 0)
		picURL, err := s.store.UploadPath(item.PicPath, filepath.Base(item.PicPath), skuOpts)
		if err != nil {
			return nil, fmt.Errorf("上传 SKU 图片失败: %w", err)
		}
		code := util.SanitizeSkuCode(pkg.MaterialCode + item.Code)
		if !util.IsValidSkuCode(code) {
			return nil, fmt.Errorf("%w: SKU 编码无效 %s", ErrInvalidImport, item.Code)
		}
		skuDTOs = append(skuDTOs, dto.SkuDTO{
			SkuCode: code,
			Specs:   map[string]string{specName: item.SpecValue},
			Price:   0,
			Stock:   0,
			Weight:  0,
			Pic:     picURL,
		})
		valuePics[item.SpecValue] = picURL
	}

	for _, item := range pkg.Skus {
		specValues = append(specValues, dto.SkuSpecValueDTO{
			Value: item.SpecValue,
			Pic:   valuePics[item.SpecValue],
		})
	}

	media := &dto.ProductMediaDTO{DetailPics: detailURLs}
	if videosDTO.Ratio11 != "" || videosDTO.Ratio34 != "" || videosDTO.Ratio169 != "" || videosDTO.Ratio916 != "" {
		media.Videos = videosDTO
	}

	albumPics := mainURLs
	pic := ""
	if len(mainURLs) > 0 {
		pic = mainURLs[0]
		if len(mainURLs) > 1 {
			albumPics = mainURLs[1:]
		} else {
			albumPics = nil
		}
	}

	in := &dto.ProductDTO{
		Name:           name,
		SubTitle:       strings.TrimSpace(form.SubTitle),
		MaterialCode:   pkg.MaterialCode,
		Source:         pkg.Source,
		BrandID:        form.BrandID,
		CategoryID:     form.CategoryID,
		Pic:            pic,
		AlbumPics:      albumPics,
		Media:          media,
		Unit:           "件",
		PublishStatus:  0,
		VerifyStatus:   1,
		SkuSpecs:       []dto.SkuSpecDTO{{Name: specName, Values: specValues}},
		Skus:           skuDTOs,
		ChannelVisible: "both",
	}

	in.Finalize = true
	return s.products.Update(productID, in)
}

func (s *ProductImportService) validateVideos(pkg *productimport.ParsedPackage) error {
	for i := range pkg.Videos {
		v := &pkg.Videos[i]
		w, h, err := mediainfo.VideoDimensions(v.Path)
		if err != nil {
			return fmt.Errorf("%w: %s（%s）", ErrInvalidImport, err.Error(), filepath.Base(v.Path))
		}
		ratio, err := mediainfo.DetectVideoRatio(w, h)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrInvalidImport, err.Error())
		}
		v.Ratio = ratio
	}
	return nil
}

func importVideoResource(ratio string) (string, error) {
	switch ratio {
	case "1:1":
		return "import_video_11", nil
	case "3:4":
		return "import_video_34", nil
	case "16:9":
		return "import_video_169", nil
	case "9:16":
		return "import_video_916", nil
	default:
		return "", fmt.Errorf("不支持的视频比例 %s", ratio)
	}
}

func assignImportVideo(dst *dto.ProductVideosDTO, ratio, url string) error {
	switch ratio {
	case "1:1":
		if dst.Ratio11 != "" {
			return fmt.Errorf("视频文件夹中只能有一个 1:1 视频")
		}
		dst.Ratio11 = url
	case "3:4":
		if dst.Ratio34 != "" {
			return fmt.Errorf("视频文件夹中只能有一个 3:4 视频")
		}
		dst.Ratio34 = url
	case "16:9":
		if dst.Ratio169 != "" {
			return fmt.Errorf("视频文件夹中只能有一个 16:9 视频")
		}
		dst.Ratio169 = url
	case "9:16":
		if dst.Ratio916 != "" {
			return fmt.Errorf("视频文件夹中只能有一个 9:16 视频")
		}
		dst.Ratio916 = url
	default:
		return fmt.Errorf("不支持的视频比例 %s", ratio)
	}
	return nil
}

func (s *ProductImportService) uploadOpts(productID uint64, resource string, skuID uint64) storage.UploadOptions {
	return storage.UploadOptions{
		Scope:     "product",
		ProductID: productID,
		Resource:  resource,
		SkuID:     skuID,
	}
}

func (s *ProductImportService) uploadFiles(paths []string, opts storage.UploadOptions) ([]string, error) {
	urls := make([]string, 0, len(paths))
	for _, p := range paths {
		url, err := s.store.UploadPath(p, filepath.Base(p), opts)
		if err != nil {
			return nil, fmt.Errorf("上传 %s 失败: %w", filepath.Base(p), err)
		}
		urls = append(urls, url)
	}
	return urls, nil
}
