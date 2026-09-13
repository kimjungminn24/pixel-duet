// Mirrors shapes.go so the browser and the model draw the same pixels.

export function lineCells(x0, y0, x1, y1) {
  const pts = [];
  const dx = Math.abs(x1 - x0), dy = -Math.abs(y1 - y0);
  const sx = x0 < x1 ? 1 : -1, sy = y0 < y1 ? 1 : -1;
  let e = dx + dy;
  for (;;) {
    pts.push([x0, y0]);
    if (x0 === x1 && y0 === y1) return pts;
    const e2 = 2 * e;
    if (e2 >= dy) { e += dy; x0 += sx; }
    if (e2 <= dx) { e += dx; y0 += sy; }
  }
}

export function rectCells(x0, y0, x1, y1) {
  const left = Math.min(x0, x1), right = Math.max(x0, x1);
  const top = Math.min(y0, y1), bottom = Math.max(y0, y1);
  const pts = [];
  for (let x = left; x <= right; x++) pts.push([x, top], [x, bottom]);
  for (let y = top + 1; y < bottom; y++) pts.push([left, y], [right, y]);
  return pts;
}

// Circles are kept square and clamped to the w by h canvas.
export function shapeCells(kind, x0, y0, x1, y1, w, h) {
  if (kind === "line") return lineCells(x0, y0, x1, y1);
  if (kind === "rect") return rectCells(x0, y0, x1, y1);
  if (kind === "circle" || kind === "solidCircle") {
    const sx = x1 < x0 ? -1 : 1, sy = y1 < y0 ? -1 : 1;
    const room = Math.min(sx > 0 ? w - 1 - x0 : x0, sy > 0 ? h - 1 - y0 : y0);
    const side = Math.min(Math.max(Math.abs(x1 - x0), Math.abs(y1 - y0)), room);
    x1 = x0 + sx * side;
    y1 = y0 + sy * side;
  }
  const left = Math.min(x0, x1), right = Math.max(x0, x1);
  const top = Math.min(y0, y1), bottom = Math.max(y0, y1);
  const pts = [];
  if (kind === "solidRect") {
    for (let y = top; y <= bottom; y++) for (let x = left; x <= right; x++) pts.push([x, y]);
    return pts;
  }
  const cx = (left + right) / 2, cy = (top + bottom) / 2;
  const rx = (right - left + 1) / 2 - .25, ry = (bottom - top + 1) / 2 - .25;
  const inside = (x, y) => ((x - cx) / rx) ** 2 + ((y - cy) / ry) ** 2 <= 1;
  const filled = kind === "solidCircle";
  for (let y = top; y <= bottom; y++) {
    for (let x = left; x <= right; x++) {
      if (!inside(x, y)) continue;
      const edge = !inside(x - 1, y) || !inside(x + 1, y) || !inside(x, y - 1) || !inside(x, y + 1);
      if (filled || edge) pts.push([x, y]);
    }
  }
  return pts;
}

export function floodCells(cells, w, h, x, y) {
  if (x < 0 || x >= w || y < 0 || y >= h) return [];
  const target = cells[y * w + x];
  const seen = new Uint8Array(w * h);
  const stack = [[x, y]], out = [];
  while (stack.length) {
    const [px, py] = stack.pop();
    if (px < 0 || px >= w || py < 0 || py >= h) continue;
    const k = py * w + px;
    if (seen[k] || cells[k] !== target) continue;
    seen[k] = 1;
    out.push([px, py]);
    stack.push([px + 1, py], [px - 1, py], [px, py + 1], [px, py - 1]);
  }
  return out;
}
