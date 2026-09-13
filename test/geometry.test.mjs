import { test } from "node:test";
import assert from "node:assert/strict";
import { lineCells, rectCells, shapeCells, floodCells } from "../internal/web/static/js/geometry.js";

const has = (pts, x, y) => pts.some(([px, py]) => px === x && py === y);

test("a line touches every column and both ends", () => {
  const pts = lineCells(2, 1, 9, 4);
  assert.equal(pts.length, 8);
  assert.ok(has(pts, 2, 1) && has(pts, 9, 4));
  for (let x = 2; x <= 9; x++) assert.ok(pts.some(([px]) => px === x), `column ${x}`);
  assert.equal(lineCells(9, 4, 2, 1).length, pts.length);
});

test("a rectangle outline has no inside", () => {
  assert.equal(rectCells(4, 4, 1, 2).length, 10);
  assert.equal(shapeCells("solidRect", 1, 2, 4, 4, 8, 8).length, 12);
});

test("a circle is round and stays square near the edge", () => {
  const rows = new Map();
  for (const [, y] of shapeCells("solidCircle", 7, 7, 13, 13, 16, 16)) rows.set(y, (rows.get(y) ?? 0) + 1);
  assert.deepEqual([...rows.values()], [3, 5, 7, 7, 7, 5, 3]);
  assert.equal(shapeCells("circle", 7, 7, 13, 13, 16, 16).length, 16);

  const clamped = shapeCells("solidCircle", 6, 0, 30, 3, 8, 8);
  assert.ok(clamped.every(([x, y]) => x <= 7 && y <= 1));
});

test("a flood fill stops at another color", () => {
  const w = 3, h = 3, cells = new Uint8Array(w * h);
  for (let y = 0; y < h; y++) cells[y * w + 1] = 7; // a wall down the middle
  const left = floodCells(cells, w, h, 0, 0);
  assert.equal(left.length, 3);
  assert.ok(left.every(([x]) => x === 0));
  assert.deepEqual(floodCells(cells, w, h, 9, 9), []);
});
