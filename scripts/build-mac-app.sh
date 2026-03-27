#!/bin/bash
set -e

VERSION=${1:-"dev"}
BINARY="./fracture"
APP_NAME="FRACTURE"
APP_DIR="dist/${APP_NAME}.app"

echo "► Criando ${APP_NAME}.app v${VERSION}..."

# Criar estrutura
mkdir -p "${APP_DIR}/Contents/MacOS"
mkdir -p "${APP_DIR}/Contents/Resources"

# Copiar binário
cp "${BINARY}" "${APP_DIR}/Contents/MacOS/fracture"
chmod +x "${APP_DIR}/Contents/MacOS/fracture"

# Criar Info.plist
cat > "${APP_DIR}/Contents/Info.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleExecutable</key>
  <string>fracture</string>
  <key>CFBundleIdentifier</key>
  <string>com.fracture.app</string>
  <key>CFBundleName</key>
  <string>FRACTURE</string>
  <key>CFBundleDisplayName</key>
  <string>FRACTURE</string>
  <key>CFBundleVersion</key>
  <string>${VERSION}</string>
  <key>CFBundleShortVersionString</key>
  <string>${VERSION}</string>
  <key>CFBundlePackageType</key>
  <string>APPL</string>
  <key>CFBundleSignature</key>
  <string>FRCT</string>
  <key>LSMinimumSystemVersion</key>
  <string>12.0</string>
  <key>LSUIElement</key>
  <false/>
  <key>NSHighResolutionCapable</key>
  <true/>
  <key>CFBundleIconFile</key>
  <string>AppIcon</string>
</dict>
</plist>
EOF

echo "✓ Info.plist criado"

# Criar ícone simples (placeholder)
# Ícone real requer iconutil — apenas documenta se não disponível
if command -v iconutil &> /dev/null && [ -f "assets/icon.png" ]; then
  mkdir -p /tmp/fracture.iconset
  sips -z 16 16     assets/icon.png --out /tmp/fracture.iconset/icon_16x16.png
  sips -z 32 32     assets/icon.png --out /tmp/fracture.iconset/icon_16x16@2x.png
  sips -z 32 32     assets/icon.png --out /tmp/fracture.iconset/icon_32x32.png
  sips -z 64 64     assets/icon.png --out /tmp/fracture.iconset/icon_32x32@2x.png
  sips -z 128 128   assets/icon.png --out /tmp/fracture.iconset/icon_128x128.png
  sips -z 256 256   assets/icon.png --out /tmp/fracture.iconset/icon_128x128@2x.png
  sips -z 256 256   assets/icon.png --out /tmp/fracture.iconset/icon_256x256.png
  sips -z 512 512   assets/icon.png --out /tmp/fracture.iconset/icon_256x256@2x.png
  sips -z 512 512   assets/icon.png --out /tmp/fracture.iconset/icon_512x512.png
  iconutil -c icns /tmp/fracture.iconset -o "${APP_DIR}/Contents/Resources/AppIcon.icns"
  echo "✓ Ícone criado"
fi

# Assinar se certificado disponível
if [ -n "$APPLE_DEVELOPER_ID" ]; then
  echo "► Assinando com ${APPLE_DEVELOPER_ID}..."
  codesign --force --deep --sign "${APPLE_DEVELOPER_ID}" "${APP_DIR}"
  echo "✓ App assinado"
else
  echo "⚠ Sem APPLE_DEVELOPER_ID — app não assinado"
  echo "  Usuários precisarão: clique direito → Abrir"
fi

# Criar DMG para distribuição
if command -v hdiutil &> /dev/null; then
  echo "► Criando DMG..."
  DMG_PATH="dist/FRACTURE-${VERSION}-macOS.dmg"
  hdiutil create -volname "FRACTURE" \
    -srcfolder "dist/${APP_NAME}.app" \
    -ov -format UDZO \
    "${DMG_PATH}" 2>/dev/null
  echo "✓ DMG criado: ${DMG_PATH}"
fi

echo ""
echo "✅ FRACTURE.app pronto em dist/"
echo "   Arraste para Applications ou dê duplo clique para rodar."
