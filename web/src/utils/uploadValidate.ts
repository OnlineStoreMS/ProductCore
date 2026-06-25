export interface AspectRatioRule {
  w: number
  h: number
  label: string
  /** 宽高比允许偏差，默认 5% */
  tolerance?: number
}

export interface UploadValidateRules {
  kind: 'image' | 'video'
  maxSizeMB: number
  aspectRatio?: AspectRatioRule
  acceptExt?: string[]
  /** 详情图专用 */
  minWidth?: number
  maxWidth?: number
  maxHeight?: number
  maxHeightWidthRatio?: number
}

export const MEDIA_UPLOAD_RULES = {
  main: {
    kind: 'image',
    maxSizeMB: 10,
    aspectRatio: { w: 1, h: 1, label: '1:1' },
    acceptExt: ['jpg', 'jpeg', 'png'],
  },
  pic34: {
    kind: 'image',
    maxSizeMB: 5,
    aspectRatio: { w: 3, h: 4, label: '3:4' },
    acceptExt: ['jpg', 'jpeg', 'png'],
  },
  video11: {
    kind: 'video',
    maxSizeMB: 200,
    aspectRatio: { w: 1, h: 1, label: '1:1' },
    acceptExt: ['mp4'],
  },
  video34: {
    kind: 'video',
    maxSizeMB: 200,
    aspectRatio: { w: 3, h: 4, label: '3:4' },
    acceptExt: ['mp4'],
  },
  video169: {
    kind: 'video',
    maxSizeMB: 200,
    aspectRatio: { w: 16, h: 9, label: '16:9' },
    acceptExt: ['mp4'],
  },
  video916: {
    kind: 'video',
    maxSizeMB: 200,
    aspectRatio: { w: 9, h: 16, label: '9:16' },
    acceptExt: ['mp4'],
  },
  materialWhite: {
    kind: 'image',
    maxSizeMB: 5,
    aspectRatio: { w: 1, h: 1, label: '1:1' },
    acceptExt: ['jpg', 'jpeg', 'png'],
  },
  materialTransparent: {
    kind: 'image',
    maxSizeMB: 3,
    aspectRatio: { w: 1, h: 1, label: '1:1' },
    acceptExt: ['png'],
  },
  materialGuide34: {
    kind: 'image',
    maxSizeMB: 5,
    aspectRatio: { w: 3, h: 4, label: '3:4' },
    acceptExt: ['jpg', 'jpeg', 'png'],
  },
  materialLong: {
    kind: 'image',
    maxSizeMB: 5,
    aspectRatio: { w: 2, h: 3, label: '2:3' },
    acceptExt: ['jpg', 'jpeg', 'png'],
  },
  detail: {
    kind: 'image',
    maxSizeMB: 5,
    acceptExt: ['jpg', 'jpeg', 'png'],
    minWidth: 620,
    maxWidth: 1290,
    maxHeight: 2000,
    maxHeightWidthRatio: 2,
  },
  skuSpec: {
    kind: 'image',
    maxSizeMB: 5,
    acceptExt: ['jpg', 'jpeg', 'png'],
  },
} satisfies Record<string, UploadValidateRules>

function fileExt(name: string): string {
  const i = name.lastIndexOf('.')
  return i >= 0 ? name.slice(i + 1).toLowerCase() : ''
}

function matchAspectRatio(width: number, height: number, rule: AspectRatioRule): boolean {
  if (width <= 0 || height <= 0) return false
  const actual = width / height
  const expected = rule.w / rule.h
  const tolerance = rule.tolerance ?? 0.05
  return Math.abs(actual - expected) / expected <= tolerance
}

function loadImageSize(file: File): Promise<{ width: number; height: number }> {
  return new Promise((resolve, reject) => {
    const url = URL.createObjectURL(file)
    const img = new Image()
    img.onload = () => {
      URL.revokeObjectURL(url)
      resolve({ width: img.naturalWidth, height: img.naturalHeight })
    }
    img.onerror = () => {
      URL.revokeObjectURL(url)
      reject(new Error('无法读取图片尺寸'))
    }
    img.src = url
  })
}

function loadVideoSize(file: File): Promise<{ width: number; height: number }> {
  return new Promise((resolve, reject) => {
    const url = URL.createObjectURL(file)
    const video = document.createElement('video')
    video.preload = 'metadata'
    video.onloadedmetadata = () => {
      URL.revokeObjectURL(url)
      resolve({ width: video.videoWidth, height: video.videoHeight })
    }
    video.onerror = () => {
      URL.revokeObjectURL(url)
      reject(new Error('无法读取视频尺寸'))
    }
    video.src = url
  })
}

function formatExtHint(exts: string[]): string {
  return exts.map((e) => e.toUpperCase()).join('/')
}

/** 校验通过返回 null，否则返回错误提示 */
export async function validateUploadFile(
  file: File,
  rules: UploadValidateRules,
): Promise<string | null> {
  const maxBytes = rules.maxSizeMB * 1024 * 1024
  if (file.size > maxBytes) {
    const label = rules.kind === 'video' ? '视频' : '图片'
    return `${label}大小不能超过 ${rules.maxSizeMB}MB`
  }

  if (rules.acceptExt?.length) {
    const ext = fileExt(file.name)
    const mimeOk =
      rules.kind === 'video'
        ? !file.type || file.type.startsWith('video/')
        : !file.type || file.type.startsWith('image/')
    if (!rules.acceptExt.includes(ext)) {
      const label = rules.kind === 'video' ? '视频' : '图片'
      return `${label}仅支持 ${formatExtHint(rules.acceptExt)} 格式`
    }
    if (!mimeOk) {
      return rules.kind === 'video' ? '请上传视频文件' : '请上传图片文件'
    }
  }

  let width = 0
  let height = 0
  try {
    if (rules.kind === 'image') {
      const size = await loadImageSize(file)
      width = size.width
      height = size.height
    } else {
      const size = await loadVideoSize(file)
      width = size.width
      height = size.height
    }
  } catch {
    return rules.kind === 'video' ? '无法读取视频信息，请检查文件是否有效' : '无法读取图片信息，请检查文件是否有效'
  }

  if (rules.aspectRatio && !matchAspectRatio(width, height, rules.aspectRatio)) {
    const label = rules.kind === 'video' ? '视频' : '图片'
    return `${label}宽高比需为 ${rules.aspectRatio.label}（当前 ${width}×${height}）`
  }

  if (rules.minWidth && width < rules.minWidth) {
    return `图片宽度不能小于 ${rules.minWidth}px（当前 ${width}px）`
  }
  if (rules.maxWidth && width > rules.maxWidth) {
    return `图片宽度不能大于 ${rules.maxWidth}px（当前 ${width}px）`
  }
  if (rules.maxHeight && height > rules.maxHeight) {
    return `图片高度不能大于 ${rules.maxHeight}px（当前 ${height}px）`
  }
  if (rules.maxHeightWidthRatio && height / width > rules.maxHeightWidthRatio) {
    return `图片高宽比不能大于 ${rules.maxHeightWidthRatio}（当前约 ${(height / width).toFixed(2)}）`
  }

  return null
}

export function acceptFromRules(rules: UploadValidateRules): string {
  if (rules.kind === 'video') {
    return rules.acceptExt?.includes('mp4') ? 'video/mp4,.mp4' : 'video/*'
  }
  const mimes = (rules.acceptExt || ['jpg', 'jpeg', 'png']).map((ext) => {
    if (ext === 'jpg' || ext === 'jpeg') return 'image/jpeg'
    if (ext === 'png') return 'image/png'
    return `image/${ext}`
  })
  return [...new Set(mimes)].join(',')
}
