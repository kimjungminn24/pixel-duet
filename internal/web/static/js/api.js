import { showStatus } from "./status.js";
import { t } from "./i18n.js";

// Keeps one stroke command under the server's 4 KB body limit.
const STROKE_CHUNK = 400;

let queue = Promise.resolve();

// Commands are serialized so they reach the hub in order.
export function send(line) {
  const request = queue.then(async () => {
    const res = await fetch("/cmd", { method: "POST", body: line });
    const text = await res.text();
    if (!res.ok || text.startsWith("err ")) throw new Error(text || t("cmd.badReply"));
    return text;
  });
  queue = request.catch((err) => showStatus(t("cmd.failed", { message: err.message }), "orange"));
  return queue;
}

export function sendStroke(color, pts) {
  for (let i = 0; i < pts.length; i += STROKE_CHUNK) {
    const part = pts.slice(i, i + STROKE_CHUNK).map(([x, y]) => x + "," + y).join(" ");
    send(`stroke ${color} ${part}`);
  }
}

export function connect(handlers) {
  const es = new EventSource("/events");
  es.onmessage = (ev) => {
    const m = JSON.parse(ev.data);
    handlers[m.type]?.(m);
  };
  es.onopen = handlers.open;
  es.onerror = handlers.error;
  return es;
}

export async function galleryPath() {
  const res = await fetch("/gallery");
  if (!res.ok) throw new Error(t("gallery.failed"));
  return res.text();
}

export async function saveToGallery(name) {
  const res = await fetch("/save", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ name, folder: "" }),
  });
  const text = await res.text();
  if (!res.ok) throw new Error(text);
  return text;
}
