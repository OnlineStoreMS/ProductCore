package service

import (
	"fmt"
	"os"
	"strings"

	"productcore/internal/dto"
	"productcore/internal/pkg/mediainfo"
	"productcore/internal/pkg/util"
	"productcore/internal/storage"
)

func videoUploadResource(ratio string) (string, error) {
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

func assignVideoByRatio(dst *dto.ProductVideosDTO, ratio, url string) error {
	if dst == nil {
		return fmt.Errorf("视频数据为空")
	}
	switch ratio {
	case "1:1":
		if dst.Ratio11 != "" {
			return fmt.Errorf("只能有一个 1:1 视频")
		}
		dst.Ratio11 = url
	case "3:4":
		if dst.Ratio34 != "" {
			return fmt.Errorf("只能有一个 3:4 视频")
		}
		dst.Ratio34 = url
	case "16:9":
		if dst.Ratio169 != "" {
			return fmt.Errorf("只能有一个 16:9 视频")
		}
		dst.Ratio169 = url
	case "9:16":
		if dst.Ratio916 != "" {
			return fmt.Errorf("只能有一个 9:16 视频")
		}
		dst.Ratio916 = url
	default:
		return fmt.Errorf("不支持的视频比例 %s", ratio)
	}
	return nil
}

func collectRemoteVideoURLs(product *dto.ProductDTO, extra []string) []string {
	seen := map[string]struct{}{}
	var out []string
	add := func(raw string) {
		u := strings.TrimSpace(raw)
		if !isCollectableVideoURL(u) {
			return
		}
		if _, ok := seen[u]; ok {
			return
		}
		seen[u] = struct{}{}
		out = append(out, u)
	}
	for _, u := range extra {
		add(u)
	}
	if product == nil {
		return out
	}
	add(product.ProductVideo)
	if product.Media != nil && product.Media.Videos != nil {
		v := product.Media.Videos
		add(v.Ratio11)
		add(v.Ratio34)
		add(v.Ratio169)
		add(v.Ratio916)
	}
	return out
}

func isCollectableVideoURL(raw string) bool {
	u := strings.ToLower(strings.TrimSpace(raw))
	if u == "" || !strings.HasPrefix(u, "http") {
		return false
	}
	if strings.Contains(u, ".m3u8") {
		return false
	}
	return strings.Contains(u, ".mp4") ||
		strings.Contains(u, ".mov") ||
		strings.Contains(u, ".webm") ||
		strings.Contains(u, ".avi") ||
		strings.Contains(u, "cloudvideo")
}

func stripRemoteVideos(product *dto.ProductDTO) {
	if product == nil {
		return
	}
	product.ProductVideo = ""
	if product.Media != nil {
		product.Media.Videos = nil
	}
}

func (s *ProductService) ingestRemoteVideos(id uint64, urls []string) (ok int, skipped int, err error) {
	if s == nil || s.store == nil || id == 0 {
		return 0, 0, nil
	}
	urls = collectRemoteVideoURLs(nil, urls)
	if len(urls) == 0 {
		return 0, 0, nil
	}

	p, err := s.repo.GetByID(id)
	if err != nil {
		return 0, 0, err
	}

	videos := &dto.ProductVideosDTO{}
	for _, raw := range urls {
		tmp, name, dlErr := storage.DownloadRemote(raw, 0)
		if dlErr != nil {
			skipped++
			continue
		}
		uploaded, upErr := s.uploadCollectedVideo(id, tmp, name, videos)
		_ = os.Remove(tmp)
		if upErr != nil || !uploaded {
			skipped++
			continue
		}
		ok++
	}
	if ok == 0 {
		return 0, skipped, nil
	}

	media := dto.ParseProductMedia(p.MediaJSON)
	if media == nil {
		media = &dto.ProductMediaDTO{}
	}
	media.Videos = videos
	media = dto.ProductMediaWithoutPics34(media)
	mediaJSON := ""
	if media != nil {
		mediaJSON = util.ToJSON(media)
	}
	productVideo := videos.Ratio11
	if err := s.repo.UpdateFields(id, map[string]interface{}{
		"product_video": productVideo,
		"media_json":    mediaJSON,
	}); err != nil {
		return ok, skipped, err
	}
	return ok, skipped, nil
}

func (s *ProductService) uploadCollectedVideo(productID uint64, path, name string, dst *dto.ProductVideosDTO) (bool, error) {
	w, h, err := mediainfo.VideoDimensions(path)
	if err != nil {
		return false, err
	}
	ratio, err := mediainfo.DetectVideoRatio(w, h)
	if err != nil {
		return false, err
	}
	res, err := videoUploadResource(ratio)
	if err != nil {
		return false, err
	}
	url, err := s.store.UploadPath(path, name, storage.UploadOptions{
		Scope:     "product",
		ProductID: productID,
		Resource:  res,
	})
	if err != nil {
		return false, err
	}
	if err := assignVideoByRatio(dst, ratio, url); err != nil {
		return false, err
	}
	return true, nil
}
