import { state, digitAt } from "./state.js";
import { send, sendStroke } from "./api.js";
import { cv, cellOf, draw, setPanMode } from "./canvas.js";
import { lineCells, shapeCells, floodCells } from "./geometry.js";
import { selectColor } from "./palette.js";

export const TOOL_KEYS = {
  b: "pen", e: "eraser", l: "line", r: "rect", u: "solidRect",
  o: "circle", p: "solidCircle", f: "fill", i: "pick",
};

const coordinates = document.getElementById("coordinates");
const gridBtn = document.getElementById("gridBtn");

// Undo replays the overwritten colors as ordinary strokes; the server has no undo.
const undoStack = [];
let gesture = null; // {cells: [{x, y, prev}], seen: Set} while a pointer is down

let down = false;
let button = 0;
let start = null;
let last = null;

function remember(pts) {
  for (const [x, y] of pts) {
    const key = y * state.w + x;
    if (gesture.seen.has(key)) continue;
    gesture.seen.add(key);
    gesture.cells.push({ x, y, prev: digitAt(x, y) });
  }
}

function paint(color, pts) {
  remember(pts);
  sendStroke(color, pts);
}

function endGesture() {
  if (gesture?.cells.length) undoStack.push(gesture.cells);
  gesture = null;
}

export function undo() {
  const cells = undoStack.pop();
  if (!cells) return;
  const byPrev = new Map();
  for (const { x, y, prev } of cells) {
    if (!byPrev.has(prev)) byPrev.set(prev, []);
    byPrev.get(prev).push([x, y]);
  }
  for (const [prev, pts] of byPrev) sendStroke(prev, pts);
}

export function clearUndo() {
  undoStack.length = 0;
}

export function setTool(tool) {
  state.tool = tool;
  setPanMode(false);
  for (const b of document.querySelectorAll("#toolbar button")) {
    const active = b.id === "t-" + tool;
    b.classList.toggle("active", active);
    b.setAttribute("aria-pressed", active);
  }
}

export function toggleGrid() {
  state.gridOn = !state.gridOn;
  gridBtn.classList.toggle("active", state.gridOn);
  gridBtn.setAttribute("aria-pressed", state.gridOn);
  draw();
}

// Right mouse button erases.
function brushColor() {
  return button === 2 || state.tool === "eraser" ? 0 : state.color;
}

function shapePreview(color, x1, y1) {
  state.preview = { color, pts: shapeCells(state.tool, ...start, x1, y1, state.w, state.h) };
  draw();
}

function onPointerDown(ev) {
  if (ev.button !== 0 && ev.button !== 2) return;
  const cell = cellOf(ev);
  if (!cell) return;
  cv.setPointerCapture(ev.pointerId);
  down = true;
  button = ev.button;
  start = last = cell;
  gesture = { cells: [], seen: new Set() };
  const color = brushColor();
  switch (state.tool) {
    case "pen":
    case "eraser":
      paint(color, [cell]);
      break;
    case "fill":
      if (digitAt(...cell) !== color) {
        remember(floodCells(state.cells, state.w, state.h, ...cell));
        send(`fill ${cell[0]} ${cell[1]} ${color}`);
      }
      break;
    case "pick":
      selectColor(digitAt(...cell));
      down = false;
      gesture = null;
      setTool("pen");
      break;
    default:
      shapePreview(color, ...cell);
  }
}

function onPointerMove(ev) {
  const cell = cellOf(ev);
  if (!cell) return;
  coordinates.textContent = `(${cell[0]}, ${cell[1]})`;
  if (!down) return;
  const color = brushColor();
  if (state.tool === "pen" || state.tool === "eraser") {
    paint(color, lineCells(...last, ...cell));
    last = cell;
  } else if (state.tool !== "fill" && state.tool !== "pick") {
    shapePreview(color, ...cell);
  }
}

function onPointerUp() {
  if (!down) return;
  down = false;
  if (state.preview) {
    paint(state.preview.color, state.preview.pts);
    state.preview = null;
    draw();
  }
  endGesture();
}

function onPointerCancel() {
  down = false;
  state.preview = null;
  endGesture();
  draw();
}

export function initTools() {
  for (const tool of Object.values(TOOL_KEYS)) {
    document.getElementById("t-" + tool).onclick = () => setTool(tool);
  }
  gridBtn.onclick = toggleGrid;
  document.getElementById("undo").onclick = undo;

  cv.addEventListener("pointerdown", onPointerDown);
  cv.addEventListener("pointermove", onPointerMove);
  cv.addEventListener("pointercancel", onPointerCancel);
  cv.addEventListener("pointerleave", () => { coordinates.textContent = "(-, -)"; });
  cv.addEventListener("contextmenu", (ev) => ev.preventDefault());
  addEventListener("pointerup", onPointerUp);
}
