const fs = require('fs')
const path = require('path')
const { execSync } = require('child_process')

const REPO = 'atbeta/picfast'
const BIN_NAME = process.platform === 'win32' ? 'picfast.exe' : 'picfast'

function getTarget() {
  const p = process.platform
  const a = process.arch
  if (p === 'darwin' && a === 'arm64') return 'darwin-arm64'
  if (p === 'darwin' && a === 'x64') return 'darwin-amd64'
  if (p === 'linux' && a === 'arm64') return 'linux-arm64'
  if (p === 'linux' && a === 'x64') return 'linux-amd64'
  if (p === 'win32' && a === 'x64') return 'windows-amd64'
  if (p === 'win32' && a === 'arm64') return 'windows-arm64'
  return null
}

const target = getTarget()
if (!target) {
  console.warn(`picfast: unsupported platform ${process.platform}-${process.arch}, skipping binary download`)
  process.exit(0)
}

const version = require('./package.json').version
const url = `https://github.com/${REPO}/releases/download/picfast-v${version}/picfast-${target}`
const binDir = path.join(__dirname, 'bin')
const binPath = path.join(binDir, BIN_NAME)

if (fs.existsSync(binPath)) {
  process.exit(0)
}

try {
  fs.mkdirSync(binDir, { recursive: true })
  console.log(`picfast: downloading ${url}`)
  execSync(`curl -fsSL "${url}" -o "${binPath}"`, { stdio: 'pipe' })
} catch {
  console.error('picfast: failed to download binary')
  process.exit(0)
}

try {
  fs.chmodSync(binPath, 0o755)
} catch {
  // ok
}
