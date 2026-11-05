const tabState = new Map();
const STATE_KEY = "pcCollectTabState";
const persistStore =
  chrome.storage && chrome.storage.session ? chrome.storage.session : chrome.storage.local;

function persistTab(tabId) {
  const state = tabState.get(tabId);
  persistStore.get(STATE_KEY, (all) => {
    const map = (all && all[STATE_KEY]) || {};
    if (state) map[String(tabId)] = state;
    else delete map[String(tabId)];
    persistStore.set({ [STATE_KEY]: map });
  });
}

function hydrateTab(tabId, done) {
  if (tabState.has(tabId)) {
    done(tabState.get(tabId));
    return;
  }
  persistStore.get(STATE_KEY, (all) => {
    const saved = all && all[STATE_KEY] && all[STATE_KEY][String(tabId)];
    const state = saved && typeof saved === "object" ? saved : emptyState();
    tabState.set(tabId, state);
    done(state);
  });
}

function hydrateAll(done) {
  persistStore.get(STATE_KEY, (all) => {
    const map = (all && all[STATE_KEY]) || {};
    Object.keys(map).forEach((id) => {
      const tabId = Number(id);
      if (!tabState.has(tabId) && map[id]) tabState.set(tabId, map[id]);
    });
    if (done) done();
  });
}

chrome.tabs.onRemoved.addListener((tabId) => {
  tabState.delete(tabId);
  persistTab(tabId);
});

chrome.runtime.onMessage.addListener((msg, sender, sendResponse) => {
  const tabId = sender.tab && sender.tab.id;
  if (msg && msg.type === "pc-collect-harvest" && tabId) {
    hydrateTab(tabId, () => {
      mergeHarvest(tabId, msg.payload || {});
      persistTab(tabId);
      pushToTab(tabId);
      sendResponse({ ok: true });
    });
    return true;
  }
  if (msg && msg.type === "pc-collect-get") {
    hydrateTab(tabId, (state) => sendResponse(state || emptyState()));
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
    const sameItem = incoming.itemId && cur.itemId && incoming.itemId === cur.itemId;
    if (!sameItem) cur = emptyState();
  }
  if (incoming.itemId && cur.itemId && incoming.itemId !== cur.itemId) {
    cur = emptyState();
  }
  if (incoming.itemId) cur.itemId = incoming.itemId;
  if (incoming.url && /item\.taobao|detail\.tmall/i.test(incoming.url)) cur.url = incoming.url;
  if (incoming.platform && incoming.url && /item\.taobao|detail\.tmall/i.test(incoming.url)) {
    cur.platform = incoming.platform;
  }
  const trusted = storeMatchesPage(cur, incoming) || acceptClickedStore(cur, incoming);
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
  if (incoming.title && incoming.clickedTitle && (incoming.titleFromPage || trusted)) applyTitle(incoming.title);
  if (incoming.zzbTitle && cur.clickedTitle && trusted) {
    const next = String(incoming.zzbTitle)
      .replace(/[\uE000-\uF8FF]/g, "")
      .replace(/\s+/g, " ")
      .trim();
    if (!isJunkTitle(next) && next.length >= 8) cur.title = next;
  }
  if (incoming.clickedMedia) cur.clickedMedia = true;
  if (incoming.clickedSku) cur.clickedSku = true;
  if (incoming.clickedTitle) cur.clickedTitle = true;
  // 主图/视频跟「手机端主图视频SKU」走 fileList_tb，点完就收，不等 SKU 工具。
  // 视频整表覆盖（包括空数组），避免没视频的商品沿用上一件 mp4。
  if (trusted && cur.clickedMedia && incoming.fromZzbStore) {
    if (Array.isArray(incoming.images) && incoming.images.length) cur.images = incoming.images;
    if (Array.isArray(incoming.videos)) cur.videos = incoming.videos;
  }
  // SKU：当前页整表打底，再并入已抓到的 0 库存行和价格，避免无库存规格被冲掉。
  if (cur.clickedSku && Array.isArray(incoming.skus) && incoming.skus.length) {
    if (trusted && (incoming.fromZzbStore || incoming.replaceSkus)) {
      cur.skus = mergeSkuState(incoming.skus, cur.skus);
    } else {
      cur.skus = mergeSkuState(cur.skus, incoming.skus);
    }
  }
  tabState.set(tabId, cur);
}

function storeMatchesPage(cur, incoming) {
  if (incoming.replaceSkus && !incoming.fromZzbStore && !incoming.zzbItemId) return true;
  if (!incoming.fromZzbStore && !incoming.zzbTitle && !incoming.zzbItemId) return false;
  const zzb = String(incoming.zzbItemId || "");
  const page = String(incoming.itemId || (cur && cur.itemId) || "");
  return !!(zzb && page && zzb === page);
}

/** 点过主图/SKU 后，至尊宝 pack 可能还是上一件，但 fileList 已经是当前页。 */
function acceptClickedStore(cur, incoming) {
  if (!incoming.fromZzbStore) return false;
  const hasMedia = (incoming.images && incoming.images.length) || (incoming.videos && incoming.videos.length);
  const hasSkus = incoming.skus && incoming.skus.length;
  if (cur.clickedMedia && hasMedia && incoming.zzbPackMismatch) return true;
  if (cur.clickedSku && hasSkus && incoming.zzbPackMismatch) return true;
  if (cur.clickedMedia && hasMedia && !incoming.zzbItemId) return true;
  if (cur.clickedSku && hasSkus && !incoming.zzbItemId) return true;
  return false;
}

function skuStateKey(row) {
  if (!row) return "";
  const specs = row.specs;
  if (specs && typeof specs === "object") {
    const keys = Object.keys(specs).filter((k) => specs[k] != null && String(specs[k]).trim());
    if (keys.length) {
      return keys
        .slice()
        .sort()
        .map((k) => k + "=" + String(specs[k]).replace(/\s+/g, " ").trim())
        .join("\0");
    }
  }
  return String(row.specValue || row.name || "")
    .replace(/\s+/g, " ")
    .trim();
}

function mergeSkuState(oldList, nextList) {
  const byKey = new Map();
  const add = (row) => {
    if (!row) return;
    const key = skuStateKey(row);
    if (!key) return;
    const prev = byKey.get(key);
    if (!prev) {
      byKey.set(
        key,
        Object.assign({}, row, {
          specValue: row.specValue || key,
          name: row.name || row.specValue || key,
          specs: row.specs ? Object.assign({}, row.specs) : undefined,
          dimensions: row.dimensions ? row.dimensions.slice() : undefined,
        })
      );
      return;
    }
    if (row.specs && (!prev.specs || !Object.keys(prev.specs).length)) {
      prev.specs = Object.assign({}, row.specs);
      prev.dimensions = (row.dimensions || Object.keys(row.specs)).slice();
    }
    if (row.price > 0) {
      prev.price = row.price;
      prev.originalPrice = row.originalPrice || row.price;
    } else if (!(prev.price > 0) && row.originalPrice > 0) {
      prev.price = row.originalPrice;
      prev.originalPrice = row.originalPrice;
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
  hydrateAll(() => {
    tabState.forEach((state, tabId) => {
      pollOneSkuTab(state, tabId);
    });
  });
}

function pollOneSkuTab(state, tabId) {
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
        persistTab(tabId);
        pushToTab(tabId);
      }
    );
}

function scrapeSkuMainWorld() {
  const text = ((document.body && document.body.innerText) || "").replace(/\r/g, "");
  if (/资源一键下载|全选 \(/.test(text) && !/请输入SKU名称|计算价格|颜色分类[:：]/.test(text)) return [];
  if (!/原价/.test(text) || !/库存/.test(text)) {
    if (!/请输入SKU名称|颜色分类[:：]/.test(text)) return [];
  }
  const rows = [];
  const seen = {};
  const DEFAULT_SPEC_NAME = "商品规格";
  const SPEC_LABELS = [
    "颜色分类",
    "尺码分类",
    "套餐类型",
    "网络类型",
    "商品规格",
    "颜色",
    "尺码",
    "尺寸",
    "规格",
    "套餐",
    "款式",
    "型号",
    "版本",
  ];
  const escapeRegExp = (s) => String(s).replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  const looksLikeSpecLabel = (name) => {
    const n = String(name || "").trim();
    return !!(n && n.length <= 16 && /[\u4e00-\u9fff]/.test(n) && !/[A-Za-z0-9]/.test(n));
  };
  const parseSkuName = (skuName) => {
    const raw = String(skuName || "")
      .replace(/\uFEFF/g, "")
      .replace(/\t/g, "")
      .replace(/\s+/g, " ")
      .trim();
    const specs = {};
    const dimensions = [];
    const addSpec = (name, value) => {
      const n = String(name || "").trim();
      const v = String(value || "")
        .replace(/^[;；|,\s]+|[;；|,\s]+$/g, "")
        .trim();
      if (!n || !v || specs[n]) return;
      dimensions.push(n);
      specs[n] = v;
    };
    const semiParts = raw.split(/[;；]/).map((s) => s.trim()).filter(Boolean);
    if (semiParts.length > 1) {
      semiParts.forEach((part) => {
        const m = part.match(/^([^:：]{1,16})[:：](.+)$/);
        if (m && looksLikeSpecLabel(m[1])) addSpec(m[1].trim(), m[2].trim());
      });
    }
    if (dimensions.length < 2) {
      const re = new RegExp("(" + SPEC_LABELS.map(escapeRegExp).join("|") + ")\\s*[:：]\\s*", "g");
      const matches = [];
      let m;
      while ((m = re.exec(raw))) {
        matches.push({ name: m[1], index: m.index, end: m.index + m[0].length });
      }
      if (matches.length >= 2 || (matches.length === 1 && dimensions.length === 0)) {
        Object.keys(specs).forEach((k) => delete specs[k]);
        dimensions.length = 0;
        for (let i = 0; i < matches.length; i++) {
          const start = matches[i].end;
          const end = i + 1 < matches.length ? matches[i + 1].index : raw.length;
          addSpec(
            matches[i].name,
            raw.slice(start, end).replace(/^[;；|,\s]+|[;；|,\s]+$/g, "").trim()
          );
        }
      }
    }
    if (dimensions.length === 0) {
      const wrapped = raw.match(/^([^:：()（）]+)[（(](.+)[）)]$/);
      if (wrapped && looksLikeSpecLabel(wrapped[1])) addSpec(wrapped[1].trim(), wrapped[2].trim());
    }
    if (dimensions.length === 0) {
      const idxAscii = raw.indexOf(":");
      const idxFull = raw.indexOf("：");
      let splitAt = -1;
      if (idxAscii >= 0 && idxFull >= 0) splitAt = Math.min(idxAscii, idxFull);
      else if (idxAscii >= 0) splitAt = idxAscii;
      else if (idxFull >= 0) splitAt = idxFull;
      if (splitAt > 0) {
        const name = raw.slice(0, splitAt).trim();
        const value = raw.slice(splitAt + 1).trim();
        if (looksLikeSpecLabel(name)) addSpec(name, value || raw);
        else addSpec(DEFAULT_SPEC_NAME, value || raw);
      } else {
        addSpec(DEFAULT_SPEC_NAME, raw);
      }
    }
    const displayValue = dimensions.map((d) => specs[d]).join("/");
    return {
      specs,
      dimensions: dimensions.slice(),
      specName: dimensions[0] || DEFAULT_SPEC_NAME,
      specValue: displayValue || raw,
    };
  };
  const rowKey = (row) => {
    const specs = row && row.specs;
    if (specs && typeof specs === "object") {
      const keys = Object.keys(specs).filter((k) => specs[k] != null && String(specs[k]).trim());
      if (keys.length) {
        return keys
          .slice()
          .sort()
          .map((k) => k + "=" + String(specs[k]).replace(/\s+/g, " ").trim())
          .join("\0");
      }
    }
    return String((row && row.specValue) || "").trim();
  };
  const parseMoney = (v) => {
    const text = String(v == null ? "" : v).replace(/\s+/g, "").trim();
    if (!text || /^(?:-|—|–|\*|无|无库存|空|-1)$/.test(text)) return 0;
    const n = Number(text.replace(/,/g, "").replace(/[^\d.-]/g, ""));
    return isFinite(n) && n > 0 ? n : 0;
  };
  const parseStockNum = (v) => {
    const stockText = String(v == null ? "" : v).replace(/\s+/g, "").trim();
    if (!stockText || /^(?:-|—|–|\*|无|无库存|空|-1)$/.test(stockText)) return 0;
    const stockNum = Math.round(Number(stockText.replace(/,/g, "").replace(/[^\d.-]/g, "")));
    return isFinite(stockNum) && stockNum > 0 ? stockNum : 0;
  };
  const push = (name, price, stock) => {
    const raw = String(name || "")
      .replace(/\s*商品ID[:：]\s*\d+\s*/g, " ")
      .replace(/\s+/g, " ")
      .trim();
    const parsed = parseSkuName(raw);
    const specValue = parsed.specValue;
    if (!specValue || /^\d+$/.test(specValue) || /请输入SKU|计算价格设置/.test(specValue)) return;
    const keySrc = {
      specs: parsed.specs,
      specValue,
    };
    const key = rowKey(keySrc);
    if (!key) return;
    const priceNum = parseMoney(price);
    const stockNum = parseStockNum(stock);
    const prev = seen[key];
    if (prev) {
      if (priceNum > 0) {
        prev.price = priceNum;
        prev.originalPrice = priceNum;
      }
      if (stockNum > 0) prev.stock = stockNum;
      return;
    }
    const row = {
      name: raw || specValue,
      specs: parsed.specs,
      dimensions: parsed.dimensions,
      specName: parsed.specName,
      specValue,
      pic: "",
      price: priceNum,
      originalPrice: priceNum,
      stock: stockNum,
    };
    seen[key] = row;
    rows.push(row);
  };
  const harvestObjects = (list) => {
    (list || []).forEach((item) => {
      if (!item || typeof item !== "object") return;
      const name =
        item["规格名称"] ||
        item["SKU名称"] ||
        item.skuName ||
        item.name ||
        (item.specName && item.specValue ? item.specName + ":" + item.specValue : "");
      const price =
        item["原价"] ||
        item["销售价"] ||
        item["市场价"] ||
        item.originPrice ||
        item.originalPrice ||
        item.salePrice ||
        item.price;
      const stock = item["库存"] != null ? item["库存"] : item.quantity != null ? item.quantity : item.stock;
      if (name) push(name, price, stock);
    });
  };
  try {
    const cache = window.layui && window.layui.table && window.layui.table.cache;
    if (cache) Object.keys(cache).forEach((k) => harvestObjects(cache[k]));
  } catch (e) {}
  try {
    document.querySelectorAll(".el-table, .layui-table-view").forEach((el) => {
      const vue = el.__vue__;
      const data = vue && vue.store && vue.store.states && vue.store.states.data;
      if (Array.isArray(data)) harvestObjects(data);
    });
  } catch (e) {}
  const re =
    /([^\n]{2,120}[:：][^\n]{1,160})\n\s*商品ID[:：]\s*\d+\s*\n\s*(\d+(?:\.\d+)?|-)\s*\n\s*([^\n]*)\s*\n\s*([^\n]*)/g;
  let m;
  while ((m = re.exec(text))) push(m[1], parseMoney(m[2]) || parseMoney(m[3]), m[4]);
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
