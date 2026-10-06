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
  };
}

function mergeHarvest(tabId, incoming) {
  const cur = tabState.get(tabId) || emptyState();
  if (incoming.title && (!cur.title || incoming.title.length >= cur.title.length)) {
    cur.title = incoming.title;
  }
    if (incoming.itemId) cur.itemId = incoming.itemId;
    if (incoming.url && /item\.taobao|detail\.tmall/i.test(incoming.url)) cur.url = incoming.url;
    if (incoming.platform && incoming.url && /item\.taobao|detail\.tmall/i.test(incoming.url)) {
      cur.platform = incoming.platform;
    }
  if (incoming.clickedMedia) cur.clickedMedia = true;
  if (incoming.clickedSku) cur.clickedSku = true;
  if (Array.isArray(incoming.images) && incoming.images.length) {
    cur.images = mergeImages(cur.images, incoming.images);
  }
  if (Array.isArray(incoming.videos) && incoming.videos.length) {
    cur.videos = uniqueStrings(cur.videos.concat(incoming.videos));
  }
  if (Array.isArray(incoming.skus) && incoming.skus.length) {
    if (!cur.skus.length || incoming.skus.length >= cur.skus.length) {
      cur.skus = incoming.skus;
    }
  }
  tabState.set(tabId, cur);
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
      .trim();
    const idxAscii = text.indexOf(":");
    const idxFull = text.indexOf("：");
    let splitAt = -1;
    if (idxAscii >= 0 && idxFull >= 0) splitAt = Math.min(idxAscii, idxFull);
    else if (idxAscii >= 0) splitAt = idxAscii;
    else if (idxFull >= 0) splitAt = idxFull;
    const specValue = splitAt <= 0 ? text : text.slice(splitAt + 1).trim();
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
    const n = Number(String(price || "").replace(/[^\d.]/g, ""));
    const stockText = String(stock || "").replace(/\s+/g, "").trim();
    let stockNum = 0;
    if (stockText && !/^(?:-|—|–|\*|无|无库存|空)$/.test(stockText)) {
      stockNum = Math.round(Number(stockText.replace(/[^\d.]/g, "")));
      if (!isFinite(stockNum) || stockNum < 0) stockNum = 0;
    }
    rows.push({
      name: specValue,
      specName,
      specValue,
      pic: "",
      price: isFinite(n) ? n : 0,
      originalPrice: isFinite(n) ? n : 0,
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
