const tabState = new Map();
const STATE_KEY = "pcCollectTabState";
const persistStore =
  chrome.storage && chrome.storage.session ? chrome.storage.session : chrome.storage.local;

function persistTab(tabId, done) {
  const state = tabState.get(tabId);
  persistStore.get(STATE_KEY, (all) => {
    const map = (all && all[STATE_KEY]) || {};
    if (state) map[String(tabId)] = state;
    else delete map[String(tabId)];
    persistStore.set({ [STATE_KEY]: map }, () => {
      if (typeof done === "function") done();
    });
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
    resolveHarvestTab(sender, msg.payload || {}, (targetId) => {
      hydrateTab(targetId, () => {
        mergeHarvest(targetId, msg.payload || {});
        persistTab(targetId);
        pushToTab(targetId);
        sendResponse({ ok: true });
      });
    });
    return true;
  }
  if (msg && msg.type === "pc-collect-get") {
    hydrateTab(tabId, (state) => sendResponse(state || emptyState()));
    return true;
  }
  if (msg && msg.type === "pc-collect-reset") {
    hydrateTab(tabId, (state) => {
      const next = emptyState();
      next.itemId = (state && state.itemId) || "";
      next.url = (state && state.url) || "";
      next.platform = (state && state.platform) || "taobao";
      tabState.set(tabId, next);
      persistTab(tabId, () => {
        pushToTab(tabId);
        sendResponse({ ok: true });
      });
    });
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
  if (msg && msg.type === "pc-collect-hook-clicks" && sender.tab && sender.tab.id) {
    injectClickHook(sender.tab.id);
  }
  return false;
});

chrome.tabs.onUpdated.addListener((tabId, info) => {
  if (info.status === "complete") injectClickHook(tabId);
});

function injectClickHook(tabId) {
  chrome.scripting.executeScript(
    {
      target: { tabId, allFrames: true },
      world: "MAIN",
      func: hookZzbToolClicks,
    },
    () => {
      void chrome.runtime.lastError;
    }
  );
}

function hookZzbToolClicks() {
  if (window.__pcCollectClickHook) return "exists";
  window.__pcCollectClickHook = true;
  const labelOf = (n) => {
    if (!n || n.nodeType !== 1) return "";
    const t = String(n.innerText || n.textContent || "").replace(/\s+/g, "").trim();
    return t && t.length <= 30 ? t : "";
  };
  const kindOf = (ev) => {
    const els = [];
    const path = ev.composedPath ? ev.composedPath() : [];
    for (let i = 0; i < path.length && i < 20; i++) {
      if (path[i] && path[i].nodeType === 1) els.push(path[i]);
    }
    if (!els.length) {
      let n = ev.target;
      for (let i = 0; i < 10 && n; i++, n = n.parentElement) {
        if (n.nodeType === 1) els.push(n);
      }
    }
    for (let i = 0; i < Math.min(els.length, 10); i++) {
      const label = labelOf(els[i]);
      if (!label) continue;
      if (/^SKU工具$/.test(label)) return "sku";
      if (/手机端主图视频SKU|主图视频SKU|手机端主图视频/.test(label)) return "media";
    }
    return "";
  };
  const fire = (ev) => {
    const kind = kindOf(ev);
    if (!kind) return;
    window.__pcCollectLastKind = kind;
    window.postMessage({ __pcCollect: 1, kind: kind }, "*");
  };
  document.addEventListener("click", fire, true);
  document.addEventListener("pointerup", fire, true);
  return "hooked";
}

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

function resolveHarvestTab(sender, payload, done) {
  const tab = sender.tab;
  const tabId = tab && tab.id;
  const url = String((tab && tab.url) || "");
  if (!tabId) return done(tabId);
  if (/item\.taobao|detail\.tmall/i.test(url)) return done(tabId);
  const opener = tab.openerTabId;
  if (opener) return done(opener);
  const itemId = String(
    (payload && (payload.itemId || payload.packItemId || payload.listItemId || payload.zzbItemId)) || ""
  );
  hydrateAll(() => {
    let found = 0;
    tabState.forEach((state, id) => {
      if (found) return;
      if (itemId && state.itemId && String(state.itemId) === itemId) found = id;
    });
    if (!found) {
      tabState.forEach((state, id) => {
        if (found) return;
        if (state.clickedMedia || state.clickedSku) found = id;
      });
    }
    done(found || tabId);
  });
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
  if (incoming.clickedMedia) cur.clickedMedia = true;
  if (incoming.clickedSku) cur.clickedSku = true;
  if (incoming.clickedTitle) cur.clickedTitle = true;
  const trusted = storeMatchesPage(cur, incoming) || acceptClickedStore(cur, incoming);
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
  if (incoming.zzbTitle && trusted && (incoming.clickedTitle || cur.clickedTitle)) {
    const next = String(incoming.zzbTitle)
      .replace(/[\uE000-\uF8FF]/g, "")
      .replace(/\s+/g, " ")
      .trim();
    if (!isJunkTitle(next) && next.length >= 8 && (!cur.title || incoming.clickedTitle)) cur.title = next;
  }
  // 主图/详情/视频：点过「手机端主图视频」后才收。ID 对得上再入库，避免点之前把上一件的 fileList 灌进来。
  if (cur.clickedMedia && incoming.fromZzbStore && trusted) {
    if (Array.isArray(incoming.images) && incoming.images.length) cur.images = incoming.images;
    if (Array.isArray(incoming.videos)) cur.videos = incoming.videos;
  }
  // SKU：点过「SKU工具」后才收。必须是当前商品的规格表 / 对得上 ID 的 pack，不能把上一件并进来。
  if (cur.clickedSku && Array.isArray(incoming.skus) && incoming.skus.length) {
    const idOk = storeMatchesPage(cur, incoming);
    const pageId = String(cur.itemId || incoming.itemId || "");
    const skus = incoming.skus.filter((row) => !row || !row.itemId || !pageId || String(row.itemId) === pageId);
    if (skus.length && (incoming.replaceSkus || (incoming.fromZzbStore && idOk))) {
      cur.skus = overlaySkuState(skus, cur.skus);
    }
  }
  tabState.set(tabId, cur);
}

function storeMatchesPage(cur, incoming) {
  if (incoming.replaceSkus && !incoming.fromZzbStore && !incoming.zzbItemId) return true;
  if (!incoming.fromZzbStore && !incoming.zzbTitle && !incoming.zzbItemId && !incoming.listItemId && !incoming.packItemId) {
    return false;
  }
  const page = String(incoming.itemId || (cur && cur.itemId) || "");
  const zzb = String(incoming.zzbItemId || incoming.listItemId || incoming.packItemId || "");
  const listId = String(incoming.listItemId || "");
  const packId = String(incoming.packItemId || "");
  if (listId && page && listId === page) return true;
  if (packId && page && packId === page) return true;
  return !!(zzb && page && zzb === page);
}

/** iframe 常没有商品 URL；点过工具，或 pack/fileList 已是当前商品时收下弹层数据。 */
function acceptClickedStore(cur, incoming) {
  if (!incoming.fromZzbStore) return false;
  const hasMedia = (incoming.images && incoming.images.length) || (incoming.videos && incoming.videos.length);
  const hasSkus = incoming.skus && incoming.skus.length;
  const page = String((cur && cur.itemId) || incoming.itemId || "");
  const listId = String(incoming.listItemId || "");
  const packId = String(incoming.packItemId || "");
  const zzb = String(incoming.zzbItemId || "");
  const clickedMedia = !!(incoming.clickedMedia || cur.clickedMedia);
  const clickedSku = !!(incoming.clickedSku || cur.clickedSku);
  if (listId && page && listId === page) return true;
  if (packId && page && packId === page) return true;
  if (clickedMedia && hasMedia && (!(cur.images && cur.images.length) || incoming.zzbPackMismatch || !zzb || (zzb && page && zzb !== page))) {
    return true;
  }
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

/** 去掉规格名后的取值指纹，用来把「颜色分类:黑」和「黑/M」对上。 */
function skuValueKey(row) {
  if (!row) return "";
  const vals = [];
  const specs = row.specs;
  if (specs && typeof specs === "object") {
    Object.keys(specs).forEach((k) => {
      const v = String(specs[k] == null ? "" : specs[k]).replace(/\s+/g, " ").trim();
      if (v) vals.push(v);
    });
  }
  if (!vals.length) {
    const sv = String(row.specValue || row.name || "")
      .replace(/\s+/g, " ")
      .trim();
    if (sv) vals.push(sv);
  }
  const parts = [];
  vals.forEach((v) => {
    String(v)
      .split(/[/／;；|]+/)
      .forEach((bit) => {
        const s = String(bit || "")
          .replace(/^[^:：]{1,16}[:：]/, "")
          .replace(/\s+/g, " ")
          .trim();
        if (s) parts.push(s);
      });
  });
  return parts.sort().join("\0");
}

function overlaySkuState(baseList, extraList) {
  const extra = new Map();
  const extraVal = new Map();
  (extraList || []).forEach((row) => {
    const key = skuStateKey(row);
    const vkey = skuValueKey(row);
    if (key) extra.set(key, row);
    if (vkey) extraVal.set(vkey, row);
  });
  return (baseList || [])
    .map((row) => {
      if (!row) return null;
      const key = skuStateKey(row);
      if (!key) return null;
      const prevExact = extra.get(key);
      const prev = prevExact || extraVal.get(skuValueKey(row));
      const out = Object.assign({}, row, {
        specValue: row.specValue || key,
        name: row.name || row.specValue || key,
        specs: row.specs ? Object.assign({}, row.specs) : undefined,
        dimensions: row.dimensions ? row.dimensions.slice() : undefined,
      });
      if (!(out.price > 0) && prev && prev.price > 0) {
        out.price = prev.price;
        out.originalPrice = prev.originalPrice || prev.price;
      }
      if (!(out.originalPrice > 0) && prev && prev.originalPrice > 0) out.originalPrice = prev.originalPrice;
      if (!(out.stock > 0) && prev && prev.stock > 0) out.stock = prev.stock;
      if (!out.pic && prevExact && prevExact.pic) out.pic = prevExact.pic;
      return out;
    })
    .filter(Boolean);
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

setInterval(pollCollectFrames, 1500);

function pollCollectFrames() {
  hydrateAll(() => {
    tabState.forEach((state, tabId) => {
      injectClickHook(tabId);
      chrome.scripting.executeScript(
        {
          target: { tabId, allFrames: true },
          world: "MAIN",
          func: () => {
            const kind = window.__pcCollectLastKind || "";
            if (kind) window.__pcCollectLastKind = "";
            return kind;
          },
        },
        (results) => {
          if (!chrome.runtime.lastError && results) {
            let kind = "";
            results.forEach((item) => {
              if (item && item.result) kind = item.result;
            });
            if (kind === "media") mergeHarvest(tabId, { clickedMedia: true, itemId: state.itemId });
            if (kind === "sku") mergeHarvest(tabId, { clickedSku: true, itemId: state.itemId });
            if (kind === "title") mergeHarvest(tabId, { clickedTitle: true, itemId: state.itemId });
            if (kind) {
              persistTab(tabId);
              pushToTab(tabId);
              state = tabState.get(tabId) || state;
            }
          }
          if (state.itemId || state.clickedMedia) pollOneMediaTab(state, tabId);
          pollOneSkuTab(state, tabId);
        }
      );
    });
  });
}

function pollOneMediaTab(state, tabId) {
  chrome.scripting.executeScript(
    {
      target: { tabId, allFrames: true },
      world: "MAIN",
      func: scrapeZzbStoreMainWorld,
    },
    (results) => {
      if (chrome.runtime.lastError || !results) return;
      let images = [];
      let videos = [];
      let skus = [];
      let listItemId = "";
      let packItemId = "";
      let zzbTitle = "";
      results.forEach((item) => {
        const data = item && item.result;
        if (!data) return;
        if (Array.isArray(data.images) && data.images.length > images.length) images = data.images;
        if (Array.isArray(data.videos) && data.videos.length >= videos.length) videos = data.videos;
        if (Array.isArray(data.skus) && data.skus.length > skus.length) skus = data.skus;
        if (data.listItemId) listItemId = data.listItemId;
        if (data.packItemId) packItemId = data.packItemId;
        if (data.zzbTitle && String(data.zzbTitle).length > String(zzbTitle).length) zzbTitle = data.zzbTitle;
      });
      if (!images.length && !videos.length && !skus.length && !zzbTitle) return;
      const packOk = !packItemId || !state.itemId || String(packItemId) === String(state.itemId);
      const listOk = !listItemId || !state.itemId || String(listItemId) === String(state.itemId);
      mergeHarvest(tabId, {
        fromZzbStore: true,
        itemId: state.itemId,
        zzbPackMismatch: !packOk,
        listItemId,
        packItemId,
        zzbItemId: listItemId || packItemId,
        images: packOk || listOk ? images : [],
        videos: packOk || listOk ? videos : [],
        skus: packOk ? skus : [],
        zzbTitle: packOk ? zzbTitle : "",
      });
      persistTab(tabId);
      pushToTab(tabId);
    }
  );
}

function scrapeZzbStoreMainWorld() {
  const skip = /favicon|sprite|\/icon|logo_|\.svg(\?|$)|zzbtool\.com|\.js(\?|$)|css(\?|$)/i;
  const parse = (raw) => {
    try {
      return raw ? JSON.parse(raw) : null;
    } catch (e) {
      return null;
    }
  };
  const readKey = (key) => parse(localStorage.getItem(key) || sessionStorage.getItem(key));
  const list = [];
  const seen = {};
  ["fileList_tb", "fileList_tm", "fileList_ali", "fileList"].forEach((key) => {
    const arr = readKey(key);
    if (!Array.isArray(arr)) return;
    arr.forEach((item) => {
      if (!item) return;
      const mark = String(item.url || "") + "\0" + String(item.cateType || "") + "\0" + String(item.name || "");
      if (seen[mark]) return;
      seen[mark] = 1;
      list.push(item);
    });
  });
  const pack = readKey("sku复制传值数据");
  const clean = (u) => {
    if (!u) return "";
    let s = String(u).trim().split(/[\s"'<>;)\\]/)[0];
    if (!s) return "";
    if (s.indexOf("//") === 0) s = "http:" + s;
    if (s.indexOf(".jpg_.webp") > 0) s = s.replace(".jpg_.webp", ".jpg");
    if (!/^https?:/i.test(s)) return "";
    return s;
  };
  const isImg = (u) => {
    if (!u || skip.test(u)) return false;
    if (!/alicdn\.com|aliimg|imgextra|tbcdn/i.test(u)) return false;
    return /\.(jpe?g|png|webp)(\?|$)/i.test(u) || /imgextra|\/bao\/uploaded\//i.test(u);
  };
  const isVid = (u) => /\.(mp4|mov|webm|avi)(\?|$)/i.test(u) || /cloudvideo/i.test(u);
  const images = [];
  const videos = [];
  const skus = [];
  const seenImg = {};
  const seenSku = {};
  const parseSkuLabel = (raw) => {
    const s = String(raw || "")
      .replace(/\uFEFF/g, "")
      .replace(/\s+/g, " ")
      .trim();
    if (!s) return null;
    const wrapped = s.match(/^([^:：()（）]{1,16})[（(](.+)[）)]$/);
    if (wrapped && /[\u4e00-\u9fff]/.test(wrapped[1]) && !/[A-Za-z0-9]/.test(wrapped[1])) {
      const name = wrapped[1].trim();
      const value = wrapped[2].trim();
      return {
        name: s,
        specName: name,
        specValue: value,
        specs: { [name]: value },
        dimensions: [name],
      };
    }
    return {
      name: s,
      specName: "商品规格",
      specValue: s,
      specs: { 商品规格: s },
      dimensions: ["商品规格"],
    };
  };
  const skuKey = (row) => {
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
  const addSku = (item, usePic) => {
    if (!item) return;
    const parsed = parseSkuLabel(item.name || item.skuName || item["规格名称"] || item["SKU名称"] || "");
    if (!parsed || !parsed.specValue) return;
    const key = skuKey(parsed);
    if (!key || seenSku[key]) return;
    seenSku[key] = 1;
    const url = usePic ? clean(item.url) : "";
    const firstPrice = (...vals) => {
      for (let i = 0; i < vals.length; i++) {
        const n = Number(vals[i]);
        if (isFinite(n) && n > 0) return n;
      }
      return 0;
    };
    const priceNum = firstPrice(
      item.originPrice,
      item.originalPrice,
      item.salePrice,
      item.skuPrice,
      item["原价"],
      item["销售价"],
      item["市场价"],
      item.price
    );
    const qty = Number(item.quantity != null ? item.quantity : item.stock);
    skus.push({
      name: parsed.name,
      specName: parsed.specName,
      specValue: parsed.specValue,
      specs: parsed.specs,
      dimensions: parsed.dimensions,
      pic: url && isImg(url) ? url : "",
      price: priceNum,
      originalPrice: priceNum,
      stock: isFinite(qty) && qty > 0 ? qty : 0,
    });
  };
  list.forEach((item) => {
    const url = clean(item.url);
    const cate = Number(item.cateType);
    const ft = Number(item.fileType);
    if (ft === 2 || cate === 4 || isVid(url)) {
      if (url && isVid(url)) videos.push({ src: url });
      if (cate === 2) addSku(item, true);
      return;
    }
    if (cate === 2) addSku(item, true);
    if (!isImg(url) || seenImg[url]) return;
    seenImg[url] = 1;
    let kind = "main";
    if (cate === 3) kind = "detail";
    else if (cate === 2) kind = "sku";
    images.push({ src: url, kind: kind });
  });
  const pushSkuList = (arr) => {
    if (!Array.isArray(arr)) return;
    arr.forEach(addSku);
  };
  pushSkuList(pack && pack.skuInfos);
  pushSkuList(pack && pack.skuList);
  pushSkuList(pack && pack.skus);
  if (pack && pack.myItemInfo) {
    pushSkuList(pack.myItemInfo.skuInfos);
    pushSkuList(pack.myItemInfo.skuList);
    pushSkuList(pack.myItemInfo.skus);
    if (pack.myItemInfo.item) {
      pushSkuList(pack.myItemInfo.item.skuInfos);
      pushSkuList(pack.myItemInfo.item.skus);
    }
  }
  const pickId = (raw) => (/^\d{5,}$/.test(String(raw || "").trim()) ? String(raw).trim() : "");
  let listItemId = "";
  list.forEach((item) => {
    if (listItemId) return;
    listItemId =
      pickId(item.itemId) ||
      pickId(item.numIid) ||
      pickId(item.num_iid) ||
      pickId(item.goodsId) ||
      ((String(item.url || "").match(/[?&]id=(\d{5,})/) || [])[1] || "");
  });
  const packItem = pack && pack.myItemInfo && pack.myItemInfo.item;
  const packItemId = String((pack && pack.itemId) || (packItem && (packItem.itemId || packItem.numIid)) || "");
  const zzbTitle = String((packItem && packItem.title) || (pack && pack.itemTitle) || "")
    .replace(/\s+/g, " ")
    .trim();
  if (pack && pack.myItemInfo) {
    const info = pack.myItemInfo;
    const addVidPack = (u) => {
      const s = clean(u);
      if (!s || !isVid(s)) return;
      if (videos.some((v) => v.src === s)) return;
      videos.push({ src: s });
    };
    addVidPack(info.pcVideo);
    if (Array.isArray(info.item && info.item.videos)) {
      info.item.videos.forEach((v) => addVidPack(v && v.url));
    }
  }
  return { images, videos, skus, listItemId, packItemId, zzbTitle };
}

function pollOneSkuTab(state, tabId) {
    if (!state.clickedSku) return;
    chrome.scripting.executeScript(
      {
        target: { tabId, allFrames: true },
        world: "MAIN",
        func: scrapeSkuMainWorld,
        args: [String(state.itemId || "")],
      },
      (results) => {
        if (chrome.runtime.lastError || !results) return;
        const byKey = new Map();
        const byVal = new Map();
        results.forEach((item) => {
          (item.result || []).forEach((row) => {
            const key = skuStateKey(row);
            const vkey = skuValueKey(row);
            if (!key) return;
            const exact = byKey.get(key);
            const prev = exact || (vkey && byVal.get(vkey));
            if (!prev) {
              byKey.set(key, row);
              if (vkey) byVal.set(vkey, row);
              return;
            }
            if (row.price > 0 && !(prev.price > 0)) {
              prev.price = row.price;
              prev.originalPrice = row.originalPrice || row.price;
            }
            if (!(prev.originalPrice > 0) && row.originalPrice > 0) prev.originalPrice = row.originalPrice;
            if (row.stock > 0) prev.stock = row.stock;
            if (row.pic && !prev.pic && exact) prev.pic = row.pic;
            if (row.specs && Object.keys(row.specs).length > Object.keys(prev.specs || {}).length) {
              prev.specs = row.specs;
              prev.dimensions = row.dimensions;
              prev.specName = row.specName;
              prev.specValue = row.specValue;
              prev.name = row.name || prev.name;
              byKey.set(skuStateKey(prev), prev);
            }
          });
        });
        const rows = Array.from(new Set(byKey.values()));
        if (!rows.length) return;
        mergeHarvest(tabId, { skus: rows, replaceSkus: true, itemId: state.itemId });
        persistTab(tabId);
        pushToTab(tabId);
      }
    );
}

function scrapeSkuMainWorld(pageItemId) {
  const wantId = String(pageItemId || "").trim();
  const text = ((document.body && document.body.innerText) || "").replace(/\r/g, "");
  const skuTool =
    /请输入SKU名称|计算价格设置|颜色分类[:：]/.test(text) ||
    (/原价/.test(text) && /库存/.test(text) && /名称/.test(text));
  const isResource = /资源一键下载|全选 \(/.test(text) && !skuTool;
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
    return isFinite(n) && n > 0 && n < 1000000 ? n : 0;
  };
  const parseStockNum = (v) => {
    const stockText = String(v == null ? "" : v).replace(/\s+/g, "").trim();
    if (!stockText || /^(?:-|—|–|\*|无|无库存|空|-1)$/.test(stockText)) return 0;
    const stockNum = Math.round(Number(stockText.replace(/,/g, "").replace(/[^\d.-]/g, "")));
    return isFinite(stockNum) && stockNum > 0 ? stockNum : 0;
  };
  const push = (name, price, stock) => {
    const rawName = String(name || "");
    const rowItemId = (rawName.match(/商品ID[:：]\s*(\d{5,})/) || [])[1] || "";
    if (wantId && rowItemId && rowItemId !== wantId) return;
    const raw = rawName
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
    if (rowItemId) row.itemId = rowItemId;
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
      const price = [
        item["原价"],
        item["销售价"],
        item["市场价"],
        item.originPrice,
        item.originalPrice,
        item.salePrice,
        item.skuPrice,
        item.price,
      ].find((v) => parseMoney(v) > 0);
      const stock = item["库存"] != null ? item["库存"] : item.quantity != null ? item.quantity : item.stock;
      const id = String(item.itemId || item.numIid || item["商品ID"] || "").trim();
      if (name) push(id && /^\d{5,}$/.test(id) ? name + " 商品ID:" + id : name, price, stock);
    });
  };
  if (skuTool && !isResource) {
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
      /([^\n]{2,120}[:：][^\n]{1,160})\n\s*商品ID[:：]\s*\d+\s*\n\s*(\d+(?:\.\d+)?|-|-1)\s*\n\s*([^\n]*)\s*\n\s*([^\n]*)/g;
    let m;
    while ((m = re.exec(text))) push(m[1], parseMoney(m[2]) || parseMoney(m[3]), m[4]);
  }
  const valueKey = (row) => {
    const vals = [];
    const specs = row && row.specs;
    if (specs && typeof specs === "object") {
      Object.keys(specs).forEach((k) => {
        const v = String(specs[k] == null ? "" : specs[k]).replace(/\s+/g, " ").trim();
        if (v) vals.push(v);
      });
    }
    if (!vals.length) {
      const sv = String((row && (row.specValue || row.name)) || "")
        .replace(/\s+/g, " ")
        .trim();
      if (sv) vals.push(sv);
    }
    const parts = [];
    vals.forEach((v) => {
      String(v)
        .split(/[/／;；|]+/)
        .forEach((bit) => {
          const s = String(bit || "")
            .replace(/^[^:：]{1,16}[:：]/, "")
            .replace(/\s+/g, " ")
            .trim();
          if (s) parts.push(s);
        });
    });
    return parts.sort().join("\0");
  };
  const fillExisting = (name, price, stock) => {
    const priceNum = parseMoney(price);
    if (!(priceNum > 0)) return;
    const raw = String(name || "")
      .replace(/\s*商品ID[:：]\s*\d+\s*/g, " ")
      .replace(/\s+/g, " ")
      .trim();
    const parsed = parseSkuName(raw);
    const specValue = parsed.specValue;
    if (!specValue) return;
    const key = rowKey({ specs: parsed.specs, specValue: specValue });
    let prev = seen[key];
    if (!prev) {
      const vkey = valueKey({ specs: parsed.specs, specValue: specValue, name: raw });
      if (vkey) {
        const keys = Object.keys(seen);
        for (let i = 0; i < keys.length; i++) {
          if (valueKey(seen[keys[i]]) === vkey) {
            prev = seen[keys[i]];
            break;
          }
        }
      }
    }
    if (prev) {
      if (!(prev.price > 0)) {
        prev.price = priceNum;
        prev.originalPrice = priceNum;
      }
      return;
    }
    if (/item\.taobao|detail\.tmall/i.test(location.href || "")) push(raw, priceNum, stock);
  };
  const tbMoney = (p) => {
    if (p == null) return 0;
    if (typeof p === "object") {
      const fromText = tbMoney(p.priceText) || tbMoney(p.price);
      if (fromText > 0) return fromText;
      if (p.priceMoney != null) {
        const n = Number(p.priceMoney);
        if (!isFinite(n) || n <= 0) return 0;
        if (n >= 1000) return n / 100;
        return n < 1000000 ? n : 0;
      }
      return 0;
    }
    return parseMoney(p);
  };
  const harvestSkuBaseCore = (root) => {
    if (!root || typeof root !== "object") return;
    const hits = { core: [], base: [] };
    const seenObj = typeof WeakSet === "function" ? new WeakSet() : null;
    const prefer = (k) => (/sku|item|data|res|loader|home|props|core|module/i.test(k) ? 0 : 1);
    const find = (obj, depth) => {
      if (!obj || typeof obj !== "object" || depth > 16) return;
      if (seenObj) {
        if (seenObj.has(obj)) return;
        seenObj.add(obj);
      }
      if (obj.sku2info && typeof obj.sku2info === "object") hits.core.push(obj);
      if (obj.skuCore && obj.skuCore.sku2info) hits.core.push(obj.skuCore);
      if (Array.isArray(obj.skus) && Array.isArray(obj.props)) hits.base.push(obj);
      if (obj.skuBase && Array.isArray(obj.skuBase.skus) && Array.isArray(obj.skuBase.props)) hits.base.push(obj.skuBase);
      if (hits.core.length && hits.base.length) return;
      if (Array.isArray(obj)) {
        const n = Math.min(obj.length, 120);
        for (let i = 0; i < n; i++) {
          try {
            find(obj[i], depth + 1);
          } catch (e) {}
          if (hits.core.length && hits.base.length) return;
        }
        return;
      }
      const keys = Object.keys(obj).sort((a, b) => prefer(a) - prefer(b));
      const limit = Math.min(keys.length, 100);
      for (let i = 0; i < limit; i++) {
        try {
          find(obj[keys[i]], depth + 1);
        } catch (e) {}
        if (hits.core.length && hits.base.length) return;
      }
    };
    find(root, 0);
    const infoMap = {};
    hits.core.forEach((core) => {
      const map = (core && core.sku2info) || {};
      Object.keys(map).forEach((id) => {
        if (id === "0") return;
        if (!infoMap[id]) infoMap[id] = map[id];
      });
    });
    hits.base.forEach((base) => {
      const vidName = {};
      (base.props || []).forEach((prop) => {
        const pid = String((prop && prop.pid) || "");
        const propName = String((prop && prop.name) || "").trim();
        ((prop && prop.values) || []).forEach((val) => {
          if (!val) return;
          const vid = String(val.vid || val.id || "");
          const name = String(val.name || val.value || "").trim();
          if (!name) return;
          const packed = propName ? propName + ":" + name : name;
          if (pid && vid) vidName[pid + ":" + vid] = packed;
          if (vid) vidName[vid] = packed;
        });
      });
      (base.skus || []).forEach((sku) => {
        if (!sku) return;
        const skuId = String(sku.skuId || sku.sku_id || "");
        if (skuId === "0") return;
        const info = infoMap[skuId] || infoMap[sku.skuId] || sku;
        const spec = String(sku.propPath || sku.props || "")
          .split(";")
          .map((pp) => vidName[pp] || vidName[String(pp).split(":").pop()] || "")
          .filter(Boolean)
          .join(";") || String(sku.propName || sku.skuName || "").trim();
        const price = tbMoney(info && info.price) || tbMoney(sku.price);
        const stock = info && info.quantity != null ? info.quantity : sku.quantity;
        if (spec) fillExisting(spec, price, stock);
      });
    });
  };
  const roots = [];
  const pushRoot = (obj) => {
    if (obj && typeof obj === "object") roots.push(obj);
  };
  try {
    const ice = window.__ICE_APP_CONTEXT__;
    const ld = ice && ice.loaderData;
    if (ld && typeof ld === "object") {
      Object.keys(ld).forEach((k) => {
        const node = ld[k];
        pushRoot(node);
        pushRoot(node && node.data);
        pushRoot(node && node.data && node.data.res);
      });
    }
  } catch (e) {}
  ["__INITIAL_DATA__", "__INITIAL_STATE__", "__ICE_APP_CONTEXT__", "__NEXT_DATA__", "Hub", "_item_detail", "g_config"].forEach(
    (key) => {
      try {
        if (window[key]) roots.push(window[key]);
      } catch (e) {}
    }
  );
  try {
    document.querySelectorAll("script").forEach((s) => {
      const t = s.textContent || "";
      if (t.length < 80 || t.length > 800000) return;
      if (!/sku2info|skuCore|skuBase/.test(t)) return;
      const start = t.indexOf("{");
      const end = t.lastIndexOf("}");
      if (start < 0 || end <= start) return;
      try {
        roots.push(JSON.parse(t.slice(start, end + 1)));
      } catch (e) {}
    });
  } catch (e) {}
  roots.forEach((root) => {
    try {
      harvestSkuBaseCore(root);
    } catch (e) {}
  });
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
