#!/usr/bin/env node
// Runs the platform binary with this process's arguments; nothing is written to stdout.
"use strict";

const fs = require("fs");
const { spawn } = require("child_process");
const { install, binaryPath } = require("../install.js");

async function main() {
  let exe = binaryPath();
  if (!fs.existsSync(exe)) {
    exe = await install(); // postinstall was skipped (--ignore-scripts)
  }
  const child = spawn(exe, process.argv.slice(2), { stdio: "inherit" });
  for (const sig of ["SIGINT", "SIGTERM"]) {
    process.on(sig, () => child.kill(sig));
  }
  child.on("error", (err) => {
    process.stderr.write(`pixelduet: ${err.message}\n`);
    process.exit(1);
  });
  child.on("exit", (code, signal) => {
    process.exit(signal ? 1 : code ?? 1);
  });
}

main().catch((err) => {
  process.stderr.write(`pixelduet: ${err.message}\n`);
  process.exit(1);
});
