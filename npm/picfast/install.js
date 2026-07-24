const fs = require('fs')
const path = require('path')
const { execSync } = require('child_process')

const REPO = 'atbeta/picfast'
const MIN_BINARY_BYTES = 1024 * 1024

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

function needsDownload(binPath) {
  try {
    return fs.statSync(binPath).size < MIN_BINARY_BYTES
  } catch {
    return true
  }
}

const target = getTarget()
if (!target) {
  console.warn(
    `picfast: unsupported platform ${process.platform}-${process.arch}, skipping binary download`
  )
  process.exit(0)
}

const version = require('./package.json').version
const asset = target.startsWith('windows-')
  ? `picfast-${target}.exe`
  : `picfast-${target}`
const url = `https://github.com/${REPO}/releases/download/picfast-v${version}/${asset}`
const binDir = path.join(__dirname, 'bin')
const binName = process.platform === 'win32' ? 'picfast.exe' : 'picfast-native'
const binPath = path.join(binDir, binName)

if (!needsDownload(binPath)) {
  process.exit(0)
}

try {
  fs.mkdirSync(binDir, { recursive: true })
  console.log(`picfast: downloading ${url}`)
  execSync(`curl -fsSL "${url}" -o "${binPath}"`, { stdio: 'pipe' })
} catch (err) {
  console.error('picfast: failed to download binary from', url)
  if (err && err.stderr) {
    process.stderr.write(String(err.stderr))
  }
  process.exit(1)
}

try {
  const st = fs.statSync(binPath)
  if (st.size < MIN_BINARY_BYTES) {
    console.error(`picfast: downloaded file looks too small (${st.size} bytes)`)
    process.exit(1)
  }
} catch {
  console.error('picfast: downloaded binary missing after curl')
  process.exit(1)
}

try {
  fs.chmodSync(binPath, 0o755)
} catch {
  // ok
}
