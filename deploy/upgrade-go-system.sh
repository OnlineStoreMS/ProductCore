#!/usr/bin/env bash
# 将系统 /usr/local/go 升级为 Go 1.25（需 sudo）
# 用法: sudo bash deploy/upgrade-go-system.sh
# 可选: sudo bash deploy/upgrade-go-system.sh /path/to/go1.25.0.linux-amd64.tar.gz

set -euo pipefail

VERSION=1.25.0
ARCH=linux-amd64
TARBALL="go${VERSION}.${ARCH}.tar.gz"
URL="https://go.dev/dl/${TARBALL}"

if [[ "${EUID:-$(id -u)}" -ne 0 ]]; then
  echo "请使用 sudo 运行: sudo bash $0"
  exit 1
fi

cleanup() {
  [[ -n "${TMP:-}" && -f "${TMP:-}" ]] && rm -f "$TMP"
}
TMP=""

download_to() {
  local dest="$1"
  if command -v wget >/dev/null 2>&1; then
    wget -q --show-progress -O "$dest" "$URL"
  elif command -v curl >/dev/null 2>&1; then
    curl -fsSL --progress-bar -o "$dest" "$URL"
  else
    echo "错误: 需要 wget 或 curl"
    exit 1
  fi
}

resolve_tarball() {
  if [[ -n "${1:-}" ]]; then
    if [[ ! -f "$1" ]]; then
      echo "错误: 找不到安装包 $1" >&2
      exit 1
    fi
    echo "$1"
    return
  fi

  # 优先复用普通用户已下载的安装包（避免 root 覆盖 /tmp 下他人文件）
  local pkg user_home
  if [[ -n "${SUDO_USER:-}" ]]; then
    user_home="$(getent passwd "$SUDO_USER" | cut -d: -f6)"
    if [[ -n "$user_home" && -f "${user_home}/${TARBALL}" ]]; then
      echo "使用已有安装包: ${user_home}/${TARBALL}" >&2
      echo "${user_home}/${TARBALL}"
      return
    fi
  fi
  for pkg in "/tmp/${TARBALL}"; do
    if [[ -f "$pkg" && -r "$pkg" ]]; then
      echo "使用已有安装包: ${pkg}" >&2
      echo "$pkg"
      return
    fi
  done

  TMP="$(mktemp /var/tmp/go-install.XXXXXX.tar.gz)"
  trap cleanup EXIT
  echo "下载 Go ${VERSION} ..." >&2
  download_to "$TMP"

  if [[ ! -s "$TMP" ]]; then
    echo "错误: 下载失败或文件为空" >&2
    exit 1
  fi
  echo "$TMP"
}

TARBALL_PATH="$(resolve_tarball "${1:-}")"

echo "安装到 /usr/local/go ..."
rm -rf /usr/local/go
tar -C /usr/local -xzf "$TARBALL_PATH"

if [[ ! -x /usr/local/go/bin/go ]]; then
  echo "错误: 安装后未找到 /usr/local/go/bin/go"
  exit 1
fi

cat >/etc/profile.d/go.sh <<'EOF'
# Go 1.25 — 不要设置 GOROOT，由 go 命令自动识别
export PATH=/usr/local/go/bin:$PATH
EOF

chmod 644 /etc/profile.d/go.sh
rm -f /etc/profile.d/go1_21.sh

echo ""
echo "完成:"
/usr/local/go/bin/go version
echo "GOROOT=$(env -u GOROOT /usr/local/go/bin/go env GOROOT)"
echo ""
echo "请在新终端执行: source /etc/profile  或重新登录"
echo "若曾配置 ~/.local/go，可删除该目录并清理 ~/.bashrc 中的 Go 覆盖项。"
