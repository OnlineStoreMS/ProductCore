/**
 * 抖店发布 Step1 — 单次 chrome_javascript 可执行体
 * 在 MCP 工具 chrome_javascript 的 code 参数中粘贴本文件全文（或 readFile 后传入）
 *
 * 环境变量（由 run-step1.mjs 注入到 code 字符串）：
 *   TITLE, IMAGE_URL, CREATE_URL, CLICK_NEXT
 */
async function douyinStep1({
  title,
  imageUrl,
  createUrl = 'https://fxg.jinritemai.com/ffa/g/create',
  clickNext = true,
  waitMs = 5000,
}) {
  if (location.href.indexOf('/ffa/g/create') === -1) {
    location.href = createUrl;
    await new Promise((r) => setTimeout(r, 2500));
  }

  const titleInput = document.querySelector('#pg-title-input');
  if (!titleInput) throw new Error('未找到标题输入框 #pg-title-input，请确认在抖店商品发布页');

  titleInput.focus();
  titleInput.value = title;
  titleInput.dispatchEvent(new Event('input', { bubbles: true }));
  titleInput.dispatchEvent(new Event('change', { bubbles: true }));

  const fileInput = document.querySelectorAll(
    'label.index-module_button__st1_R input[type=file][accept="image/*"]'
  )[0];
  if (!fileInput) throw new Error('未找到主图 file input');

  const resp = await fetch(imageUrl);
  if (!resp.ok) throw new Error(`拉取主图失败: ${resp.status} ${imageUrl}`);
  const blob = await resp.blob();
  const file = new File([blob], 'main.jpg', { type: blob.type || 'image/jpeg' });
  const dt = new DataTransfer();
  dt.items.add(file);
  fileInput.files = dt.files;
  fileInput.dispatchEvent(new Event('input', { bubbles: true }));
  fileInput.dispatchEvent(new Event('change', { bubbles: true }));

  await new Promise((r) => setTimeout(r, waitMs));

  const nextBtn = [...document.querySelectorAll('button')].find((b) =>
    b.innerText?.includes('下一步')
  );
  const result = {
    url: location.href,
    title: titleInput.value,
    titleLen: titleInput.value.length,
    hasMainPreview: !!document.querySelector('label.index-module_button__st1_R img'),
    nextDisabled: nextBtn?.disabled ?? true,
    categoryHint: document.body.innerText.match(/居家日用|自行车|配件|运动/)?.[0],
    toast: document.body.innerText.match(/上传失败|请上传|已预填|成功/g) || [],
  };

  if (clickNext && nextBtn && !nextBtn.disabled) {
    nextBtn.click();
    await new Promise((r) => setTimeout(r, 3000));
    result.afterClick = {
      url: location.href,
      tabs: [...document.querySelectorAll('[role=tab]')].slice(0, 8).map((t) => t.innerText?.trim()),
    };
  }

  return result;
}

// 当作为 chrome_javascript 注入时，run-step1.mjs 会替换下方占位符
return await douyinStep1({
  title: '__TITLE__',
  imageUrl: '__IMAGE_URL__',
  createUrl: '__CREATE_URL__',
  clickNext: __CLICK_NEXT__,
  waitMs: __WAIT_MS__,
});
