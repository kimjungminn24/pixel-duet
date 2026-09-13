// Downloads the pixelduet binary for this platform from the GitHub release matching package.json.
// stderr only: the bridge subcommand speaks JSON-RPC on stdout.
"use strict";

const fs = require("fs");
const path = require("path");
const pkg = require("./package.json");

const VENDOR = path.join(__dirname, "vendor");

function target() {
  const os = { linux: "linux", darwin: "darwin", win32: "windows" }[process.platform];
  const arch = { x64: "amd64", arm64: "arm64" }[process.arch];
  if (!os || !arch) {
    throw new Error(`no pixelduet build for ${process.platform}/${process.arch}`);
  }
  return { os, arch, ext: os === "windows" ? ".exe" : "" };
}

function binaryPath() {
  return path.join(VENDOR, "pixelduet" + target().ext);
}

function releaseUrl() {
  const repo = pkg.repository.url.replace(/^.*github\.com\//, "").replace(/\.git$/, "");
  const { os, arch, ext } = target();
  return `https://github.com/${repo}/releases/download/v${pkg.version}/pixelduet-${os}-${arch}${ext}`;
}

async function install() {
  const url = releaseUrl();
  const dest = binaryPath();
  process.stderr.write(`pixelduet: downloading ${url}\n`);
  const res = await fetch(url);
  if (!res.ok) {
    throw new Error(`download failed: ${res.status} ${res.statusText} for ${url}`);
  }
  fs.mkdirSync(VENDOR, { recursive: true });
  const tmp = dest + ".part";
  fs.writeFileSync(tmp, Buffer.from(await res.arrayBuffer()));
  fs.chmodSync(tmp, 0o755);
  fs.renameSync(tmp, dest);
  return dest;
}

module.exports = { install, binaryPath };

if (require.main === module) {
  install().catch((err) => {
    process.stderr.write(`pixelduet: ${err.message}\n`);
    process.exit(1);
  });
}
