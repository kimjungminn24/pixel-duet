import { state, DIGITS, CHECKER } from "./state.js";
import { send } from "./api.js";
import { draw } from "./canvas.js";
import { t } from "./i18n.js";

const swatches = document.getElementById("palette");
const chip = document.getElementById("selectedChip");
const hex = document.getElementById("selectedHex");
const panel = document.getElementById("colorPanel");
const minus = document.getElementById("paletteMinus");
const plus = document.getElementById("palettePlus");

const COLUMNS = [8, 6, 4, 3];
let sizeStep = 0;

const checkerFill = `repeating-conic-gradient(${CHECKER[0]} 0 25%, ${CHECKER[1]} 0 50%) 0 0 / 12px 12px`;

export function selectColor(i) {
  state.color = i;
  renderPalette();
}

export function renderPalette() {
  const { palette, color } = state;
  hex.textContent = palette[color] ? palette[color].toUpperCase() : "ERASER";
  chip.style.background = palette[color] || "transparent";
  swatches.replaceChildren(...palette.map((c, i) => {
    const b = document.createElement("button");
    b.className = "swatch" + (i === color ? " sel" : "");
    b.style.background = c || checkerFill;
    b.title = i === 0 ? t("palette.eraser") : t("palette.color", { digit: DIGITS[i], hex: c });
    b.setAttribute("aria-label", b.title);
    b.setAttribute("aria-pressed", i === color);
    b.onclick = () => selectColor(i);
    return b;
  }));
}

function renderColorPanel() {
  const inputs = [];
  for (let i = 1; i < state.palette.length; i++) {
    const label = document.createElement("label");
    const input = document.createElement("input");
    input.type = "color";
    input.value = state.palette[i];
    input.oninput = () => send(`setcolor ${i} ${input.value.slice(1)}`);
    label.append(input, document.createTextNode(DIGITS[i]));
    inputs.push(label);
  }
  panel.replaceChildren(...inputs);
}

export function setPalette(colors) {
  state.palette = [null, ...colors.slice(1)];
  renderPalette();
  renderColorPanel();
  draw();
}

function resizeSwatches(step) {
  sizeStep = Math.max(0, Math.min(COLUMNS.length - 1, sizeStep + step));
  swatches.style.setProperty("--palette-columns", COLUMNS[sizeStep]);
  minus.disabled = sizeStep === 0;
  plus.disabled = sizeStep === COLUMNS.length - 1;
}

export function initPalette() {
  minus.onclick = () => resizeSwatches(-1);
  plus.onclick = () => resizeSwatches(1);
  resizeSwatches(0);
  document.getElementById("colors").onclick = function () {
    this.setAttribute("aria-expanded", panel.classList.toggle("open"));
  };
  renderPalette();
  renderColorPanel();
}
