import { send } from "./api.js";

const logEl = document.getElementById("log");
const note = document.getElementById("note");

export function renderLog(entries) {
  logEl.replaceChildren(...entries.flatMap((e) => {
    const who = document.createElement("div");
    who.className = "who " + e.who;
    who.textContent = e.who;
    const msg = document.createElement("div");
    msg.className = "msg";
    msg.textContent = e.text;
    return [who, msg];
  }));
  logEl.scrollTop = logEl.scrollHeight;
}

function sendNote() {
  const text = note.value.trim();
  if (!text) return;
  send("note " + text);
  note.value = "";
}

export function initChat() {
  document.getElementById("send").onclick = sendNote;
  // isComposing: Enter that commits an IME composition must not send.
  note.addEventListener("keydown", (ev) => {
    if (ev.key === "Enter" && !ev.isComposing) sendNote();
  });
}
