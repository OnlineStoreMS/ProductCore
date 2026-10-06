const tabState = new Map();

chrome.runtime.onMessage.addListener((msg, sender, sendResponse) => {
  const tabId = sender.tab && sender.tab.id;
  if (msg && msg.type === "pc-collect-harvest" && tabId) {
    mergeHarvest(tabId, msg.payload || {});
    pushToTab(tabId);
    sendResponse({ ok: true });
    return true;
  }
  if (msg && msg.type === "pc-collect-get") {
    sendResponse(tabState.get(tabId) || emptyState());
    return true;
  }
  if (msg && msg.type === "pc-collect-upload") {
    upload(msg.payload)
      .then((result) => sendResponse(result))
      .catch((err) => sendResponse({ ok: false, error: String(err && err.message ? err.message : err) }));
    return true;
  }
  if (msg && msg.type === "pc-collect-meta") {
    loadMeta()
      .then((result) => sendResponse(result))
      .catch((err) => sendResponse({ ok: false, error: String(err && err.message ? err.message : err) }));
    return true;
  }
  return false;
});

function emptyState() {
  return {
    title: "",
    itemId: "",
    url: "",
    platform: "taobao",
    images: [],
    videos: [],
    skus: [],
    clickedMedia: false,
    clickedSku: false,
    clickedTitle: false,
  };
}

function isJunkTitle(next) {
  return (
    !next ||
    /^tb\d{4,}/i.test(next) ||
    /^[a-z]{1,4}\d{5,}(_\d+)?$/i.test(next) ||
    /^\d{4}[-/.]\d{1,2}[-/.]\d{1,2}$/.test(next) ||
    /网页无障碍|好评率|满意度|88VIP|小时发货/.test(next) ||
    (next.match(/%/g) || []).length >= 2
  );
}

function mergeHarvest(tabId, incoming) {
  let cur = tabState.get(tabId) || emptyState();
  if (incoming.resetSession) {
    cur = emptyState();
  }
  if (incoming.itemId && cur.itemId && incoming.itemId !== cur.itemId) {
    cur = emptyState();
  }
  if (incoming.clickedTitle) cur.clickedTitle = true;
  const applyTitle = (raw) => {
    const next = String(raw || "")
      .replace(/[\uE000-\uF8FF]/g, "")
      .replace(/\s*已售(?:\s*\d+\+?\s*件?|完)?\s*/g, " ")
      .replace(/\s+/g, " ")
      .trim();
    if (isJunkTitle(next)) return;
    if (!cur.title || isJunkTitle(cur.title) || incoming.clickedTitle) cur.title = next;
  };
  if (incoming.title && incoming.clickedTitle) applyTitle(incoming.title);
  if (incoming.zzbTitle && cur.clickedTitle) {
    const next = String(incoming.zzbTitle)
      .replace(/[\uE000-\uF8FF]/g, "")
      .replace(/\s+/g, " ")
      .trim();
    if (!isJunkTitle(next) && next.length >= 8) cur.title = next;
  }
  if (incoming.itemId) cur.itemId = incoming.itemId;
  if (incoming.url && /item\.taobao|detail\.tmall/i.test(incoming.url)) cur.url = incoming.url;
  if (incoming.platform && incoming.url && /item\.taobao|detail\.tmall/i.test(incoming.url)) {
    cur.platform = incoming.platform;
  }
    if (incoming.clickedMedia) cur.clickedMedia = true;
    if (incoming.clickedSku) cur.clickedSku = true;
    if (incoming.clickedTitle) cur.clickedTitle = true;
    // 主图/视频跟「手机端主图视频SKU」走 fileList_tb，点完就收，不等 SKU 工具。
    // 视频整表覆盖（包括空数组），避免没视频的商品沿用上一件 mp4。
    if (cur.clickedMedia && incoming.fromZzbStore) {
      if (Array.isArray(incoming.images) && incoming.images.length) cur.images = incoming.images;
      if (Array.isArray(incoming.videos)) cur.videos = incoming.videos;
    }
    if (cur.clickedSku && incoming.fromZzbStore && Array.isArray(incoming.skus) && incoming.skus.length) {
      cur.skus = mergeSkuState(cur.skus, incoming.skus);
    }
  tabState.set(tabId, cur);
}

function mergeSkuState(oldList, nextList) {
  const byKey = new Map();
  const add = (row) => {
    if (!row) return;
    const key = String(row.specValue || row.name || "")
      .replace(/\s+/g, " ")
      .trim();
    if (!key) return;
    const prev = byKey.get(key);
    if (!prev) {
      byKey.set(key, Object.assign({}, row, { specValue: key, name: row.name || key }));
      return;
    }
    if (row.price > 0) {
      prev.price = row.price;
      prev.originalPrice = row.originalPrice || row.price;
    }
    if (row.stock > 0) prev.stock = row.stock;
    if (row.pic && !prev.pic) prev.pic = row.pic;
  };
  (oldList || []).forEach(add);
  (nextList || []).forEach(add);
  return Array.from(byKey.values());
}

function mergeImages(oldList, nextList) {
  const bySrc = new Map();
  oldList.forEach((img) => {
    if (img && img.src) bySrc.set(img.src, img);
  });
  nextList.forEach((img) => {
    if (!img || !img.src) return;
    const prev = bySrc.get(img.src);
    if (!prev || (img.kind && img.kind !== "main" && prev.kind === "main")) {
      bySrc.set(img.src, img);
    }
  });
  return Array.from(bySrc.values());
}

function mergeVideos(oldList, nextList) {
  const bySrc = new Map();
  const add = (item) => {
    const src = typeof item === "string" ? String(item).trim() : item && item.src ? String(item.src).trim() : "";
    if (!src) return;
    const ratio = item && typeof item === "object" ? String(item.ratio || "") : "";
    const prev = bySrc.get(src) || { src, ratio: "" };
    if (ratio && !prev.ratio) prev.ratio = ratio;
    bySrc.set(src, prev);
  };
  (oldList || []).forEach(add);
  (nextList || []).forEach(add);
  return Array.from(bySrc.values());
}

function uniqueStrings(list) {
  const seen = new Set();
  const out = [];
  list.forEach((s) => {
    const v = String(s || "").trim();
    if (!v || seen.has(v)) return;
    seen.add(v);
    out.push(v);
  });
  return out;
}

function pushToTab(tabId) {
  chrome.tabs.sendMessage(tabId, { type: "pc-collect-state", state: tabState.get(tabId) }, () => {
    void chrome.runtime.lastError;
  });
}

setInterval(pollSkuFrames, 1500);

function pollSkuFrames() {
  tabState.forEach((state, tabId) => {
    if (!state.clickedSku) return;
    chrome.scripting.executeScript(
      {
        target: { tabId, allFrames: true },
        world: "MAIN",
        func: scrapeSkuMainWorld,
      },
      (results) => {
        if (chrome.runtime.lastError || !results) return;
        const rows = [];
        results.forEach((item) => {
          if (Array.isArray(item.result)) rows.push.apply(rows, item.result);
        });
        if (!rows.length) return;
        mergeHarvest(tabId, { skus: rows });
        pushToTab(tabId);
      }
    );
  });
}

function scrapeSkuMainWorld() {
  const text = ((document.body && document.body.innerText) || "").replace(/\r/g, "");
  if (/资源一键下载|全选 \(/.test(text) && !/请输入SKU名称|计算价格|颜色分类[:：]/.test(text)) return [];
  if (!/原价/.test(text) || !/库存/.test(text)) {
    if (!/请输入SKU名称|颜色分类[:：]/.test(text)) return [];
  }
  const rows = [];
  const seen = {};
  const parseSkuName = (skuName) => {
    const text = String(skuName || "")
      .replace(/\uFEFF/g, "")
      .replace(/\t/g, "")
      .replace(/\s+/g, " ")
      .trim();
    const idxAscii = text.indexOf(":");
    const idxFull = text.indexOf("：");
    let splitAt = -1;
    if (idxAscii >= 0 && idxFull >= 0) splitAt = Math.min(idxAscii, idxFull);
    else if (idxAscii >= 0) splitAt = idxAscii;
    else if (idxFull >= 0) splitAt = idxFull;
    let specValue = splitAt > 0 ? text.slice(splitAt + 1).trim() : text;
    const wrapped = specValue.match(/^([^:：()（）]+)[（(](.+)[）)]$/);
    if (wrapped) {
      const prefix = wrapped[1].trim();
      const inner = wrapped[2].trim();
      if (inner && prefix && /[\u4e00-\u9fff]/.test(prefix) && !/[A-Za-z0-9]/.test(prefix) && prefix.length <= 16) {
        specValue = inner;
      }
    }
    return { specName: "商品规格", specValue: specValue || text };
  };
  const push = (name, price, stock) => {
    const raw = String(name || "")
      .replace(/\s*商品ID[:：]\s*\d+\s*/g, " ")
      .replace(/\s+/g, " ")
      .trim();
    const parsed = parseSkuName(raw);
    const specName = "商品规格";
    const specValue = parsed.specValue;
    if (!specValue || /^\d+$/.test(specValue) || /请输入SKU|计算价格设置/.test(specValue)) return;
    const key = specName + "\0" + specValue;
    if (seen[key]) return;
    seen[key] = 1;
    const n = Number(String(price == null ? "" : price).replace(/,/g, "").replace(/[^\d.-]/g, ""));
    const priceNum = isFinite(n) && n > 0 ? n : 0;
    const stockText = String(stock || "").replace(/\s+/g, "").trim();
    let stockNum = 0;
    if (stockText && !/^(?:-|—|–|\*|无|无库存|空|-1)$/.test(stockText)) {
      stockNum = Math.round(Number(stockText.replace(/,/g, "").replace(/[^\d.-]/g, "")));
      if (!isFinite(stockNum) || stockNum < 0) stockNum = 0;
    }
    rows.push({
      name: specValue,
      specName,
      specValue,
      pic: "",
      price: priceNum,
      originalPrice: priceNum,
      stock: stockNum,
    });
  };
  const re =
    /([^\n]{2,80}[:：][^\n]{1,80})\n\s*商品ID[:：]\s*\d+\s*\n\s*(\d+(?:\.\d+)?)\s*\n\s*([^\n]*)\s*\n\s*([^\n]*)/g;
  let m;
  while ((m = re.exec(text))) push(m[1], m[2], m[4]);
  return rows;
}

async function loadMeta() {
  const auth = await authContext();
  const headers = {};
  if (auth.token) headers.Authorization = "Bearer " + auth.token;
  const [brandsRes, categoryRes] = await Promise.all([
    fetch(auth.apiBase + "/api/v1/admin/brands", { credentials: "include", headers }),
    fetch(auth.apiBase + "/api/v1/admin/categories/tree", { credentials: "include", headers }),
  ]);
  const brandsBody = await readJson(brandsRes);
  const categoryBody = await readJson(categoryRes);
  if (!brandsBody.ok) return brandsBody;
  if (!categoryBody.ok) return categoryBody;
  return { ok: true, brands: brandsBody.data || [], categories: categoryBody.data || [] };
}

async function upload(payload) {
  const auth = await authContext();
  const headers = { "Content-Type": "application/json" };
  if (auth.token) headers.Authorization = "Bearer " + auth.token;
  const res = await fetch(auth.apiBase + "/api/v1/admin/product-collects/ingest", {
    method: "POST",
    credentials: "include",
    headers,
    body: JSON.stringify(payload),
  });
  const body = await readJson(res);
  if (!body.ok) return body;
  return { ok: true, task: body.data };
}

async function authContext() {
  const cfg = await chrome.storage.local.get(["apiBase", "token"]);
  const apiBase = normalizeApiBase(cfg.apiBase);
  const token = String(cfg.token || "").trim() || (await readAccessCookie(apiBase));
  return { apiBase, token };
}

async function readJson(res) {
  const text = await res.text();
  let data = {};
  try {
    data = JSON.parse(text);
  } catch (e) {
    return { ok: false, error: "接口返回不是 JSON：" + String(text || "").slice(0, 180) };
  }
  if (!res.ok || (data.code && data.code !== 200)) {
    const hint =
      res.status === 401
        ? "未登录。请在本 Chrome 打开商品系统登录一次，Token 可留空。"
        : data.message || "请求失败 HTTP " + res.status;
    return { ok: false, error: hint };
  }
  return { ok: true, data: data.data };
}

function readAccessCookie(apiBase) {
  const urls = [];
  if (apiBase && /^https?:/i.test(apiBase)) urls.push(apiBase);
  urls.push("https://osms.zfcycle.com/", "https://osms.zfcycle.com/apps/product/");
  return new Promise((resolve) => {
    const tryNext = (i) => {
      if (i >= urls.length) {
        resolve("");
        return;
      }
      chrome.cookies.get({ url: urls[i], name: "uc_access" }, (ck) => {
        if (ck && ck.value) {
          resolve(ck.value);
          return;
        }
        tryNext(i + 1);
      });
    };
    tryNext(0);
  });
}

function normalizeApiBase(raw) {
  let s = String(raw || "").trim().replace(/\/+$/, "");
  if (!s || s === "http://127.0.0.1:8090") {
    s = "https://osms.zfcycle.com/apps/product";
  }
  s = s.replace(/\/+$/, "");
  if (/^https?:\/\/osms\.zfcycle\.com$/i.test(s)) {
    s = "https://osms.zfcycle.com/apps/product";
  }
  s = s.replace(/\/api\/v1\/admin$/i, "").replace(/\/api\/v1$/i, "");
  return s.replace(/\/+$/, "");
}
