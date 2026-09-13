const el = document.getElementById("saved");

// tone: "green", "orange" or "" for plain text.
export function showStatus(text, tone = "") {
  el.textContent = text;
  el.style.color = tone ? `var(--${tone})` : "";
}
