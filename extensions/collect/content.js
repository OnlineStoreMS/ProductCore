(() => {
  const SKIP = /favicon|sprite|\/icon|logo_|\.svg(\?|$)|zzbtool\.com|\.js(\?|$)|\.css(\?|$)/i;
  const isTop = window === window.top;
  const href = location.href || "";

  document.addEventListener(
    "click",
    (ev) => {
      const t = ((ev.target && (ev.target.innerText || ev.target.textContent)) || "").replace(/\s+/g, "");
      if (t.indexOf("手机端主图视频SKU") >= 0 || t.indexOf("主图视频SKU") >= 0) {
        report({ clickedMedia: true });
      }
      if (t.indexOf("SKU工具") >= 0) {
        report({ clickedSku: true });
      }
      if (t.indexOf("悬浮标题") >= 0 || (t.indexOf("标题采集") >= 0 && t.indexOf("主图") < 0)) {
        report({ clickedTitle: true });
      }
    },
    true
  );

  setInterval(scan, 1500);
  scan();

  if (isTop && /item\.taobao\.com|detail\.tmall\.com/i.test(href)) {
    mountPanel();
  }

  function scan() {
    const payload = {
      platform: /tmall/i.test(href) ? "tmall" : "taobao",
    };
    if (/item\.taobao\.com|detail\.tmall\.com/i.test(href)) {
      payload.url = href;
    }
    Object.assign(payload, extractTitle());
    const media = extractMedia();
    if (media.images.length || media.videos.length) {
      payload.images = media.images;
      payload.videos = media.videos;
    }
    const skus = extractSkus();
    if (skus.length) payload.skus = skus;
    if (payload.title || payload.images || payload.skus || payload.videos) {
      report(payload);
    }
  }

  function looksLikeTitle(raw) {
    const t = cleanTitle(raw);
    if (t.length < 8 || t.length > 200) return "";
    if (
      /验证码|评价|销量|收藏|购物车|登录|资源一键下载|手机端主图|SKU工具|数据工具|全选 \(|详情页/.test(
        t
      )
    ) {
      return "";
    }
    return t;
  }

  function cleanTitle(raw) {
    let t = String(raw || "")
      .replace(/\uFEFF/g, "")
      .replace(/\s+/g, " ")
      .trim();
    t = t.replace(/[-|｜]?\s*(淘宝网|天猫|tmall\.com).*$/i, "").trim();
    t = t.replace(/[，,。.\s]*已售(?:\s*\d+\+?\s*件?|完)?\s*/g, " ");
    t = t.replace(/[，,。.\s]*(月销|销量|付款|人付款)\s*\d+\+?\s*/g, " ");
    return t.replace(/\s+/g, " ").trim();
  }

  function extractTitle() {
    const out = {};
    const id = (href.match(/[?&]id=(\d+)/) || [])[1] || "";
    if (id) out.itemId = id;
    const zzb = extractZzbTitle();
    if (zzb) {
      out.title = zzb.slice(0, 200);
      out.fromZzb = true;
    }
    return out;
  }

  function extractZzbTitle() {
    const cands = [];
    collectRoots().forEach((root) => {
      if (!root.querySelectorAll) return;
      const layers = Array.from(
        root.querySelectorAll(
          ".layui-layer, .layui-layer-content, .el-dialog, .el-message-box, .ant-modal"
        )
      );
      const scopes = layers.length ? layers : [root];
      scopes.forEach((scope) => {
        if (scope.id === "pc-collect-panel" || (scope.closest && scope.closest("#pc-collect-panel"))) return;
        const overlay = String(scope.innerText || "").replace(/\s+/g, " ");
        const inTitleUi = layers.length > 0 || /悬浮标题|标题采集|商品标题/.test(overlay);
        scope.querySelectorAll("input, textarea").forEach((el) => {
          if (el.closest && el.closest("#pc-collect-panel")) return;
          const ph = String(el.placeholder || el.name || el.getAttribute("aria-label") || "");
          if (/SKU|搜索|查询|价格|库存|品牌|分类/.test(ph)) return;
          if (!inTitleUi && !/标题/.test(ph)) return;
          const t = looksLikeTitle(el.value);
          if (t) cands.push(t);
        });
        const text = String(scope.innerText || "").replace(/\r/g, "");
        const labeled = text.match(/(?:商品标题|标题)[:：]\s*([^\n]{8,200})/);
        if (labeled) {
          const t = looksLikeTitle(labeled[1]);
          if (t) cands.push(t);
        }
      });
    });
    cands.sort((a, b) => b.length - a.length);
    return cands[0] || "";
  }

  function extractMedia() {
    const onTool = /oneDownload/i.test(href);
    const text = bodyText();
    const looks = onTool || /资源一键下载|全选 \(/.test(text);
    const images = [];
    const videos = [];
    const seen = {};
    const addImg = (src, kind) => {
      const u = clean(src);
      if (!u || seen[u] || !isProductImage(u)) return;
      seen[u] = 1;
      images.push({ src: u, kind: kind || "main" });
    };
    const addVid = (src) => {
      const u = cleanVideo(src);
      if (!u) return;
      for (let i = 0; i < videos.length; i++) {
        if (videos[i] === u || (videos[i] && videos[i].src === u)) return;
      }
      videos.push({ src: u });
    };
    if (!looks) return { images, videos };

    const scanRow = (el) => {
      const line = ((el.innerText || "") + " " + (el.getAttribute("class") || "")).replace(/\s+/g, " ");
      if (/教程|导出文件|批量下载/.test(line) && !el.querySelector("img, video, a[href]")) return;
      const vid = firstVideo(el);
      if (vid || isVideoLine(line)) {
        addVid(vid || pickSrc(el.querySelector("a[href]")));
        return;
      }
      const img = el.querySelector("img");
      const link = el.querySelector("a[href]");
      const src = pickSrc(img) || pickSrc(link) || firstAlicdn(el.innerHTML || "");
      if (!src) return;
      addImg(src, kindFromLine(line));
    };

    document.querySelectorAll("table tr, .el-table__row, .layui-table tr, [class*='oneDownload'] li").forEach(scanRow);

    if (images.length === 0) {
      document.querySelectorAll("img, a[href]").forEach((el) => {
        const row = el.closest("tr, .el-table__row, li, [class*='row']") || el.parentElement;
        const line = row ? (row.innerText || "") : "";
        if (isVideoLine(line) || isVideoURL(pickSrc(el))) {
          addVid(pickSrc(el));
          return;
        }
        addImg(pickSrc(el), kindFromLine(line));
      });
    }
    document.querySelectorAll("video").forEach((el) => addVid(el.currentSrc || el.src || pickSrc(el)));
    if (!videos.length) {
      collectVideoURLsFromHTML(document.body && document.body.innerHTML).forEach(addVid);
    }
    return { images, videos };
  }

  function kindFromLine(line) {
    const t = String(line || "");
    if (isVideoLine(t)) return "video";
    if (/详情页|详情图|详情0/.test(t) || (/详情/.test(t) && !/主图/.test(t))) return "detail";
    if (/SKU\s*\(|SKU0|SKU图|来源\s*SKU/.test(t) || (/SKU/.test(t) && !/主图|手机端/.test(t))) return "sku";
    if (/主图/.test(t)) return "main";
    return "main";
  }

  function isVideoLine(line) {
    const t = String(line || "").replace(/\s+/g, "");
    if (/手机端主图视频SKU|主图视频SKU/.test(t)) return false;
    return /来源视频|主图视频|商品视频|^视频$|视频\(/.test(t);
  }

  function firstVideo(el) {
    if (!el) return "";
    const v = el.querySelector && el.querySelector("video");
    if (v) return v.currentSrc || v.src || v.getAttribute("src") || "";
    const links = el.querySelectorAll ? el.querySelectorAll("a[href], source[src], video") : [];
    for (let i = 0; i < links.length; i++) {
      const u = pickSrc(links[i]);
      if (isVideoURL(u)) return u;
    }
    const html = el.innerHTML || "";
    const found = collectVideoURLsFromHTML(html);
    return found[0] || "";
  }

  function collectVideoURLsFromHTML(html) {
    const out = [];
    const re = /https?:\/\/[^"'\\\s<>]+/gi;
    const text = String(html || "");
    let m;
    while ((m = re.exec(text))) {
      if (isVideoURL(m[0])) out.push(m[0]);
    }
    return out;
  }

  function isVideoURL(u) {
    const s = String(u || "");
    if (!s || /\.m3u8(\?|$)/i.test(s)) return false;
    return /\.(mp4|mov|webm|avi)(\?|$)/i.test(s) || /cloudvideo/i.test(s);
  }

  function cleanVideo(u) {
    if (!u) return "";
    let s = String(u).trim();
    if (s.indexOf("//") === 0) s = "https:" + s;
    s = s.split(/[\s"'<>;)\\]/)[0];
    if (!/^https?:/i.test(s) || !isVideoURL(s)) return "";
    return s;
  }

  function isProductImage(u) {
    if (SKIP.test(u) || !/\.(jpg|jpeg|png|webp|gif)(\?|$)/i.test(u)) return false;
    return /alicdn\.com|aliimg|imgextra|tbcdn|cloudvideo/i.test(u);
  }

  function firstAlicdn(html) {
    const m = String(html || "").match(/https?:\/\/[^"'\\\s<>]+(?:alicdn|imgextra)[^"'\\\s<>]+/i);
    return m ? m[0] : "";
  }

  function extractSkus() {
    if (!isSkuToolPage()) return [];
    const rows = [];
    const seen = {};
    const push = (name, pic, price, original, stock) => {
      const raw = String(name || "")
        .replace(/\uFEFF/g, "")
        .replace(/\t/g, "")
        .replace(/\s*商品ID[:：]\s*\d+\s*/g, " ")
        .replace(/\s+/g, " ")
        .trim();
      const parsed = parseSkuName(raw);
      const specName = DEFAULT_SPEC_NAME;
      const specValue = parsed.specValue;
      if (!specValue || /^\d+$/.test(specValue)) return;
      if (/^(规格名称|销售价|市场价|原价|库存|商家编码|名称|计算价格)$/.test(specValue)) return;
      if (/查询中|加载中|请输入SKU/.test(specValue) || /查询中|加载中|请输入SKU/.test(specName)) return;
      const key = specName + "\0" + specValue + "\0" + String(pic || "");
      const row = {
        name: specValue,
        specName,
        specValue,
        pic: clean(pic),
        price: num(price),
        originalPrice: num(original),
        stock: parseStock(stock),
      };
      if (seen[key] && !(row.price > 0 && !(seen[key].price > 0))) return;
      seen[key] = row;
      rows.push(row);
    };
    collectRoots().forEach((root) => {
      root.querySelectorAll(".layui-table-view, .layui-table-box, .el-table, .ant-table").forEach((wrap) =>
        scanSkuTable(wrap, push)
      );
      root.querySelectorAll("table").forEach((table) => {
        if (table.closest && table.closest(".el-table, .layui-table-view, .layui-table-box, .ant-table")) return;
        scanSkuTable(table, push);
      });
    });
    parseSkuByText(push);
    return dedupeSkuRows(rows);
  }

  function isSkuToolPage() {
    const text = overlayText();
    if (/资源一键下载|全选 \(/.test(text) && !/请输入SKU名称|计算价格/.test(text)) return false;
    return (
      /请输入SKU名称|计算价格设置|颜色分类[:：]/.test(text) ||
      (/原价/.test(text) && /库存/.test(text) && /名称/.test(text))
    );
  }

  function parseSkuByText(push) {
    const raw = ((document.body && document.body.innerText) || "").replace(/\r/g, "");
    const re =
      /([^\n]{2,80}[:：][^\n]{1,80})\n\s*商品ID[:：]\s*\d+\s*\n\s*(\d+(?:\.\d+)?)\s*\n\s*([^\n]*)\s*\n\s*([^\n]*)/g;
    let m;
    while ((m = re.exec(raw))) {
      push(m[1], "", m[2], m[2], m[4]);
    }
  }

  function overlayText() {
    try {
      return ((document.body && document.body.innerText) || "").replace(/\s+/g, " ").slice(0, 12000);
    } catch (e) {
      return "";
    }
  }

  function collectRoots() {
    const out = [];
    const visit = (root) => {
      if (!root) return;
      out.push(root);
      const nodes = root.querySelectorAll ? root.querySelectorAll("*") : [];
      for (let i = 0; i < nodes.length; i++) {
        if (nodes[i].shadowRoot) visit(nodes[i].shadowRoot);
      }
    };
    visit(document);
    return out;
  }

  function scanSkuTable(root, push) {
    const headers = skuHeaders(root);
    const nameCol = colIndex(headers, ["规格名称", "SKU名称", "sku名称", "规格信息", "规格", "名称"]);
    const originCol = colIndex(headers, ["原价"]);
    const saleCol = colIndex(headers, ["销售价"]);
    const marketCol = colIndex(headers, ["市场价"]);
    const stockCol = colIndex(headers, ["库存"]);
    if (nameCol < 0 && originCol < 0 && saleCol < 0) return;
    skuBodyRows(root).forEach((tr) => {
      const cells = skuCells(tr);
      if (cells.length < 2) return;
      const img = tr.querySelector("img");
      let name = nameCol >= 0 ? cellValue(cells[nameCol]) : "";
      if (!looksLikeSpecName(name)) {
        name = "";
        for (let i = 0; i < cells.length; i++) {
          const t = cellValue(cells[i]);
          if (looksLikeSpecName(t)) {
            name = t;
            break;
          }
        }
      }
      const origin = originCol >= 0 ? moneyIn(cells[originCol]) : 0;
      const sale = saleCol >= 0 ? moneyIn(cells[saleCol]) : 0;
      const market = marketCol >= 0 ? moneyIn(cells[marketCol]) : 0;
      const stock = stockCol >= 0 ? cellValue(cells[stockCol]) : "";
      push(name, img ? pickSrc(img) : "", origin || sale, market || sale || origin, stock);
    });
  }

  function skuHeaders(root) {
    const wrap =
      (root.closest && root.closest(".layui-table-view, .layui-table-box, .el-table")) ||
      root;
    const ths = wrap.querySelectorAll(
      ".layui-table-header th, .el-table__header th, thead th"
    );
    if (ths.length) return Array.from(ths).map(cellValue);
    const cells = wrap.querySelectorAll(".el-table__header .el-table__cell, thead td");
    if (cells.length) return Array.from(cells).map(cellValue);
    const first = wrap.querySelector("tr");
    if (first && /规格名称|销售价|市场价|原价|sku名称|规格信息|^名称$|计算价格/.test(first.innerText || "")) {
      return Array.from(first.querySelectorAll("th, td, .el-table__cell")).map(cellValue);
    }
    return [];
  }

  function skuBodyRows(root) {
    const wrap =
      (root.closest && root.closest(".layui-table-view, .layui-table-box, .el-table")) ||
      root;
    let rows = Array.from(
      wrap.querySelectorAll(".layui-table-body tbody tr, .el-table__body .el-table__row, tbody tr")
    );
    if (!rows.length) {
      rows = Array.from(wrap.querySelectorAll("tr")).filter((tr) => skuCells(tr).length);
    }
    return rows.filter((tr) => {
      if (tr.querySelector("th")) return false;
      const head = (tr.innerText || "").replace(/\s+/g, "");
      return !/规格名称|SKU名称/.test(head.slice(0, 24));
    });
  }

  function skuCells(tr) {
    const tds = Array.from(tr.querySelectorAll("td"));
    if (tds.length) return tds;
    return Array.from(tr.querySelectorAll(":scope > .el-table__cell, :scope > div"));
  }

  function looksLikeSpecName(t) {
    const s = String(t || "").replace(/\s+/g, " ").trim();
    if (!s || /^\d+$/.test(s) || /^\d+(\.\d+)?$/.test(s) || /^[¥￥]/.test(s)) return false;
    if (/规格名称|销售价|市场价|原价|库存|商家编码|查询中|加载中/.test(s)) return false;
    return s.length >= 1 && s.length <= 80;
  }

  function colIndex(headers, keys) {
    for (let k = 0; k < keys.length; k++) {
      const key = keys[k];
      for (let i = 0; i < headers.length; i++) {
        const h = String(headers[i] || "").replace(/\s+/g, "");
        if (!h) continue;
        if (key === "名称") {
          if (h === "名称" || h === "SKU名称" || h === "规格名称") return i;
          continue;
        }
        if (key === "价格" && h.indexOf("市场") >= 0) continue;
        if (h === key || h.indexOf(key) >= 0) return i;
      }
    }
    return -1;
  }

  function cellValue(cell) {
    if (!cell) return "";
    const inputs = Array.from(cell.querySelectorAll("input, textarea"));
    const values = inputs.map((el) => String(el.value || "").trim()).filter(Boolean);
    const text = String(cell.innerText || "").replace(/\s+/g, " ").trim();
    if (!values.length) return text;
    const numeric = values.find((v) => /\d/.test(v));
    if (numeric && (!text || text.indexOf(numeric) < 0)) return numeric;
    return text || values[0];
  }

  function moneyIn(cell) {
    if (!cell) return 0;
    const inputs = Array.from(cell.querySelectorAll("input, textarea"));
    for (let i = 0; i < inputs.length; i++) {
      const n = num(inputs[i].value);
      if (n > 0) return n;
    }
    return num(cellValue(cell));
  }

  function dedupeSkuRows(rows) {
    const map = new Map();
    rows.forEach((row) => {
      const key = (row.specName || "") + "\0" + (row.specValue || row.name || "");
      const prev = map.get(key);
      if (!prev || (row.price > 0 && !(prev.price > 0))) map.set(key, row);
    });
    return Array.from(map.values());
  }

  function pickSrc(el) {
    if (!el) return "";
    return (
      el.getAttribute("data-original") ||
      el.getAttribute("data-src") ||
      el.getAttribute("data-url") ||
      el.currentSrc ||
      el.getAttribute("src") ||
      el.getAttribute("href") ||
      ""
    );
  }

  function clean(u) {
    if (!u) return "";
    let s = String(u).trim();
    if (s.indexOf("//") === 0) s = "https:" + s;
    s = s.split(/[\s"'<>;)\\]/)[0];
    if (!/^https?:/i.test(s)) return "";
    return s.replace(".jpg_.webp", ".jpg").replace(/_\d+x\d+\.(jpg|png|webp)/i, ".$1");
  }

  function num(v) {
    const n = Number(String(v || "").replace(/,/g, "").replace(/[^\d.]/g, ""));
    return isFinite(n) ? n : 0;
  }

  /** SKU 工具库存为「-」或无数字时按 0 */
  function parseStock(v) {
    const text = String(v || "").replace(/\s+/g, "").trim();
    if (!text || /^(?:-|—|–|\*|无|无库存|空)$/.test(text)) return 0;
    const n = Number(text.replace(/,/g, "").replace(/[^\d.]/g, ""));
    if (!isFinite(n) || n < 0) return 0;
    return Math.round(n);
  }

  function bodyText() {
    try {
      return ((document.body && document.body.innerText) || "").replace(/\s+/g, " ").slice(0, 400);
    } catch (e) {
      return "";
    }
  }

  function report(payload) {
    try {
      chrome.runtime.sendMessage({ type: "pc-collect-harvest", payload });
    } catch (e) {}
  }

  function mountPanel() {
    if (document.getElementById("pc-collect-panel")) return;
    const box = document.createElement("div");
    box.id = "pc-collect-panel";
    box.innerHTML =
      '<header>商品采集 <button type="button" data-act="hide">×</button></header>' +
      '<div class="body">' +
      '<div class="hint" id="pc-hint">请点至尊宝「悬浮标题采集」，再点「手机端主图视频SKU」。</div>' +
      '<div class="meta" id="pc-title"></div>' +
      '<div class="counts">' +
      '<div><b id="pc-main">0</b>主图</div>' +
      '<div><b id="pc-detail">0</b>详情</div>' +
      '<div><b id="pc-sku">0</b>SKU</div>' +
      '<div><b id="pc-video">0</b>视频</div>' +
      "</div>" +
      '<label class="field">品牌<select id="pc-brand"><option value="">加载中…</option></select></label>' +
      '<label class="field">分类<select id="pc-category"><option value="">加载中…</option></select></label>' +
      '<div class="actions">' +
      '<button class="upload" id="pc-upload" disabled>上传到商品系统</button>' +
      '<button class="reset" id="pc-reset" type="button">清空</button>' +
      "</div>" +
      '<div class="err" id="pc-err"></div>' +
      "</div>";
    document.documentElement.appendChild(box);
    box.querySelector("[data-act=hide]").onclick = () => box.remove();
    box.querySelector("#pc-reset").onclick = () => location.reload();
    box.querySelector("#pc-upload").onclick = () => doUpload();
    chrome.runtime.onMessage.addListener((msg) => {
      if (msg && msg.type === "pc-collect-state") render(msg.state);
    });
    chrome.runtime.sendMessage({ type: "pc-collect-get" }, (state) => {
      if (state) render(state);
    });
    loadChoices();
  }

  function loadChoices() {
    chrome.runtime.sendMessage({ type: "pc-collect-meta" }, (res) => {
      const brandEl = document.getElementById("pc-brand");
      const categoryEl = document.getElementById("pc-category");
      if (!brandEl || !categoryEl) return;
      if (!res || !res.ok) {
        const text = (res && res.error) || "品牌和分类加载失败";
        brandEl.innerHTML = '<option value="">' + text + "</option>";
        categoryEl.innerHTML = '<option value="">' + text + "</option>";
        return;
      }
      const categories = [];
      flattenCategories(res.categories || [], "", categories);
      chrome.storage.local.get(["brandId", "categoryId"], (saved) => {
        fillSelect(brandEl, res.brands || [], saved && saved.brandId, "无品牌");
        fillSelect(categoryEl, categories, saved && saved.categoryId, "无分类");
      });
    });
  }

  function flattenCategories(list, prefix, out) {
    (list || []).forEach((item) => {
      if (!item || !item.id) return;
      const name = prefix + (item.name || "");
      out.push({ id: item.id, name: name });
      if (item.children && item.children.length) flattenCategories(item.children, name + " / ", out);
    });
  }

  function fillSelect(el, items, savedId, preferName) {
    el.innerHTML = "";
    const blank = document.createElement("option");
    blank.value = "";
    blank.textContent = items.length ? "请选择" : "没有可选数据";
    el.appendChild(blank);
    items.forEach((item) => {
      const opt = document.createElement("option");
      opt.value = String(item.id);
      opt.textContent = item.name || String(item.id);
      el.appendChild(opt);
    });
    const saved = Number(savedId) || 0;
    const named = items.find((item) => item.name === preferName || String(item.name || "").endsWith(" / " + preferName));
    const pick = saved && items.some((item) => Number(item.id) === saved) ? saved : named ? Number(named.id) : 0;
    el.value = pick ? String(pick) : "";
    el.onchange = () => {
      const key = el.id === "pc-brand" ? "brandId" : "categoryId";
      chrome.storage.local.set({ [key]: Number(el.value) || 0 });
    };
    if (pick) {
      const key = el.id === "pc-brand" ? "brandId" : "categoryId";
      chrome.storage.local.set({ [key]: pick });
    }
  }

  function counts(state) {
    const images = state.images || [];
    const main = images.filter((x) => x.kind !== "detail" && x.kind !== "sku").length;
    const detail = images.filter((x) => x.kind === "detail").length;
    return { main, detail, sku: (state.skus || []).length, video: videoSrcs(state.videos).length };
  }

  function render(state) {
    const panel = document.getElementById("pc-collect-panel");
    if (!panel || !state) return;
    const c = counts(state);
    panel.querySelector("#pc-title").textContent = state.title || state.url || "";
    panel.querySelector("#pc-main").textContent = String(c.main);
    panel.querySelector("#pc-detail").textContent = String(c.detail);
    panel.querySelector("#pc-sku").textContent = String(c.sku);
    if (panel.querySelector("#pc-video")) panel.querySelector("#pc-video").textContent = String(c.video);
    let hint = "请点至尊宝「悬浮标题采集」，再点「手机端主图视频SKU」。";
    if (state.clickedTitle && !state.title) hint = "已点标题采集，正在等标题…";
    if (state.title) hint = "标题已抓到。请点至尊宝「手机端主图视频SKU」。";
    if (c.main > 0) hint = "主图已抓到。请再点「SKU工具」抓规格表。";
    if (c.main > 0 && c.video > 0) hint = "主图和视频已抓到。请再点「SKU工具」抓规格表。";
    if (c.main > 0 && c.sku > 0) hint = "可以上传了。先选品牌和分类，再点上传。";
    if (c.main > 0 && c.sku > 0 && c.video > 0) hint = "可以上传了。视频会按 1:1 / 3:4 / 16:9 / 9:16 入库。";
    if (state.clickedMedia && c.main === 0) hint = "已点主图工具，正在等弹层加载…";
    if (state.clickedSku && c.sku === 0) hint = "已点 SKU 工具，正在等规格表加载…";
    panel.querySelector("#pc-hint").textContent = hint;
    panel.querySelector("#pc-hint").className = "hint " + (c.main > 0 ? "ok" : "wait");
    panel.querySelector("#pc-upload").disabled = !(state.title && c.main > 0);
    panel.dataset.state = JSON.stringify(state);
  }

  function doUpload() {
    const panel = document.getElementById("pc-collect-panel");
    const err = panel.querySelector("#pc-err");
    const brandId = Number((panel.querySelector("#pc-brand") || {}).value) || 0;
    const categoryId = Number((panel.querySelector("#pc-category") || {}).value) || 0;
    if (!brandId || !categoryId) {
      err.className = "err";
      err.textContent = "请先选择品牌和分类。";
      return;
    }
    err.className = "err";
    err.textContent = "上传中…";
    const state = JSON.parse(panel.dataset.state || "{}");
    const product = toProduct(state, brandId, categoryId);
    chrome.storage.local.set({ brandId: brandId, categoryId: categoryId });
    chrome.runtime.sendMessage(
      {
        type: "pc-collect-upload",
        payload: {
          productUrl: state.url,
          platform: state.platform || "taobao",
          product,
          videos: videoSrcs(state.videos),
        },
      },
      (res) => {
        if (!res || !res.ok) {
          err.textContent = (res && res.error) || "上传失败，请在扩展图标里填写 API 和 Token。";
          return;
        }
        const id = res.task && res.task.productId;
        err.className = "ok";
        err.textContent = id ? "已写入草稿商品 #" + id : res.task && res.task.message ? res.task.message : "已上传";
      }
    );
  }

  function toProduct(state, brandId, categoryId) {
    const itemId = state.itemId || ((state.url || "").match(/[?&]id=(\d+)/) || [])[1] || "";
    const images = state.images || [];
    const main = [];
    const detail = [];
    const skuPics = [];
    images.forEach((img) => {
      if (!img || !img.src) return;
      if (img.kind === "detail") detail.push(img.src);
      else if (img.kind === "sku") skuPics.push(img.src);
      else main.push(img.src);
    });
    if (!main.length && skuPics.length) main.push(skuPics[0]);
    const built = buildImportedSkus(state.skus || [], skuPics);
    const skus = built.skus;
    const priced = skus.filter((s) => s.price > 0);
    const price = priced.length ? Math.min.apply(null, priced.map((s) => s.price)) : 0;
    const original = skus.reduce((m, s) => Math.max(m, s.marketPrice || s.price || 0), 0);
    const stock = skus.reduce((m, s) => m + (s.stock || 0), 0);
    return {
      name: state.title || "未命名商品",
      subTitle: "",
      materialCode: "",
      source: "淘宝",
      productSn: itemId,
      brandId: brandId,
      categoryId: categoryId,
      pic: main[0] || "",
      albumPics: unique(main),
      media: {
        detailPics: unique(detail),
      },
      price,
      originalPrice: original,
      stock,
      unit: "件",
      publishStatus: 0,
      isDraft: 1,
      verifyStatus: 1,
      description: state.title || "",
      detailHtml: unique(detail)
        .map((u) => "<p><img src=\"" + u.replace(/"/g, "&quot;") + "\" /></p>")
        .join(""),
      skuSpecs: built.skuSpecs,
      skus,
      channelVisible: "both",
    };
  }

  /** 与 web/src/utils/skuCsvImport.ts 一致：规格值取冒号后，规格名固定「商品规格」 */
  var DEFAULT_SPEC_NAME = "商品规格";

  function parseSkuName(skuName) {
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
    return { specName: DEFAULT_SPEC_NAME, specValue: specValue || text };
  }

  function nextSkuCode(used) {
    const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789";
    for (let attempt = 0; attempt < 10000; attempt++) {
      let code = "";
      for (let i = 0; i < 8; i++) code += chars[Math.floor(Math.random() * chars.length)];
      if (!used[code]) {
        used[code] = 1;
        return code;
      }
    }
    return "SKU" + String(Object.keys(used).length + 1);
  }

  function buildImportedSkus(rows, skuPics) {
    const used = {};
    const byKey = new Map();
    rows.forEach((row, i) => {
      const parsed = parseSkuName(row.name || "");
      const specName = DEFAULT_SPEC_NAME;
      const specValue = String(row.specValue || parsed.specValue || "").trim();
      if (!specValue) return;
      const pic = row.pic || skuPics[i] || "";
      const item = {
        specName,
        specValue,
        pic,
        price: row.price || 0,
        marketPrice: row.originalPrice || row.price || 0,
        stock: row.stock || 0,
      };
      const key = item.specName + "\0" + item.specValue;
      const prev = byKey.get(key);
      if (!prev || (item.price > 0 && !(prev.price > 0))) byKey.set(key, item);
    });
    const items = Array.from(byKey.values());
    const specOrder = [];
    const specValues = {};
    items.forEach((item) => {
      if (!specValues[item.specName]) {
        specValues[item.specName] = [];
        specOrder.push(item.specName);
      }
      const bucket = specValues[item.specName];
      const found = bucket.find((v) => v.value === item.specValue);
      if (!found) bucket.push({ value: item.specValue, pic: item.pic || "" });
      else if (!found.pic && item.pic) found.pic = item.pic;
    });
    return {
      skus: items.map((item) => ({
        skuCode: nextSkuCode(used),
        specs: { [item.specName]: item.specValue },
        price: item.price,
        marketPrice: item.marketPrice,
        stock: item.stock,
        pic: item.pic,
      })),
      skuSpecs: specOrder.map((name) => ({ name, values: specValues[name] })),
    };
  }

  function unique(list) {
    const seen = new Set();
    const out = [];
    list.forEach((s) => {
      if (!s || seen.has(s)) return;
      seen.add(s);
      out.push(s);
    });
    return out;
  }

  function videoSrcs(list) {
    const out = [];
    (list || []).forEach((item) => {
      const src = typeof item === "string" ? item : item && item.src;
      if (src && isVideoURL(src) && out.indexOf(src) < 0) out.push(src);
    });
    return out;
  }
})();
