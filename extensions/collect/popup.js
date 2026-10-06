const DEFAULT_API = "https://osms.zfcycle.com/apps/product";
const apiBase = document.getElementById("apiBase");
const token = document.getElementById("token");
const msg = document.getElementById("msg");

chrome.storage.local.get(["apiBase", "token"]).then((cfg) => {
  apiBase.value = cfg.apiBase || DEFAULT_API;
  token.value = cfg.token || "";
  refreshLoginHint();
});

document.getElementById("save").addEventListener("click", async () => {
  let base = (apiBase.value || "").trim().replace(/\/+$/, "") || DEFAULT_API;
  if (/^https?:\/\/osms\.zfcycle\.com$/i.test(base)) {
    base = DEFAULT_API;
  }
  await chrome.storage.local.set({
    apiBase: base,
    token: token.value.trim(),
  });
  apiBase.value = base;
  msg.className = "ok";
  msg.textContent = "已保存。上传地址：" + base + "/api/v1/admin/product-collects/ingest";
});

document.getElementById("login").addEventListener("click", async () => {
  let base = (apiBase.value || "").trim().replace(/\/+$/, "") || DEFAULT_API;
  chrome.tabs.create({ url: base + "/", active: true });
});

document.getElementById("check").addEventListener("click", () => {
  refreshLoginHint();
});

function refreshLoginHint() {
  const base = (apiBase.value || DEFAULT_API).trim() || DEFAULT_API;
  const urls = [base, "https://osms.zfcycle.com/", "https://osms.zfcycle.com/apps/product/"];
  const tryNext = (i) => {
    if (i >= urls.length) {
      msg.className = "wait";
      msg.textContent = "本 Chrome 尚未登录商品系统。点「在本浏览器打开并登录」，登完再点检测。";
      return;
    }
    chrome.cookies.get({ url: urls[i], name: "uc_access" }, (ck) => {
      if (ck && ck.value) {
        msg.className = "ok";
        msg.textContent = "已检测到登录 Cookie，上传时不用填 Token。";
        return;
      }
      tryNext(i + 1);
    });
  };
  tryNext(0);
}
