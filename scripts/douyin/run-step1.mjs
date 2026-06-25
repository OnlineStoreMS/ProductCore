#!/usr/bin/env node
/**
 * 抖店铺货 Step1 — 通过 streamable-mcp-server 自动执行
 *
 * 前置条件：
 *   1. Chrome 已安装 mcp-chrome 扩展，Bridge 运行在 http://127.0.0.1:12306/mcp
 *   2. 已登录抖店，扩展已连接
 *   3. 若 Cursor 已占用 MCP 连接，请先断开 Cursor 的 streamable-mcp-server 再运行本脚本
 *
 * 用法：
 *   cd scripts/douyin && npm install && node run-step1.mjs
 *   node run-step1.mjs --title "公路自行车碟刹来令片RT70" --image ../../data/poc/douyin/main01.jpg
 *   node run-step1.mjs --no-next          # 只填标题+主图，不点下一步
 *   node run-step1.mjs --tab-id 1991679424
 */

import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, resolve } from 'node:path';
import { Client } from '@modelcontextprotocol/sdk/client/index.js';
import { StreamableHTTPClientTransport } from '@modelcontextprotocol/sdk/client/streamableHttp.js';

const __dirname = dirname(fileURLToPath(import.meta.url));

const DEFAULTS = {
  mcpUrl: process.env.MCP_URL || 'http://127.0.0.1:12306/mcp',
  title: '公路自行车碟刹来令片RT70国产140mm配内锁盖',
  image: resolve(__dirname, '../../data/poc/douyin/main01.jpg'),
  createUrl: 'https://fxg.jinritemai.com/ffa/g/create',
  clickNext: true,
  tabId: undefined,
  waitMs: 5000,
};

function parseArgs(argv) {
  const opts = { ...DEFAULTS };
  for (let i = 2; i < argv.length; i++) {
    const a = argv[i];
    if (a === '--no-next') opts.clickNext = false;
    else if (a === '--title') opts.title = argv[++i];
    else if (a === '--image') opts.image = resolve(process.cwd(), argv[++i]);
    else if (a === '--image-url') opts.imageUrl = argv[++i];
    else if (a === '--tab-id') opts.tabId = Number(argv[++i]);
    else if (a === '--mcp-url') opts.mcpUrl = argv[++i];
    else if (a === '--wait') opts.waitMs = Number(argv[++i]);
    else if (a === '--help' || a === '-h') {
      console.log(`
用法: node run-step1.mjs [选项]

  --title <文本>       商品标题
  --image <路径>       本地主图（转为 data URL 注入页面）
  --image-url <url>    远程主图 URL（与 --image 二选一）
  --tab-id <数字>      指定 Chrome 标签页 ID
  --no-next            不点击「下一步」
  --wait <毫秒>        上传后等待时间，默认 5000
  --mcp-url <url>      MCP 地址，默认 http://127.0.0.1:12306/mcp
`);
      process.exit(0);
    }
  }
  return opts;
}

function buildStep1Code(opts) {
  let imageUrl;
  if (opts.imageUrl) {
    imageUrl = opts.imageUrl;
  } else {
    const buf = readFileSync(opts.image);
    const b64 = buf.toString('base64');
    const mime = opts.image.toLowerCase().endsWith('.png') ? 'image/png' : 'image/jpeg';
    imageUrl = `data:${mime};base64,${b64}`;
  }

  const inline = readFileSync(resolve(__dirname, 'step1-inline.js'), 'utf8');
  const patched = inline
    .replace("'__TITLE__'", JSON.stringify(opts.title))
    .replace("'__IMAGE_URL__'", JSON.stringify(imageUrl))
    .replace("'__CREATE_URL__'", JSON.stringify(opts.createUrl))
    .replace('__CLICK_NEXT__', String(opts.clickNext))
    .replace('__WAIT_MS__', String(opts.waitMs));

  return patched;
}

function extractText(result) {
  if (!result?.content) return result;
  const text = result.content
    .filter((c) => c.type === 'text')
    .map((c) => c.text)
    .join('\n');
  try {
    const parsed = JSON.parse(text);
    if (parsed.result !== undefined) return parsed.result;
    return parsed;
  } catch {
    return text;
  }
}

async function main() {
  const opts = parseArgs(process.argv);
  console.log('MCP:', opts.mcpUrl);
  console.log('标题:', opts.title);
  console.log('主图:', opts.imageUrl || opts.image);
  console.log('点下一步:', opts.clickNext);

  const transport = new StreamableHTTPClientTransport(new URL(opts.mcpUrl));
  const client = new Client({ name: 'douyin-step1', version: '1.0' });

  try {
    await client.connect(transport);
  } catch (e) {
    console.error('\n无法连接 MCP。常见原因：');
    console.error('  - Chrome 扩展 / Bridge 未启动');
    console.error('  - Cursor 已占用连接 → 在 Cursor 设置里暂时禁用 streamable-mcp-server 后重试');
    console.error('错误:', e.message);
    process.exit(1);
  }

  try {
    if (opts.tabId) {
      const nav = await client.callTool({
        name: 'chrome_switch_tab',
        arguments: { tabId: opts.tabId },
      });
      console.log('切换标签:', extractText(nav));
    } else {
      const tabs = await client.callTool({ name: 'get_windows_and_tabs', arguments: {} });
      const info = extractText(tabs);
      const douyin = info?.windows?.flatMap((w) => w.tabs || []).find((t) =>
        t.url?.includes('fxg.jinritemai.com')
      );
      if (douyin) {
        console.log('使用抖店标签 tabId=', douyin.tabId, douyin.url?.slice(0, 60));
        await client.callTool({
          name: 'chrome_switch_tab',
          arguments: { tabId: douyin.tabId },
        });
      } else {
        console.log('未找到抖店标签，导航到发布页...');
        await client.callTool({
          name: 'chrome_navigate',
          arguments: { url: opts.createUrl },
        });
      }
    }

    const code = buildStep1Code(opts);
    const jsResult = await client.callTool({
      name: 'chrome_javascript',
      arguments: {
        code,
        timeoutMs: 60000,
        tabId: opts.tabId,
      },
    });

    const out = extractText(jsResult);
    console.log('\n=== Step1 结果 ===');
    console.log(JSON.stringify(out, null, 2));

    if (out?.nextDisabled && opts.clickNext) {
      console.warn('\n警告: 「下一步」仍不可点，请检查主图是否上传成功。');
      process.exit(2);
    }
  } finally {
    await client.close();
  }
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
