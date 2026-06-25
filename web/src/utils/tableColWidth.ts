/** 按文本内容估算表格列宽（px） */
export function estimateColWidth(text: string, min = 56, max = 220): number {
  let w = 16
  for (const ch of text) {
    w += ch.charCodeAt(0) > 127 ? 14 : 8
  }
  return Math.min(max, Math.max(min, w))
}

export function colWidthFromTexts(label: string, cells: string[], min = 56, max = 220): number {
  let width = estimateColWidth(label, min, max)
  for (const text of cells) {
    width = Math.max(width, estimateColWidth(text, min, max))
  }
  return width
}

export function priceColWidth(label: string, values: (number | undefined)[], min = 52, max = 76): number {
  const cells = values.map((v) => (v ? `¥${v}` : '-'))
  return colWidthFromTexts(label, cells, min, max)
}
