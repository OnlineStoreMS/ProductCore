# 抖店商品ID导出

在抖店 **商品管理 → 商品列表** 自动翻页，导出当前筛选条件下的全部商品 ID，用逗号拼接。

只读当前已登录店铺的列表，不改商品。

## 安装

1. 打开 `chrome://extensions`
2. 打开「开发者模式」
3. 「加载已解压的扩展程序」
4. 选本目录：`ProductCore/extensions/doudian-ids`

## 使用

1. 用已登录抖店后台的浏览器打开 [商品列表](https://fxg.jinritemai.com/ffa/g/list?sov_goodsType=0&tab=all)
2. 需要哪个范围，先在页面上选好 Tab / 筛选（全部、在售、仓库等）
3. 右下角点 **导出全部商品ID**
4. 完成后点 **复制** 或 **下载 txt**

优先复用页面自己的列表接口 `/product/tproduct/list` 按页拉取（抖店首页是 `page=0`）；接口没抓到时，会回到第一页再点「下一页」扫完。
