(() => {
  const SKIP = /favicon|sprite|\/icon|logo_|\.svg(\?|$)|zzbtool\.com|\.js(\?|$)|\.css(\?|$)/i;
  const isTop = window === window.top;
  const href = location.href || "";

  let hoveredTitle = "";

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
    },
    true
  );

  document.addEventListener(
    "mouseover",
    (ev) => {
      const t = pickHoveredTitle(ev.target);
      if (!t || t === hoveredTitle) return;
      hoveredTitle = t;
      report({ title: t });
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
    if (payload.title || payload.images || payload.skus) {
      report(payload);
    }
  }

  function looksLikeTitle(raw) {
    const t = String(raw || "")
      .replace(/\s+/g, " ")
      .trim();
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

  function pickHoveredTitle(el) {
    if (!el || (el.closest && el.closest("#pc-collect-panel"))) return "";
    const node =
      (el.closest &&
        el.closest(
          "h1, #J_Title, [class*='ItemTitle'], [class*='mainTitle'], [class*='itemTitle']"
        )) ||
      null;
    return looksLikeTitle(node ? node.innerText || node.textContent : "");
  }

  function extractTitle() {
    const out = {};
    const id = (href.match(/[?&]id=(\d+)/) || [])[1] || "";
    if (id) out.itemId = id;
    const cands = [];
    if (hoveredTitle) cands.push(hoveredTitle);
    document
      .querySelectorAll(
        "h1, #J_Title h3, #J_Title, [class*='ItemTitle'], [class*='mainTitle'], [class*='itemTitle']"
      )
      .forEach((el) => {
        const t = looksLikeTitle(el.innerText || el.textContent);
        if (t) cands.push(t);
      });
    document.querySelectorAll("input, textarea").forEach((el) => {
      const t = looksLikeTitle(el.value);
      if (t) cands.push(t);
    });
    const page = looksLikeTitle(
      (document.title || "")
        .replace(/-tmall\.com.*$/i, "")
        .replace(/-淘宝网.*$/, "")
        .replace(/-天猫.*$/, "")
    );
    if (page) cands.push(page);
    cands.sort((a, b) => b.length - a.length);
    if (cands[0]) out.title = cands[0].slice(0, 200);
    return out;
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
      const u = clean(src);
      if (!u || videos.indexOf(u) >= 0) return;
      if (/\.(mp4|m3u8)(\?|$)/i.test(u) || /cloudvideo/i.test(u)) videos.push(u);
    };
    if (!looks) return { images, videos };

    const scanRow = (el) => {
      const line = ((el.innerText || "") + " " + (el.getAttribute("class") || "")).replace(/\s+/g, " ");
      if (/教程|导出文件|批量下载/.test(line) && !el.querySelector("img")) return;
      const img = el.querySelector("img");
      const link = el.querySelector("a[href]");
      const src = pickSrc(img) || pickSrc(link) || firstAlicdn(el.innerHTML || "");
      if (!src) return;
      addImg(src, kindFromLine(line));
      addVid(pickSrc(link));
    };

    document.querySelectorAll("table tr, .el-table__row, .layui-table tr, [class*='oneDownload'] li").forEach(scanRow);

    if (images.length === 0) {
      document.querySelectorAll("img, a[href]").forEach((el) => {
        const row = el.closest("tr, .el-table__row, li, [class*='row']") || el.parentElement;
        const line = row ? (row.innerText || "") : "";
        addImg(pickSrc(el), kindFromLine(line));
        addVid(pickSrc(el));
      });
    }
    return { images, videos };
  }

  function kindFromLine(line) {
    const t = String(line || "");
    if (/详情页|详情图|详情0/.test(t) || (/详情/.test(t) && !/主图/.test(t))) return "detail";
    if (/SKU\s*\(|SKU0|SKU图|来源\s*SKU/.test(t) || (/SKU/.test(t) && !/主图|手机端/.test(t))) return "sku";
    if (/主图/.test(t)) return "main";
    return "main";
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
    const text = bodyText();
    const looks = /规格名称|销售价|市场价|商家编码|规格信息/.test(text) && !/资源一键下载|全选 \(/.test(text);
    if (!looks && !/sku/i.test(href)) return [];
    const rows = [];
    const push = (name, pic, price, original, stock, skuId) => {
      name = String(name || "").replace(/\s+/g, " ").trim();
      if (!name && !pic) return;
      if (/规格名称|销售价|市场价|库存|商家编码|查询中|加载中/.test(name) && !pic) return;
      rows.push({
        name,
        pic: clean(pic),
        price: num(price),
        originalPrice: num(original),
        stock: Math.round(num(stock)),
        skuId: String(skuId || "").trim(),
      });
    };
    document.querySelectorAll("table").forEach((table) => {
      table.querySelectorAll("tr").forEach((tr) => {
        const tds = Array.from(tr.querySelectorAll("td"));
        if (tds.length < 2) return;
        const img = tr.querySelector("img");
        const texts = tds.map((td) => (td.innerText || "").trim()).filter(Boolean);
        if (!texts.length) return;
        const head = texts.join(" ");
        if (/规格|价格|库存|SKU|商家编码/.test(head) && texts.length <= 6 && !img && !/\d+\.\d{2}/.test(head)) return;
        const price = texts.find((x) => /¥|￥|\d+\.\d{2}/.test(x));
        const stock = texts.find((x) => /库存|件/.test(x)) || texts.find((x, i) => i > 0 && /^\d+$/.test(x));
        const name = texts.find((x) => !/¥|￥|库存|件/.test(x) && !/^\d+(\.\d+)?$/.test(x)) || texts[0];
        push(name, img ? img.src : "", price, "", stock, "");
      });
    });
    if (!rows.length) {
      document.querySelectorAll(".el-table__row, .tbb_sku_item").forEach((el) => {
        const img = el.querySelector("img");
        const line = (el.innerText || "").replace(/\s+/g, " ").trim();
        if (!line) return;
        const parts = line.split(" ").filter(Boolean);
        push(parts[0], img ? img.src : "", parts.find((x) => /¥|￥|\d+\.\d{2}/.test(x)), "", parts.find((x) => /库存|件/.test(x)), "");
      });
    }
    return rows;
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
      '<div class="hint" id="pc-hint">请把鼠标放到商品标题上踩一下，再点至尊宝「手机端主图视频SKU」。</div>' +
      '<div class="meta" id="pc-title"></div>' +
      '<div class="counts">' +
      '<div><b id="pc-main">0</b>主图</div>' +
      '<div><b id="pc-detail">0</b>详情</div>' +
      '<div><b id="pc-sku">0</b>SKU</div>' +
      "</div>" +
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
  }

  function counts(state) {
    const images = state.images || [];
    const main = images.filter((x) => x.kind !== "detail" && x.kind !== "sku").length;
    const detail = images.filter((x) => x.kind === "detail").length;
    return { main, detail, sku: (state.skus || []).length, video: (state.videos || []).length };
  }

  function render(state) {
    const panel = document.getElementById("pc-collect-panel");
    if (!panel || !state) return;
    const c = counts(state);
    panel.querySelector("#pc-title").textContent = state.title || state.url || "";
    panel.querySelector("#pc-main").textContent = String(c.main);
    panel.querySelector("#pc-detail").textContent = String(c.detail);
    panel.querySelector("#pc-sku").textContent = String(c.sku);
    let hint = "请把鼠标放到商品标题上踩一下，再点至尊宝「手机端主图视频SKU」。";
    if (state.title) hint = "标题已抓到。请点至尊宝「手机端主图视频SKU」。";
    if (c.main > 0) hint = "主图已抓到。请再点「SKU工具」抓规格表。";
    if (c.main > 0 && c.sku > 0) hint = "可以上传了。品牌/分类入库后为 0，在商品系统里补。";
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
    err.textContent = "上传中…";
    const state = JSON.parse(panel.dataset.state || "{}");
    const product = toProduct(state);
    chrome.runtime.sendMessage(
      {
        type: "pc-collect-upload",
        payload: {
          productUrl: state.url,
          platform: state.platform || "taobao",
          product,
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

  function toProduct(state) {
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
    const skus = (state.skus || []).map((row, i) => {
      const name = row.name || "规格" + (i + 1);
      const pic = row.pic || skuPics[i] || "";
      const skuId = String(row.skuId || "").replace(/[^A-Za-z0-9]/g, "");
      const code = skuId ? "TB" + skuId : itemId ? "TB" + itemId + "S" + String(i + 1).padStart(2, "0") : "SKU" + String(i + 1).padStart(2, "0");
      return {
        skuCode: code.slice(0, 64),
        specs: { 规格: name },
        price: row.price || 0,
        marketPrice: row.originalPrice || row.price || 0,
        stock: row.stock || 0,
        pic,
      };
    });
    const specValues = skus.map((s) => ({ value: s.specs["规格"], pic: s.pic || "" }));
    const video = (state.videos || [])[0] || "";
    const priced = skus.filter((s) => s.price > 0);
    const price = priced.length ? Math.min.apply(null, priced.map((s) => s.price)) : 0;
    const original = skus.reduce((m, s) => Math.max(m, s.marketPrice || s.price || 0), 0);
    const stock = skus.reduce((m, s) => m + (s.stock || 0), 0);
    return {
      name: state.title || "未命名商品",
      subTitle: "",
      materialCode: itemId ? "TB" + itemId : "",
      source: state.platform || "taobao",
      productSn: itemId,
      brandId: 0,
      categoryId: 0,
      pic: main[0] || "",
      albumPics: unique(main),
      productVideo: video,
      media: {
        videos: video ? { ratio11: video } : undefined,
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
      skuSpecs: specValues.length ? [{ name: "规格", values: specValues }] : [],
      skus,
      channelVisible: "both",
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
})();
