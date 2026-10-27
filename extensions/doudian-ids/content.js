(() => {
  if (window.__pcDoudianIdsBooted) return;
  window.__pcDoudianIdsBooted = true;

  const ID_RE = /^\d{6,20}$/;
  const ID_KEYS = ["product_id", "productId", "ProductId", "productID"];
  const LIST_KEYS = /^(data|list|products|records|items|result|rows)$/i;
  const PAGE_KEYS = ["page", "pageNo", "page_no", "pageNum", "page_num", "pageIndex"];
  const SIZE_KEYS = ["pageSize", "page_size", "size", "limit", "pageCount"];
  const captured = { req: null };

  function isListPage() {
    return /\/ffa\/g\/list/.test(location.pathname || "");
  }

  function uniqIds(list) {
    const out = [];
    const seen = new Set();
    (list || []).forEach((id) => {
      const s = String(id || "").trim();
      if (!ID_RE.test(s) || seen.has(s)) return;
      seen.add(s);
      out.push(s);
    });
    return out;
  }

  function pickId(obj) {
    if (!obj || typeof obj !== "object" || Array.isArray(obj)) return "";
    for (const key of ID_KEYS) {
      const v = obj[key];
      if (v != null && ID_RE.test(String(v))) return String(v);
    }
    return "";
  }

  function extractIds(node, out) {
    if (!node || typeof node !== "object") return out;
    if (Array.isArray(node)) {
      node.forEach((item) => {
        const id = pickId(item);
        if (id) out.push(id);
        else extractIds(item, out);
      });
      return out;
    }
    const self = pickId(node);
    if (self) out.push(self);
    Object.keys(node).forEach((key) => {
      if (LIST_KEYS.test(key) || key === "data") extractIds(node[key], out);
    });
    return out;
  }

  function findTotal(node) {
    if (!node || typeof node !== "object") return 0;
    const direct = node.total ?? node.totalCount ?? node.total_count ?? node.count;
    if (Number(direct) > 0) return Number(direct);
    if (node.data && typeof node.data === "object") {
      const inner = node.data.total ?? node.data.totalCount ?? node.data.total_count;
      if (Number(inner) > 0) return Number(inner);
    }
    return 0;
  }

  function parseBody(raw) {
    if (!raw) return null;
    if (typeof raw === "string") {
      try {
        return JSON.parse(raw);
      } catch {
        try {
          return Object.fromEntries(new URLSearchParams(raw));
        } catch {
          return null;
        }
      }
    }
    if (typeof raw === "object" && !(raw instanceof FormData) && !(raw instanceof Blob)) {
      return raw;
    }
    return null;
  }

  function headersObject(init) {
    const out = {};
    const src = init && init.headers;
    if (!src) return out;
    if (typeof src.forEach === "function") {
      src.forEach((v, k) => {
        out[k] = v;
      });
      return out;
    }
    Object.keys(src).forEach((k) => {
      out[k] = src[k];
    });
    return out;
  }

  function pageInfo(url, body) {
    const query = {};
    try {
      const u = new URL(url, location.origin);
      u.searchParams.forEach((v, k) => {
        query[k] = v;
      });
    } catch {
    }
    const src = Object.assign({}, query, body && typeof body === "object" ? body : {});
    const pageKey = PAGE_KEYS.find((k) => k in src) || "page";
    const sizeKey = SIZE_KEYS.find((k) => k in src) || "pageSize";
    const pageRaw = src[pageKey];
    const page = pageRaw === undefined || pageRaw === "" ? 1 : Number(pageRaw);
    return {
      pageKey,
      sizeKey,
      page: Number.isFinite(page) ? page : 1,
      pageSize: Number(src[sizeKey]) || 20,
      zeroBased: Number(pageRaw) === 0,
    };
  }

  function rememberList(url, init, json) {
    const href = String(url || "");
    const ids = uniqIds(extractIds(json, []));
    if (ids.length < 1) return;
    const listApi = /\/product\/tproduct\/list/i.test(href);
    if (!listApi && ids.length < 8) return;
    if (/zijieapi|helpdesk|govern|mcs\./i.test(href)) return;
    const body = parseBody(init && init.body);
    const next = {
      url: href,
      method: String((init && init.method) || "GET").toUpperCase(),
      headers: headersObject(init),
      body,
      paging: pageInfo(href, body),
      total: findTotal(json),
    };
    if (listApi || !captured.req || /\/product\/tproduct\/list/i.test(captured.req.url) === false) {
      captured.req = next;
    }
  }

  const rawFetch = window.fetch.bind(window);
  window.fetch = async function (input, init) {
    const res = await rawFetch(input, init);
    try {
      const url = typeof input === "string" ? input : input && input.url;
      const clone = res.clone();
      const ct = clone.headers.get("content-type") || "";
      if (ct.includes("json")) rememberList(url, init || {}, await clone.json());
    } catch {
    }
    return res;
  };

  const RawXHR = window.XMLHttpRequest;
  function HookedXHR() {
    const xhr = new RawXHR();
    let url = "";
    let method = "GET";
    let body = null;
    xhr.addEventListener("load", () => {
      try {
        const ct = xhr.getResponseHeader("content-type") || "";
        if (!ct.includes("json")) return;
        rememberList(url, { method, body, headers: {} }, JSON.parse(xhr.responseText));
      } catch {
      }
    });
    const rawOpen = xhr.open;
    xhr.open = function (m, u) {
      method = m;
      url = u;
      return rawOpen.apply(xhr, arguments);
    };
    const rawSend = xhr.send;
    xhr.send = function (payload) {
      body = payload;
      return rawSend.apply(xhr, arguments);
    };
    return xhr;
  }
  HookedXHR.prototype = RawXHR.prototype;
  window.XMLHttpRequest = HookedXHR;

  function sleep(ms) {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }

  function collectDomIds() {
    const ids = [];
    document.querySelectorAll("a[href], [data-product-id], [data-productid]").forEach((el) => {
      const href = el.getAttribute("href") || "";
      const m = href.match(/product_id=(\d{6,20})/i) || href.match(/productId=(\d{6,20})/i);
      if (m) ids.push(m[1]);
      const data = el.getAttribute("data-product-id") || el.getAttribute("data-productid");
      if (data && ID_RE.test(data)) ids.push(data);
    });
    const text = document.body ? document.body.innerText : "";
    const re = /(?:商品)?ID[:：]\s*(\d{6,20})/g;
    let hit;
    while ((hit = re.exec(text))) ids.push(hit[1]);
    Array.from(document.querySelectorAll("tr.ecom-g-table-row")).forEach((row) => {
      const m = (row.innerText || "").match(/ID[:：]\s*(\d{6,20})/);
      if (m) ids.push(m[1]);
    });
    return uniqIds(ids);
  }

  function disabled(el) {
    if (!el) return true;
    const cls = String(el.className || "");
    return el.disabled
      || el.getAttribute("aria-disabled") === "true"
      || cls.includes("disabled")
      || cls.includes("Disabled");
  }

  function clickable(el) {
    if (!el) return null;
    return el.closest("button, a, li, [role='button']") || el;
  }

  function findPageBtn(label) {
    if (label === "1") {
      const item = document.querySelector(".ecom-g-pagination-item-1, .ant-pagination-item-1");
      if (item && !disabled(item)) return clickable(item);
    }
    const nodes = Array.from(document.querySelectorAll("button, a, li, span"));
    const hit = nodes.find((el) => (el.innerText || "").trim() === label && !disabled(el));
    return clickable(hit);
  }

  function findNextBtn() {
    const next = document.querySelector(
      ".ecom-g-pagination-next:not(.ecom-g-pagination-disabled), .ant-pagination-next:not(.ant-pagination-disabled)"
    );
    if (next && !disabled(next)) {
      return next.querySelector("button, a") || next;
    }
    return findPageBtn("下一页");
  }

  function snapshotRows() {
    return collectDomIds().join(",") + "|" + ((document.body && document.body.innerText) || "").slice(0, 800);
  }

  async function waitRowsChange(before, timeout) {
    const start = Date.now();
    while (Date.now() - start < timeout) {
      await sleep(250);
      if (snapshotRows() !== before) return true;
    }
    return false;
  }

  async function exportByApi(onProgress) {
    const req = captured.req;
    if (!req) return null;
    const ids = [];
    const seen = new Set();
    const size = Math.max(req.paging.pageSize || 20, 20);
    const zeroBased = req.paging.zeroBased || req.paging.page === 0;
    let total = req.total || 0;
    for (let i = 0; i < 500; i += 1) {
      const page = zeroBased ? i : i + 1;
      const payload = Object.assign({}, req.body || {});
      payload[req.paging.pageKey] = page;
      payload[req.paging.sizeKey] = size;
      let url = req.url;
      const headers = Object.assign({ accept: "application/json" }, req.headers || {});
      const init = { method: req.method, credentials: "include", headers };
      if (req.method === "GET") {
        const u = new URL(url, location.origin);
        u.searchParams.set(req.paging.pageKey, String(page));
        u.searchParams.set(req.paging.sizeKey, String(size));
        url = u.toString();
      } else {
        if (!headers["Content-Type"] && !headers["content-type"]) {
          headers["content-type"] = "application/json";
        }
        init.body = JSON.stringify(payload);
      }
      const res = await rawFetch(url, init);
      const json = await res.json();
      const batch = uniqIds(extractIds(json, []));
      batch.forEach((id) => {
        if (seen.has(id)) return;
        seen.add(id);
        ids.push(id);
      });
      total = findTotal(json) || total;
      onProgress({ page, count: ids.length, total, mode: "接口" });
      if (batch.length === 0) break;
      if (total && ids.length >= total) break;
      if (batch.length < size && (!total || ids.length >= total)) break;
      await sleep(280);
    }
    return ids;
  }

  async function exportByDom(onProgress) {
    const first = findPageBtn("1");
    if (first) {
      const before = snapshotRows();
      first.click();
      await waitRowsChange(before, 8000);
    }
    const ids = [];
    const seen = new Set();
    for (let page = 1; page <= 500; page += 1) {
      await sleep(200);
      collectDomIds().forEach((id) => {
        if (seen.has(id)) return;
        seen.add(id);
        ids.push(id);
      });
      onProgress({ page, count: ids.length, total: 0, mode: "翻页" });
      const next = findNextBtn();
      if (!next) break;
      const before = snapshotRows();
      next.click();
      const changed = await waitRowsChange(before, 10000);
      if (!changed) break;
    }
    return ids;
  }

  async function tryTriggerList() {
    const prev = captured.req;
    const query = Array.from(document.querySelectorAll("button")).find((el) => (el.innerText || "").trim() === "查询")
      || findPageBtn("查询")
      || findPageBtn("搜索");
    if (query) {
      captured.req = null;
      query.click();
      const start = Date.now();
      while (!captured.req && Date.now() - start < 5000) {
        await sleep(200);
      }
    }
    if (!captured.req) captured.req = prev;
  }

  function downloadText(text) {
    const blob = new Blob([text], { type: "text/plain;charset=utf-8" });
    const a = document.createElement("a");
    a.href = URL.createObjectURL(blob);
    a.download = "doudian-product-ids.txt";
    a.click();
    setTimeout(() => URL.revokeObjectURL(a.href), 2000);
  }

  function mountPanel() {
    if (!isListPage() || document.getElementById("pc-doudian-ids")) return;
    const root = document.createElement("div");
    root.id = "pc-doudian-ids";
    root.innerHTML = `
      <header>
        <span>抖店商品ID导出</span>
        <button type="button" data-act="min" title="收起">–</button>
      </header>
      <div class="body">
        <div class="hint">按当前列表筛选导出全部商品ID，自动翻页，结果用逗号拼接。</div>
        <div class="meta" data-role="status">打开列表后点导出。接口优先，失败再点页面翻页。</div>
        <div class="row">
          <button type="button" class="act" data-act="export">导出全部商品ID</button>
        </div>
        <div class="row" style="margin-top:8px">
          <button type="button" class="act secondary" data-act="copy">复制</button>
          <button type="button" class="act secondary" data-act="download">下载 txt</button>
        </div>
        <textarea data-role="out" readonly placeholder="导出结果会显示在这里"></textarea>
      </div>
    `;
    document.documentElement.appendChild(root);
    const status = root.querySelector("[data-role=status]");
    const out = root.querySelector("[data-role=out]");
    const exportBtn = root.querySelector("[data-act=export]");

    function setStatus(text, cls) {
      status.className = "meta" + (cls ? " " + cls : "");
      status.textContent = text;
    }

    root.querySelector("[data-act=min]").addEventListener("click", () => {
      const body = root.querySelector(".body");
      body.style.display = body.style.display === "none" ? "" : "none";
    });

    root.querySelector("[data-act=copy]").addEventListener("click", async () => {
      const text = out.value.trim();
      if (!text) {
        setStatus("还没有可复制的结果", "wait");
        return;
      }
      try {
        await navigator.clipboard.writeText(text);
        setStatus("已复制 " + text.split(",").filter(Boolean).length + " 个ID", "ok");
      } catch {
        out.select();
        document.execCommand("copy");
        setStatus("已复制", "ok");
      }
    });

    root.querySelector("[data-act=download]").addEventListener("click", () => {
      const text = out.value.trim();
      if (!text) {
        setStatus("还没有可下载的结果", "wait");
        return;
      }
      downloadText(text);
      setStatus("已下载 txt", "ok");
    });

    exportBtn.addEventListener("click", async () => {
      exportBtn.disabled = true;
      out.value = "";
      setStatus("正在收集商品ID…", "wait");
      try {
        await tryTriggerList();
        const ids = captured.req
          ? await exportByApi((p) => {
            setStatus(
              (p.mode || "接口") + " 第 " + p.page + " 页，已收集 " + p.count + (p.total ? " / " + p.total : "") + " 个",
              "wait"
            );
          })
          : null;
        const finalIds = ids && ids.length
          ? ids
          : await exportByDom((p) => {
            setStatus("翻页 第 " + p.page + " 页，已收集 " + p.count + " 个", "wait");
          });
        out.value = finalIds.join(",");
        setStatus("完成，共 " + finalIds.length + " 个商品ID", finalIds.length ? "ok" : "wait");
      } catch (err) {
        setStatus("导出失败：" + (err && err.message ? err.message : String(err)), "err");
      } finally {
        exportBtn.disabled = false;
      }
    });
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", mountPanel, { once: true });
  } else {
    mountPanel();
  }
  const tick = setInterval(() => {
    if (isListPage()) mountPanel();
  }, 1500);
  window.addEventListener("beforeunload", () => clearInterval(tick));
})();
