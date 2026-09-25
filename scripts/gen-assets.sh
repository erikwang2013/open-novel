#!/usr/bin/env bash
# 从 docs/pet.svg 派生全部衍生品：静止版 / 图标版 / 浏览器图标 / PWA 图标 / Flutter 贴图。
# 改了 pet.svg 就跑一次：bash scripts/gen-assets.sh
set -euo pipefail
cd "$(dirname "$0")/.."

CREAM='#FDF3E3' # pet.svg 背景圈同色，图标铺底后无接缝
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

# 静止版：把动画标签整棵摘掉。insert_comments 保住 SVG 内部的 MOOD 标记位；
# 根元素之前的注释 ElementTree 接不住，由脚本另写文件头。
python3 - docs/pet.svg docs/mascot.svg <<'PY'
import sys, xml.etree.ElementTree as ET
src, dst = sys.argv[1], sys.argv[2]
ET.register_namespace('', 'http://www.w3.org/2000/svg')
parser = ET.XMLParser(target=ET.TreeBuilder(insert_comments=True))
root = ET.parse(src, parser=parser).getroot()
parent = {c: p for p in root.iter() for c in p}
for el in [e for e in root.iter()
           if isinstance(e.tag, str) and e.tag.split('}')[-1].startswith('animate')]:
    parent[el].remove(el)
with open(dst, 'w', encoding='utf-8') as f:
    f.write('<?xml version="1.0" encoding="UTF-8"?>\n')
    f.write('''<!--
  墨灵（静止版）—— Open Novel 项目宠物。
  由 scripts/gen-assets.sh 从 docs/pet.svg 生成，勿手改；改源图后重跑脚本。
  会动的那版见 docs/pet.svg。
-->
''')
    f.write(ET.tostring(root, encoding='unicode') + '\n')
PY

# 图标版：画布裁到脸部（favicon 尺寸下整只墨灵糊成一团，只有脸看得清），
# 水印随画布一起挪进裁切框并缩到视觉等效大小（CLAUDE.md：自绘图像须带水印）。
sed -e 's|viewBox="0 0 512 512" width="512" height="512"|viewBox="160 130 192 192" width="512" height="512"|' \
    -e 's|x="440" y="498"|x="346" y="316"|' \
    -e 's|font-size="12"|font-size="5"|' \
    docs/pet.svg >docs/pet-icon.svg

if command -v rsvg-convert >/dev/null 2>&1; then
  svg2png() { rsvg-convert -w "$2" -h "$2" "$1" -o "$3"; }
else
  svg2png() { magick -background none -density 384 "$1" -resize "${2}x${2}" "$3"; }
fi

# 铺成不透明方图：pet.svg 背景圈外是透明，圆角/圆形裁切下会露白
flatten() { magick "$1" -background "$CREAM" -gravity center -flatten "$2"; }

# maskable：Android 会把图标裁成圆形，内容得留在中心 ~80% 安全区内
maskable() { # <px> <out>
  svg2png docs/pet.svg "$(( $1 * 2 ))" "$TMP/full.png"
  magick "$TMP/full.png" -resize "$(( $1 * 78 / 100 ))x" \
    -background "$CREAM" -gravity center -flatten -extent "${1}x${1}" "$2"
}

for app in apps/client/flutter apps/admin; do
  mkdir -p "$app/web/icons"

  svg2png docs/pet-icon.svg 32 "$app/web/favicon.png"
  cp docs/pet-icon.svg "$app/web/favicon.svg" # 现代浏览器直接用矢量版，标签页里会眨眼

  for px in 192 512; do
    svg2png docs/pet.svg "$px" "$TMP/i.png"
    flatten "$TMP/i.png" "$app/web/icons/Icon-$px.png"
    maskable "$px" "$app/web/icons/Icon-maskable-$px.png"
  done

  svg2png docs/pet.svg 420 "$app/assets/pet.png" # Flutter 空状态 / 登录页贴图
done

# 管理端 pubspec 另外声明了静止版贴图；声明了 Flutter 就要求文件存在，缺了构建会报错
svg2png docs/mascot.svg 420 apps/admin/assets/mascot.png

echo "衍生资源已生成：mascot.svg、pet-icon.svg、favicon.png+svg、PWA icons、pet.png、mascot.png"
