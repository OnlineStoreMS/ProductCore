import type { UploadValidateRules } from './uploadValidate'
import type { ImageAdaptParams } from '../api/upload'

/** 电商清晰度：足够清晰、体积可控（素材原图往往 4K+/数 MB） */
export function adaptParamsFromRules(
  sourceUrl: string,
  rules?: UploadValidateRules,
  crop?: { x: number; y: number; w: number; h: number },
): ImageAdaptParams {
  const params: ImageAdaptParams = {
    url: sourceUrl,
    quality: 82,
  }
  if (crop && crop.w > 0 && crop.h > 0) {
    params.cropX = crop.x
    params.cropY = crop.y
    params.cropW = crop.w
    params.cropH = crop.h
  }
  if (!rules || rules.kind !== 'image') {
    params.maxEdge = 1200
    return params
  }

  if (rules.aspectRatio) {
    params.aspectW = rules.aspectRatio.w
    params.aspectH = rules.aspectRatio.h
  }

  // 透明素材保留 PNG
  if (rules.acceptExt?.length === 1 && rules.acceptExt[0] === 'png') {
    params.preferPng = true
    params.quality = 0
  }

  // 详情图：限宽高，不强制比例
  if (rules.maxWidth || rules.maxHeight || rules.minWidth) {
    params.detailMaxW = rules.maxWidth || 1200
    params.detailMaxH = rules.maxHeight || 2000
    params.maxEdge = 0
    return params
  }

  // 主图 1:1 → 1200；SKU/白底等 1:1 → 800；3:4 → 长边 1200；2:3 → 1200
  const label = rules.aspectRatio?.label || ''
  if (label === '1:1') {
    params.maxEdge = rules.maxSizeMB >= 8 ? 1200 : 800
  } else if (label === '3:4' || label === '2:3') {
    params.maxEdge = 1200
  } else {
    params.maxEdge = 1000
  }
  return params
}
