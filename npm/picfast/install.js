const fs = require('fs')
const path = require('path')
const https = require('https')
const http = require('http')
const { URL } = require('url')

const REPO = 'atbeta/picfast'
const MIN_BINARY_BYTES = 1024 * 1024
const MAX_REDIRECTS = 10

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

function downloadFile(url, dest, redirects = 0) {
  return new Promise((resolve, reject) => {
    let parsed
    try {
      parsed = new URL(url)
    } catch (err) {
      reject(err)
      return
    }
    const lib = parsed.protocol === 'http:' ? http : https
    const req = lib.get(
      url,
      {
        headers: {
          'User-Agent': 'picfast-npm-install',
          Accept: 'application/octet-stream',
        },
      },
      (res) => {
        const status = res.statusCode || 0
        if (
          status >= 300 &&
          status < 400 &&
          res.headers.location &&
          redirects < MAX_REDIRECTS
        ) {
          res.resume()
          const next = new URL(res.headers.location, url).toString()
          downloadFile(next, dest, redirects + 1).then(resolve, reject)
          return
        }
        if (status !== 200) {
          res.resume()
          reject(new Error(`HTTP ${status} for ${url}`))
          return
        }

        const tmp = `${dest}.download`
        const out = fs.createWriteStream(tmp)
        res.pipe(out)
        out.on('finish', () => {
          out.close((closeErr) => {
            if (closeErr) {
              fs.unlink(tmp, () => reject(closeErr))
              return
            }
            try {
              fs.renameSync(tmp, dest)
              resolve()
            } catch (renameErr) {
              // Windows: replace existing file if rename across same dir fails.
              try {
                fs.copyFileSync(tmp, dest)
                fs.unlinkSync(tmp)
                resolve()
              } catch (copyErr) {
                fs.unlink(tmp, () => reject(copyErr || renameErr))
              }
            }
          })
        })
        out.on('error', (err) => {
          fs.unlink(tmp, () => reject(err))
        })
      }
    )
    req.on('error', reject)
  })
}

async function main() {
  const target = getTarget()
  if (!target) {
    console.warn(
      `picfast: unsupported platform ${process.platform}-${process.arch}, skipping binary download`
    )
    return
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
    return
  }

  fs.mkdirSync(binDir, { recursive: true })
  console.log(`picfast: downloading ${url}`)
  try {
    await downloadFile(url, binPath)
  } catch (err) {
    console.error('picfast: failed to download binary from', url)
    console.error(String(err && err.message ? err.message : err))
    process.exitCode = 1
    return
  }

  let st
  try {
    st = fs.statSync(binPath)
  } catch {
    console.error('picfast: downloaded binary missing after download')
    process.exitCode = 1
    return
  }
  if (st.size < MIN_BINARY_BYTES) {
    console.error(`picfast: downloaded file looks too small (${st.size} bytes)`)
    process.exitCode = 1
    return
  }

  try {
    fs.chmodSync(binPath, 0o755)
  } catch {
    // Windows does not use Unix execute bits.
  }
}

main().catch((err) => {
  console.error('picfast: install failed:', err)
  process.exitCode = 1
})
