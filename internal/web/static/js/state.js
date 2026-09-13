export const DIGITS = "0123456789abcdefghijklmnopqrstuvw";

export const CHECKER = ["#e5e2e8", "#cbc6d1"];

export const SPRITE_NAME = /^[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}$/;

export const DEFAULT_PALETTE = [null,
  "#000000", "#222034", "#45283c", "#663931", "#8f563b", "#df7126", "#d9a066", "#eec39a",
  "#fbf236", "#99e550", "#6abe30", "#37946e", "#4b692f", "#524b24", "#323c39", "#3f3f74",
  "#306082", "#5b6ee1", "#639bff", "#5fcde4", "#cbdbfc", "#ffffff", "#9badb7", "#847e87",
  "#696a6a", "#595652", "#76428a", "#ac3232", "#d95763", "#d77bba", "#8f974a", "#8a6f30",
];

export const state = {
  w: 32,
  h: 32,
  grid: "",                        // digit rows as sent by the server
  cells: new Uint8Array(32 * 32),  // palette index per pixel, row-major
  paused: false,
  palette: [...DEFAULT_PALETTE],
  color: 9,
  tool: "pen",
  gridOn: false,
  preview: null,                   // {color, pts} while dragging a shape
};

export function setGrid(w, h, grid) {
  const cells = new Uint8Array(w * h);
  const rows = grid.split("\n");
  for (let y = 0; y < h; y++) {
    const row = rows[y] || "";
    for (let x = 0; x < w; x++) {
      cells[y * w + x] = Math.max(0, DIGITS.indexOf(row[x] || "0"));
    }
  }
  state.w = w;
  state.h = h;
  state.grid = grid;
  state.cells = cells;
}

export function inCanvas(x, y) {
  return x >= 0 && x < state.w && y >= 0 && y < state.h;
}

export function digitAt(x, y) {
  return inCanvas(x, y) ? state.cells[y * state.w + x] : 0;
}
