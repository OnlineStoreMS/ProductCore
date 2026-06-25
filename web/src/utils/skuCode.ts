/** 规格编码：仅字母与数字，最长 64 位 */
export const SKU_CODE_MAX_LEN = 64
export const SKU_CODE_RANDOM_LEN = 8
export const SKU_CODE_PATTERN = /^[A-Za-z0-9]+$/

/** 随机编码字符集（大写字母 + 数字，如 I2EKY7G8） */
const RANDOM_CHARS = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789'

export function sanitizeSkuCode(raw: string): string {
  return raw.replace(/[^A-Za-z0-9]/g, '').slice(0, SKU_CODE_MAX_LEN)
}

export function isValidSkuCode(code: string): boolean {
  const c = code.trim()
  return c.length > 0 && c.length <= SKU_CODE_MAX_LEN && SKU_CODE_PATTERN.test(c)
}

function randomSkuCodeString(): string {
  let code = ''
  for (let i = 0; i < SKU_CODE_RANDOM_LEN; i++) {
    code += RANDOM_CHARS[Math.floor(Math.random() * RANDOM_CHARS.length)]
  }
  return code
}

/** 生成 8 位随机规格编码，确保在本批 used 集合内不重复 */
export function nextSkuCode(used: Set<string>): string {
  for (let attempt = 0; attempt < 10000; attempt++) {
    const candidate = randomSkuCodeString()
    if (!used.has(candidate)) {
      used.add(candidate)
      return candidate
    }
  }
  throw new Error('无法生成唯一规格编码，请稍后重试')
}
