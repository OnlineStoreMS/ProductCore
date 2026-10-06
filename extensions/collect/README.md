# ProductCore 采集扩展

人工点至尊宝的 **手机端主图视频SKU**、**SKU工具**，扩展只负责抓弹层、汇总，再点面板上的 **上传到商品系统**。

不自动点按钮，避免淘宝把操作当成机器执行。

## 安装（Chrome 116）

1. 打开 `chrome://extensions`
2. 打开「开发者模式」
3. 「加载已解压的扩展程序」
4. 选本目录：`ProductCore/extensions/collect`
5. 点工具栏图标：
   - API 填 `https://osms.zfcycle.com/apps/product`
   - 点 **在本浏览器打开并登录**，用采集这套 Chrome 登录一次
   - 点 **检测本机登录**，看到「已检测到登录 Cookie」后，Token 留空即可

采集必须用 **Chrome 116 这一套** 登录。日常 Chrome 里已经登录过，这里仍要再登一次，Cookie 不会自动带过来。

## 使用

1. 用已登录淘宝、已装至尊宝的 Chrome 打开商品页
2. 右下角出现「商品采集」面板
3. 人手点至尊宝 **手机端主图视频SKU**，等弹层出图
4. 再点 **SKU工具**，等规格表
5. 面板数字齐了之后点 **上传到商品系统**
6. 草稿出现在 [商品系统草稿箱](https://osms.zfcycle.com/apps/product/)（品牌/分类为 0，需在后台补）

## 接口

`POST https://osms.zfcycle.com/apps/product/api/v1/admin/product-collects/ingest`

Body 与 WindowsAgent 回写同一套 `ProductDTO`。
