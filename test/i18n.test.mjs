// The two languages have to cover the same keys, or a switch leaves a
// Korean word on an English page. Run with: node --test test/i18n.test.mjs
import { test } from "node:test";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { t, setLang } from "../internal/web/static/js/i18n.js";

const source = readFileSync(new URL("../internal/web/static/js/i18n.js", import.meta.url), "utf8");
const keysOf = (block) => [...block.matchAll(/^    "([^"]+)":/gm)].map((m) => m[1]);
const [ko, en] = source.split(/^  (?:ko|en): \{$/m).slice(1).map(keysOf);

test("ko and en define the same keys", () => {
  assert.deepEqual(ko, en);
});

test("every key the page names is defined", () => {
  const html = readFileSync(new URL("../internal/web/static/index.html", import.meta.url), "utf8");
  const named = [...html.matchAll(/data-i18n(?:-title|-aria|-placeholder)?="([^"]+)"/g)].map((m) => m[1]);
  for (const key of named) assert.ok(ko.includes(key), key);
});

test("t fills slots and falls back to the key", () => {
  setLang("en");
  assert.equal(t("size.confirm", { n: 64 }), "Start a new 64 × 64 canvas?");
  assert.equal(t("no.such.key"), "no.such.key");
  setLang("ko");
  assert.equal(t("speed.per", { s: "0.50" }), "한 획마다 0.50초 쉬기");
});
