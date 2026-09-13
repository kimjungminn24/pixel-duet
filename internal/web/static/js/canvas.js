import { state, CHECKER } from "./state.js";

export const cv = document.getElementById("cv");
const ctx = cv.getContext("2d");
const board = document.querySelector(".board");
const zoom = document.getElementById("zoom");
const zoomValue = document.getElementById("zoomValue");
const panBtn = document.getElementById("pan");

// Zoom slider is log2 of screen pixels per canvas pixel.
const MAX_ZOOM = 7;
const ZOOM_STEP = .5;

// Backing store long side; CSS scales the element to fit.
const BACKING = 1024;

let fitZoom = true;
let panMode = false;
let spacePan = false;
let panStart = null;
let scale = 16;

const clampZoom = (v) => Math.max(0, Math.min(MAX_ZOOM, v));

function fitToBoard() {
  const { w, h } = state;
  const cs = getComputedStyle(board);
  const aw = board.clientWidth - parseFloat(cs.paddingLeft) - parseFloat(cs.paddingRight);
  const ah = board.clientHeight - parseFloat(cs.paddingTop) - parseFloat(cs.paddingBottom);
  const width = fitZoom ? Math.max(1, Math.min(aw, ah * w / h)) : w * 2 ** Number(zoom.value);
  cv.style.width = width + "px";
  cv.style.height = width * h / w + "px";
  if (fitZoom) zoom.value = clampZoom(Math.log2(width / w));
  zoomValue.textContent = Math.round(width / w * 100) + "%";
}

export function draw() {
  fitToBoard();
  const { w, h, cells, palette, preview } = state;
  scale = Math.max(2, Math.floor(BACKING / Math.max(w, h)));
  cv.width = w * scale;
  cv.height = h * scale;
  for (let y = 0; y < h; y++) {
    for (let x = 0; x < w; x++) {
      const i = cells[y * w + x];
      ctx.fillStyle = i > 0 ? palette[i] : CHECKER[((x >> 1) + (y >> 1)) % 2];
      ctx.fillRect(x * scale, y * scale, scale, scale);
    }
  }
  if (preview) {
    ctx.globalAlpha = 0.6;
    ctx.fillStyle = preview.color > 0 ? palette[preview.color] : CHECKER[0];
    for (const [x, y] of preview.pts) ctx.fillRect(x * scale, y * scale, scale, scale);
    ctx.globalAlpha = 1;
  }
  if (state.gridOn && scale >= 6) drawGridLines(w, h);
}

// Line width is one screen pixel, not one backing pixel, so it survives downscaling.
function drawGridLines(w, h) {
  const shown = cv.getBoundingClientRect().width || cv.width;
  ctx.strokeStyle = "rgba(41,45,39,0.18)";
  ctx.lineWidth = Math.max(1, (w * scale) / shown);
  ctx.beginPath();
  for (let x = 1; x < w; x++) {
    ctx.moveTo(x * scale + .5, 0);
    ctx.lineTo(x * scale + .5, h * scale);
  }
  for (let y = 1; y < h; y++) {
    ctx.moveTo(0, y * scale + .5);
    ctx.lineTo(w * scale, y * scale + .5);
  }
  ctx.stroke();
}

export function cellOf(ev) {
  const r = cv.getBoundingClientRect();
  const x = Math.floor((ev.clientX - r.left) * state.w / r.width);
  const y = Math.floor((ev.clientY - r.top) * state.h / r.height);
  return (x < 0 || x >= state.w || y < 0 || y >= state.h) ? null : [x, y];
}

// Keeps the board's center point fixed across the size change.
function changeZoom(value) {
  const before = cv.getBoundingClientRect();
  const b = board.getBoundingClientRect();
  const rx = (b.left + board.clientWidth / 2 - before.left) / before.width;
  const ry = (b.top + board.clientHeight / 2 - before.top) / before.height;
  fitZoom = false;
  zoom.value = clampZoom(value);
  draw();
  const after = cv.getBoundingClientRect();
  board.scrollLeft += after.left + rx * after.width - b.left - board.clientWidth / 2;
  board.scrollTop += after.top + ry * after.height - b.top - board.clientHeight / 2;
}

function showPan() {
  board.classList.toggle("panning", panMode || spacePan);
  panBtn.setAttribute("aria-pressed", panMode);
}

export function setPanMode(on) {
  panMode = on;
  showPan();
}

export function togglePanMode() {
  setPanMode(!panMode);
}

export function setSpacePan(on) {
  spacePan = on;
  showPan();
}

function startPan(ev) {
  if (!(panMode || spacePan || ev.button === 1)) return;
  ev.preventDefault();
  ev.stopPropagation();
  panStart = { x: ev.clientX, y: ev.clientY, left: board.scrollLeft, top: board.scrollTop, id: ev.pointerId };
  board.setPointerCapture(ev.pointerId);
  board.classList.add("dragging");
}

function movePan(ev) {
  if (!panStart) return;
  ev.preventDefault();
  ev.stopPropagation();
  board.scrollLeft = panStart.left - ev.clientX + panStart.x;
  board.scrollTop = panStart.top - ev.clientY + panStart.y;
}

function endPan() {
  if (panStart && board.hasPointerCapture(panStart.id)) board.releasePointerCapture(panStart.id);
  panStart = null;
  board.classList.remove("dragging");
}

export function initCanvas() {
  zoom.oninput = () => changeZoom(Number(zoom.value));
  document.getElementById("zoomIn").onclick = () => changeZoom(Number(zoom.value) + ZOOM_STEP);
  document.getElementById("zoomOut").onclick = () => changeZoom(Number(zoom.value) - ZOOM_STEP);
  document.getElementById("zoomFit").onclick = () => { fitZoom = true; board.scrollTo(0, 0); draw(); };
  panBtn.onclick = togglePanMode;

  // Capture phase: a pan must swallow the event before the canvas paints.
  board.addEventListener("pointerdown", startPan, true);
  board.addEventListener("pointermove", movePan, true);
  board.addEventListener("pointerup", endPan);
  board.addEventListener("pointercancel", endPan);
  addEventListener("blur", () => { spacePan = false; endPan(); showPan(); });

  new ResizeObserver(draw).observe(board);
}
