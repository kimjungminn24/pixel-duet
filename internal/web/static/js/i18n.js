const STRINGS = {
  ko: {
    "lang.switch": "EN",
    "lang.title": "Switch to English",
    "transport.aria": "AI 재생 제어",
    "play": "AI 재생",
    "pause": "AI 정지",
    "connecting": "연결 중",
    "connecting.dots": "연결 중…",
    "connected": "캔버스 연결됨",
    "reconnecting": "연결 재시도 중",
    "state.playing": "진행중",
    "state.paused": "정지",
    "toolbox.aria": "그리기 도구",
    "tools": "도구",
    "tools.drawing": "그리기 도구",
    "tool.pen": "펜",
    "tool.pen.title": "펜 (B)",
    "tool.eraser": "지우개",
    "tool.eraser.title": "지우개 (E)",
    "tool.line": "직선",
    "tool.line.title": "직선 (L)",
    "tool.pick": "추출",
    "tool.pick.title": "색 추출 (I)",
    "tool.rect": "사각형",
    "tool.rect.title": "사각형 (R)",
    "tool.circle": "원",
    "tool.circle.title": "원 (O)",
    "tool.fill": "채우기",
    "tool.fill.title": "채우기 (F)",
    "tool.solidRect": "채운 사각형",
    "tool.solidRect.title": "채운 사각형 (U)",
    "tool.solidCircle": "채운 원",
    "tool.solidCircle.title": "채운 원 (P)",
    "palette": "팔레트",
    "palette.aria": "색상 팔레트",
    "palette.smaller": "팔레트 축소",
    "palette.larger": "팔레트 확대",
    "palette.edit": "편집",
    "palette.eraser": "지우개",
    "palette.color": "색상 {digit} {hex}",
    "workspace.aria": "캔버스 작업 공간",
    "zoom.aria": "캔버스 배율",
    "zoom.out": "축소",
    "zoom.out.aria": "캔버스 축소",
    "zoom.in": "확대",
    "zoom.in.aria": "캔버스 확대",
    "zoom.fit": "맞춤",
    "zoom.fit.title": "화면 맞춤",
    "pan.title": "화면 이동 (H / Space + 드래그)",
    "pan.aria": "화면 이동",
    "undo": "↶ 취소",
    "undo.title": "실행 취소 (Ctrl / ⌘ + Z)",
    "grid": "▦ 격자",
    "grid.title": "픽셀 격자 (G)",
    "canvas.aria": "픽셀 캔버스",
    "size": "캔버스 크기",
    "size.title": "크기를 바꾸면 새 캔버스를 시작합니다",
    "size.confirm": "{n} × {n} 크기의 새 캔버스를 시작할까요?",
    "clear": "초기화",
    "clear.confirm": "캔버스를 비우고 새로 그릴까요?",
    "save": "저장",
    "save.folder": "저장 폴더",
    "save.folder.placeholder": "기본 저장 폴더",
    "save.choose": "폴더 선택…",
    "save.choose.title": "저장 폴더 선택",
    "save.reset": "기본 저장 폴더로 변경",
    "save.name": "파일 이름",
    "save.path": "저장 경로",
    "save.unsupported": "폴더 선택은 Chrome 또는 Edge에서 지원됩니다.",
    "save.chooseFailed": "폴더 선택 실패: {message}",
    "save.overwrite": "{file} 파일을 덮어쓸까요?",
    "save.notConnected": "캔버스 연결을 확인하세요.",
    "save.written": "{path} 저장됨",
    "save.badName": "파일 이름은 영문 또는 숫자로 시작하고, 영문, 숫자, -, _ 만 쓸 수 있습니다 (최대 64자)",
    "save.failed": "저장 실패: {message}",
    "gallery.failed": "저장 폴더 조회 실패",
    "chat.aria": "AI와 대화",
    "chat": "대화",
    "chat.log": "대화 기록",
    "chat.placeholder": "메시지 입력",
    "chat.input": "AI에게 메시지",
    "chat.send": "메시지 보내기",
    "speed": "그리기 속도",
    "speed.aria": "그리기 속도",
    "speed.none": "최대 속도",
    "speed.per": "한 획마다 {s}초 쉬기",
    "speed.min": "빠르게",
    "speed.max": "천천히",
    "footer": "펜 B | 지우개 E | 직선 L | 사각형 R | 원 O | 채우기 F | 추출 I | 격자 G | 이동 H",
    "cmd.failed": "명령 실패: {message}",
    "cmd.badReply": "서버 응답 오류",
  },
  en: {
    "lang.switch": "한국어",
    "lang.title": "한국어로 전환",
    "transport.aria": "AI playback",
    "play": "Resume AI",
    "pause": "Pause AI",
    "connecting": "Connecting",
    "connecting.dots": "Connecting…",
    "connected": "Canvas connected",
    "reconnecting": "Reconnecting",
    "state.playing": "Playing",
    "state.paused": "Paused",
    "toolbox.aria": "Drawing tools",
    "tools": "Tools",
    "tools.drawing": "Drawing tools",
    "tool.pen": "Pen",
    "tool.pen.title": "Pen (B)",
    "tool.eraser": "Eraser",
    "tool.eraser.title": "Eraser (E)",
    "tool.line": "Line",
    "tool.line.title": "Line (L)",
    "tool.pick": "Pick",
    "tool.pick.title": "Color picker (I)",
    "tool.rect": "Rect",
    "tool.rect.title": "Rectangle (R)",
    "tool.circle": "Circle",
    "tool.circle.title": "Circle (O)",
    "tool.fill": "Fill",
    "tool.fill.title": "Fill (F)",
    "tool.solidRect": "Filled rect",
    "tool.solidRect.title": "Filled rectangle (U)",
    "tool.solidCircle": "Filled circle",
    "tool.solidCircle.title": "Filled circle (P)",
    "palette": "Palette",
    "palette.aria": "Color palette",
    "palette.smaller": "Smaller swatches",
    "palette.larger": "Larger swatches",
    "palette.edit": "Edit",
    "palette.eraser": "Eraser",
    "palette.color": "Color {digit} {hex}",
    "workspace.aria": "Canvas workspace",
    "zoom.aria": "Canvas zoom",
    "zoom.out": "Zoom out",
    "zoom.out.aria": "Zoom out",
    "zoom.in": "Zoom in",
    "zoom.in.aria": "Zoom in",
    "zoom.fit": "Fit",
    "zoom.fit.title": "Fit to window",
    "pan.title": "Pan (H / Space + drag)",
    "pan.aria": "Pan",
    "undo": "↶ Undo",
    "undo.title": "Undo (Ctrl / ⌘ + Z)",
    "grid": "▦ Grid",
    "grid.title": "Pixel grid (G)",
    "canvas.aria": "Pixel canvas",
    "size": "Canvas size",
    "size.title": "Changing the size starts a new canvas",
    "size.confirm": "Start a new {n} × {n} canvas?",
    "clear": "Clear",
    "clear.confirm": "Clear the canvas and start over?",
    "save": "Save",
    "save.folder": "Save folder",
    "save.folder.placeholder": "Default gallery folder",
    "save.choose": "Choose folder…",
    "save.choose.title": "Choose a save folder",
    "save.reset": "Back to the default folder",
    "save.name": "File name",
    "save.path": "Save path",
    "save.unsupported": "Choosing a folder needs Chrome or Edge.",
    "save.chooseFailed": "Could not choose a folder: {message}",
    "save.overwrite": "Overwrite {file}?",
    "save.notConnected": "The canvas is not connected.",
    "save.written": "Saved {path}",
    "save.badName": "File name: start with a letter or digit; letters, digits, - and _ only (64 max)",
    "save.failed": "Save failed: {message}",
    "gallery.failed": "Could not read the gallery folder",
    "chat.aria": "Chat with the AI",
    "chat": "Chat",
    "chat.log": "Conversation",
    "chat.placeholder": "Type a message",
    "chat.input": "Message to the AI",
    "chat.send": "Send",
    "speed": "Drawing speed",
    "speed.aria": "Drawing speed",
    "speed.none": "Full speed",
    "speed.per": "{s}s pause per stroke",
    "speed.min": "Fast",
    "speed.max": "Slow",
    "footer": "Pen B | Eraser E | Line L | Rect R | Circle O | Fill F | Pick I | Grid G | Pan H",
    "cmd.failed": "Command failed: {message}",
    "cmd.badReply": "Bad reply from the server",
  },
};

const STORAGE_KEY = "pixelduet-lang";
const listeners = [];

function storedLang() {
  try {
    return localStorage.getItem(STORAGE_KEY);
  } catch {
    return null;
  }
}

function browserLang() {
  const l = typeof navigator === "undefined" ? "" : navigator.language || "";
  return l.toLowerCase().startsWith("ko") ? "ko" : "en";
}

function requestedLang() {
  try {
    return new URLSearchParams(location.search).get("lang");
  } catch {
    return null;
  }
}

export let lang = STRINGS[requestedLang()] ? requestedLang()
  : STRINGS[storedLang()] ? storedLang() : browserLang();

export function t(key, vars = {}) {
  const s = STRINGS[lang][key] ?? STRINGS.ko[key] ?? key;
  return s.replace(/\{(\w+)\}/g, (_, name) => vars[name] ?? `{${name}}`);
}

export function applyTranslations(root = document) {
  document.documentElement.lang = lang;
  for (const el of root.querySelectorAll("[data-i18n]")) el.textContent = t(el.dataset.i18n);
  for (const el of root.querySelectorAll("[data-i18n-title]")) el.title = t(el.dataset.i18nTitle);
  for (const el of root.querySelectorAll("[data-i18n-aria]")) el.setAttribute("aria-label", t(el.dataset.i18nAria));
  for (const el of root.querySelectorAll("[data-i18n-placeholder]")) el.placeholder = t(el.dataset.i18nPlaceholder);
}

export function onLanguageChange(fn) {
  listeners.push(fn);
}

export function setLang(next) {
  if (!STRINGS[next] || next === lang) return;
  lang = next;
  try {
    localStorage.setItem(STORAGE_KEY, next);
  } catch { /* no storage */ }
  for (const fn of listeners) fn();
}

export function toggleLang() {
  setLang(lang === "ko" ? "en" : "ko");
}
