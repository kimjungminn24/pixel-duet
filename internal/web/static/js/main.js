import { state, setGrid } from "./state.js";
import { connect, send } from "./api.js";
import { draw, initCanvas } from "./canvas.js";
import { initTools, clearUndo } from "./tools.js";
import { initPalette, setPalette, renderPalette } from "./palette.js";
import { initChat, renderLog } from "./chat.js";
import { initSave } from "./save.js";
import { initKeyboard } from "./keyboard.js";
import { t, applyTranslations, onLanguageChange, toggleLang } from "./i18n.js";

const $ = (id) => document.getElementById(id);

const sizeSelect = $("size");
const speed = $("speed");
const speedLabel = $("speedVal");
const pauseBtn = $("pause");
const playBtn = $("play");

let live = null; // null until the event stream answers

function describeSpeed(ms) {
  return Number(ms) === 0 ? t("speed.none") : t("speed.per", { s: (Number(ms) / 1000).toFixed(2) });
}

function showLabels() {
  $("dimensions").textContent = `${state.w} × ${state.h} PX`;
  $("playState").textContent = live === null ? t("connecting")
    : t(state.paused ? "state.paused" : "state.playing");
  $("status").textContent = live === null ? t("connecting.dots")
    : t(live ? "connected" : "reconnecting");
  speedLabel.textContent = describeSpeed(speed.value);
}

function showState(m) {
  if (m.w !== state.w || m.h !== state.h) clearUndo();
  setGrid(m.w, m.h, m.grid);
  state.paused = m.paused;
  sizeSelect.value = m.w;
  pauseBtn.setAttribute("aria-pressed", m.paused);
  playBtn.setAttribute("aria-pressed", !m.paused);
  showLabels();
  draw();
}

function showConnection(isLive) {
  live = isLive;
  $("connection").dataset.state = isLive ? "live" : "offline";
  showLabels();
}

function initPlayback() {
  pauseBtn.onclick = () => send("pause");
  playBtn.onclick = () => send("resume");
  speed.oninput = () => { speedLabel.textContent = describeSpeed(speed.value); };
  speed.onchange = () => send("speed " + speed.value);
}

function initCanvasControls() {
  sizeSelect.onchange = () => {
    const n = sizeSelect.value;
    if (confirm(t("size.confirm", { n }))) send("size " + n);
    else sizeSelect.value = state.w;
  };
  $("clear").onclick = () => {
    if (confirm(t("clear.confirm"))) send("clear");
  };
}

function initLanguage() {
  applyTranslations();
  $("lang").onclick = toggleLang;
  onLanguageChange(() => {
    applyTranslations();
    showLabels();
    renderPalette();
  });
}

initLanguage();
initCanvas();
initTools();
initPalette();
initChat();
initSave();
initKeyboard();
initPlayback();
initCanvasControls();
showLabels();

connect({
  state: showState,
  log: (m) => renderLog(m.entries),
  palette: (m) => setPalette(m.colors),
  speed: (m) => { speed.value = m.ms; speedLabel.textContent = describeSpeed(m.ms); },
  open: () => showConnection(true),
  error: () => showConnection(false),
});

draw();
