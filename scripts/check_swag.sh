#!/usr/bin/env bash
# 校验 Gin handler 数量与 @Router 注解数量一致
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

handlers=$(grep -RhE 'func \(h \*[^)]+\) [A-Z][a-zA-Z0-9]+\(c \*gin\.Context\)' admin api 2>/dev/null | wc -l)
routers=$(grep -Rh '@Router' admin api 2>/dev/null | wc -l)

echo "handlers: $handlers, @Router: $routers"
if [ "$handlers" -ne "$routers" ]; then
  echo "ERROR: handler 与 @Router 数量不一致，请补全 swag 注解" >&2
  exit 1
fi

if [ ! -f docs/swagger/swagger.yaml ]; then
  echo "ERROR: docs/swagger/swagger.yaml 不存在，请运行 make swag" >&2
  exit 1
fi

paths=$(grep -c '^  /' docs/swagger/swagger.yaml 2>/dev/null || echo 0)
if [ "$paths" -eq 0 ]; then
  echo "ERROR: swagger paths 为空，请检查 swag 注解" >&2
  exit 1
fi

echo "OK: swagger 文档包含 $paths 个路径"
