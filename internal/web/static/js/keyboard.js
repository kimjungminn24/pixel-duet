import { TOOL_KEYS, setTool, toggleGrid, undo } from "./tools.js";
import { togglePanMode, setSpacePan } from "./canvas.js";

const inField = (ev) => ev.target.closest("input, textarea, select, [contenteditable]");

function onKeyDown(ev) {
  if (ev.isComposing || inField(ev)) return;
  const key = ev.key.toLowerCase();
  if ((ev.ctrlKey || ev.metaKey) && key === "z") {
    ev.preventDefault();
    undo();
    return;
  }
  if (ev.ctrlKey || ev.metaKey || ev.altKey) return;
  if (ev.code === "Space") {
    ev.preventDefault();
    setSpacePan(true);
  } else if (key === "h") {
    if (!ev.repeat) togglePanMode();
  } else if (key === "g") {
    ev.preventDefault();
    toggleGrid();
  } else if (TOOL_KEYS[key]) {
    ev.preventDefault();
    setTool(TOOL_KEYS[key]);
  }
}

function onKeyUp(ev) {
  if (ev.code === "Space") setSpacePan(false);
}

export function initKeyboard() {
  addEventListener("keydown", onKeyDown);
  addEventListener("keyup", onKeyUp);
}
