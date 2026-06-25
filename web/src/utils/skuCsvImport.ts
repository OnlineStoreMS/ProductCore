import * as XLSX from 'xlsx'
import type { SkuItem, SkuSpec } from '../types/product'
import { nextSkuCode } from './skuCode'

export interface SkuCsvRow {
  skuName: string
  specName: string
  specValue: string
  picUrl: string
  price: number
  stock: number
  materialCode: string
}

export interface SkuCsvImportResult {
  skus: SkuItem[]
  skuSpecs: SkuSpec[]
  imported: number
  picUpdated: number
}

export type SkuExportFormat = 'csv' | 'xlsx'

const MATERIAL_CODE_MISMATCH = '资料编码不匹配，无法导入'

export const SKU_CSV_HEADERS = ['sku名称', '图片', '原价', '计算价格', '库存', 'ID'] as const

const SKU_FILE_EXTENSIONS = ['.csv', '.xlsx', '.xls'] as const

function normalizeField(value: string): string {
  return value.replace(/\uFEFF/g, '').replace(/\t/g, '').trim()
}

function cellToString(cell: unknown): string {
  if (cell == null) return ''
  if (typeof cell === 'number') {
    if (Number.isInteger(cell)) return String(cell)
    return String(cell)
  }
  if (typeof cell === 'boolean') return cell ? 'true' : 'false'
  return String(cell)
}

function parseCsvLine(line: string): string[] {
  const result: string[] = []
  let cur = ''
  let inQuotes = false
  for (let i = 0; i < line.length; i++) {
    const c = line[i]
    if (c === '"') {
      if (inQuotes && line[i + 1] === '"') {
        cur += '"'
        i++
      } else {
        inQuotes = !inQuotes
      }
    } else if (c === ',' && !inQuotes) {
      result.push(cur)
      cur = ''
    } else {
      cur += c
    }
  }
  result.push(cur)
  return result
}

/** 解析「规格名:规格值」或「规格名：规格值」 */
export function parseSkuName(skuName: string): { specName: string; specValue: string } {
  const text = normalizeField(skuName)
  const idxAscii = text.indexOf(':')
  const idxFull = text.indexOf('：')
  let splitAt = -1
  if (idxAscii >= 0 && idxFull >= 0) splitAt = Math.min(idxAscii, idxFull)
  else if (idxAscii >= 0) splitAt = idxAscii
  else if (idxFull >= 0) splitAt = idxFull
  if (splitAt <= 0) {
    return { specName: '', specValue: text }
  }
  return {
    specName: text.slice(0, splitAt).trim(),
    specValue: text.slice(splitAt + 1).trim(),
  }
}

function parseNumber(value: string): number {
  const text = normalizeField(value)
  if (!text) return 0
  const n = Number.parseFloat(text)
  return Number.isFinite(n) ? n : 0
}

function findColumnIndex(headers: string[], names: string[]): number {
  for (const name of names) {
    const idx = headers.findIndex((h) => normalizeField(h) === name)
    if (idx >= 0) return idx
  }
  return -1
}

/** 从表格行解析 SKU（CSV / Excel 共用） */
export function parseSkuRows(tableRows: string[][]): SkuCsvRow[] {
  const rows = tableRows.filter((row) => row.some((c) => normalizeField(c)))
  if (!rows.length) return []

  const headerCells = rows[0].map(normalizeField)
  const nameIdx = findColumnIndex(headerCells, ['sku名称', 'SKU名称'])
  const picIdx = findColumnIndex(headerCells, ['图片'])
  const priceIdx = findColumnIndex(headerCells, ['原价'])
  const stockIdx = findColumnIndex(headerCells, ['库存'])
  const idIdx = findColumnIndex(headerCells, ['ID', 'id'])

  if (nameIdx < 0 || priceIdx < 0 || stockIdx < 0 || idIdx < 0) {
    throw new Error('文件格式不正确，需包含 sku名称、原价、库存、ID 列')
  }

  const result: SkuCsvRow[] = []
  for (let i = 1; i < rows.length; i++) {
    const cells = rows[i]
    if (!cells.some((c) => normalizeField(c))) continue

    const skuName = cells[nameIdx] ?? ''
    const { specName, specValue } = parseSkuName(skuName)
    if (!specValue) continue

    result.push({
      skuName: normalizeField(skuName),
      specName,
      specValue,
      picUrl: picIdx >= 0 ? normalizeField(cells[picIdx] ?? '') : '',
      price: parseNumber(cells[priceIdx] ?? ''),
      stock: parseNumber(cells[stockIdx] ?? ''),
      materialCode: normalizeField(cells[idIdx] ?? ''),
    })
  }
  return result
}

/** 解析至尊宝导出的 SKU CSV 文本 */
export function parseSkuCsv(text: string): SkuCsvRow[] {
  const lines = text.split(/\r?\n/).filter((line) => line.trim())
  if (!lines.length) return []
  return parseSkuRows(lines.map(parseCsvLine))
}

/** 解析 Excel 工作簿（取第一个工作表） */
export function parseSkuExcel(buffer: ArrayBuffer): SkuCsvRow[] {
  const wb = XLSX.read(buffer, { type: 'array' })
  const sheetName = wb.SheetNames[0]
  if (!sheetName) {
    throw new Error('Excel 文件中没有工作表')
  }
  const sheet = wb.Sheets[sheetName]
  const raw = XLSX.utils.sheet_to_json(sheet, { header: 1, defval: '', raw: false }) as unknown[][]
  const tableRows = raw.map((row) => (row ?? []).map(cellToString))
  return parseSkuRows(tableRows)
}

export function isSkuImportFile(file: File): boolean {
  const name = file.name.toLowerCase()
  return SKU_FILE_EXTENSIONS.some((ext) => name.endsWith(ext))
}

/** 读取 File 为文本 */
export function readFileAsText(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result ?? ''))
    reader.onerror = () => reject(new Error('读取文件失败'))
    reader.readAsText(file, 'UTF-8')
  })
}

function readFileAsArrayBuffer(file: File): Promise<ArrayBuffer> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(reader.result as ArrayBuffer)
    reader.onerror = () => reject(new Error('读取文件失败'))
    reader.readAsArrayBuffer(file)
  })
}

/** 解析 CSV / Excel 文件 */
export async function parseSkuFile(file: File): Promise<SkuCsvRow[]> {
  if (!isSkuImportFile(file)) {
    throw new Error('不支持的文件格式，请上传 .csv、.xlsx 或 .xls')
  }
  const name = file.name.toLowerCase()
  if (name.endsWith('.csv')) {
    const text = await readFileAsText(file)
    return parseSkuCsv(text)
  }
  const buffer = await readFileAsArrayBuffer(file)
  return parseSkuExcel(buffer)
}

const DEFAULT_SPEC_NAME = '商品规格'

function normalizeSpecName(name: string): string {
  const n = normalizeField(name)
  return n || DEFAULT_SPEC_NAME
}

/** 同一规格值保留最后一行 */
function dedupeImportRows(rows: SkuCsvRow[]): SkuCsvRow[] {
  const map = new Map<string, SkuCsvRow>()
  for (const row of rows) {
    const specName = normalizeSpecName(row.specName)
    const specValue = normalizeField(row.specValue)
    if (!specValue) continue
    map.set(`${specName}\0${specValue}`, { ...row, specName, specValue })
  }
  return [...map.values()]
}

function buildSkuSpecsFromImport(
  items: Array<{ specName: string; specValue: string; pic: string }>,
): SkuSpec[] {
  const bySpec = new Map<string, { order: string[]; pics: Map<string, string> }>()
  for (const item of items) {
    const name = normalizeSpecName(item.specName)
    const val = normalizeField(item.specValue)
    if (!val) continue
    if (!bySpec.has(name)) {
      bySpec.set(name, { order: [], pics: new Map() })
    }
    const bucket = bySpec.get(name)!
    if (!bucket.pics.has(val)) {
      bucket.order.push(val)
    }
    if (item.pic) {
      bucket.pics.set(val, item.pic)
    }
  }
  return [...bySpec.entries()].map(([name, { order, pics }]) => ({
    name,
    values: order.map((value) => ({
      value,
      pic: pics.get(value) || undefined,
    })),
  }))
}

function rowMatchesProduct(row: SkuCsvRow, materialCode: string): boolean {
  const idValue = normalizeField(row.materialCode)
  if (!idValue || !materialCode) {
    return false
  }
  return idValue === materialCode
}

export interface SkuPicUploader {
  (sourceUrl: string): Promise<string>
}

/** 全量导入：按文件重建 SKU 列表、规格与规格编码 */
export async function applySkuFileImport(
  rows: SkuCsvRow[],
  productMaterialCode: string,
  uploadPic: SkuPicUploader,
): Promise<SkuCsvImportResult> {
  const materialCode = normalizeField(productMaterialCode)
  if (!materialCode) {
    throw new Error('当前商品缺少资料编码，无法导入')
  }
  const deduped = dedupeImportRows(rows)
  if (!deduped.length) {
    throw new Error('文件中没有有效的 SKU 数据')
  }

  for (const row of deduped) {
    if (!rowMatchesProduct(row, materialCode)) {
      throw new Error(
        `${MATERIAL_CODE_MISMATCH}（文件: ${row.materialCode || '空'}，当前商品: ${materialCode}）`,
      )
    }
  }

  const picCache = new Map<string, string>()
  let picUpdated = 0
  const picFailures: string[] = []

  async function resolvePic(sourceUrl: string): Promise<string> {
    const src = normalizePicUrl(sourceUrl)
    if (!src) return ''
    const cached = picCache.get(src)
    if (cached) return cached
    try {
      const uploaded = await uploadPic(src)
      picCache.set(src, uploaded)
      picUpdated++
      return uploaded
    } catch {
      picFailures.push(src)
      return ''
    }
  }

  const importItems: Array<{
    specName: string
    specValue: string
    pic: string
    price: number
    stock: number
  }> = []

  for (const row of deduped) {
    const pic = await resolvePic(row.picUrl)
    importItems.push({
      specName: normalizeSpecName(row.specName),
      specValue: row.specValue,
      pic,
      price: row.price,
      stock: row.stock,
    })
  }

  if (picFailures.length) {
    throw new Error(
      `以下图片上传失败：${picFailures.slice(0, 2).map((u) => u.slice(0, 48)).join('、')}${picFailures.length > 2 ? ' 等' : ''}`,
    )
  }

  const usedCodes = new Set<string>()
  const skus: SkuItem[] = importItems.map((item) => ({
    skuCode: nextSkuCode(usedCodes),
    specs: { [item.specName]: item.specValue },
    price: item.price,
    costPrice: 0,
    stock: item.stock,
    pic: item.pic || undefined,
  }))

  const skuSpecs = buildSkuSpecsFromImport(importItems)

  return {
    skus,
    skuSpecs,
    imported: skus.length,
    picUpdated,
  }
}

function normalizePicUrl(url: string): string {
  return normalizeField(url)
}

function escapeCsvField(value: string): string {
  return `"${value.replace(/"/g, '""')}"`
}

function formatCsvLine(fields: string[]): string {
  return `${fields.map(escapeCsvField).join(',')},`
}

/** 构建至尊宝兼容的 sku名称（规格名:规格值） */
export function buildSkuCsvName(
  sku: SkuItem,
  specColumns: string[],
): string {
  if (!specColumns.length) {
    const vals = Object.values(sku.specs || {}).filter(Boolean)
    return vals.length ? `商品规格:${vals.join(' ')}` : '商品规格:默认'
  }
  if (specColumns.length === 1) {
    const name = specColumns[0]
    const value = sku.specs[name]?.trim() || '-'
    return `${name}:${value}`
  }
  const parts = specColumns
    .map((name) => sku.specs[name]?.trim())
    .filter(Boolean)
  return `${specColumns[0]}:${parts.join(' ')}`
}

export interface SkuCsvExportOptions {
  productId: number
  materialCode: string
  skus: SkuItem[]
  specColumns: string[]
  resolvePic: (sku: SkuItem) => string
}

/** 构建表格行（CSV / Excel 共用） */
export function buildSkuTableRows(options: SkuCsvExportOptions): string[][] {
  const { materialCode, skus, specColumns, resolvePic } = options
  const idValue = normalizeField(materialCode)
  const rows: string[][] = [[...SKU_CSV_HEADERS]]

  for (const sku of skus) {
    const price = sku.price ?? 0
    const stock = sku.stock ?? 0
    rows.push([
      buildSkuCsvName(sku, specColumns),
      resolvePic(sku),
      Number.isFinite(price) ? String(price) : '',
      '',
      Number.isFinite(stock) ? String(stock) : '',
      idValue,
    ])
  }
  return rows
}

/** 生成至尊宝格式 SKU CSV（ID 列为商品 ID） */
export function buildSkuCsv(options: SkuCsvExportOptions): string {
  return buildSkuTableRows(options)
    .map((row) => formatCsvLine(row))
    .join('\n')
}

/** 生成 Excel 文件 Blob */
export function buildSkuExcelBlob(options: SkuCsvExportOptions): Blob {
  const rows = buildSkuTableRows(options)
  const ws = XLSX.utils.aoa_to_sheet(rows)
  const wb = XLSX.utils.book_new()
  XLSX.utils.book_append_sheet(wb, ws, 'SKU')
  const buffer = XLSX.write(wb, { bookType: 'xlsx', type: 'array' })
  return new Blob([buffer], {
    type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
  })
}

export function buildSkuExportFileName(productId: number): string {
  return `SKU_商品ID_${productId}`
}

export function downloadSkuExport(options: SkuCsvExportOptions, format: SkuExportFormat): void {
  if (!options.productId) {
    throw new Error('商品 ID 无效，无法导出')
  }
  if (!normalizeField(options.materialCode)) {
    throw new Error('当前商品缺少资料编码，无法导出')
  }
  const baseName = buildSkuExportFileName(options.productId)
  if (format === 'csv') {
    const csv = buildSkuCsv(options)
    const blob = new Blob(['\uFEFF', csv], { type: 'text/csv;charset=utf-8' })
    triggerDownload(blob, `${baseName}.csv`)
    return
  }
  triggerDownload(buildSkuExcelBlob(options), `${baseName}.xlsx`)
}

function triggerDownload(blob: Blob, filename: string): void {
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  link.click()
  URL.revokeObjectURL(url)
}
