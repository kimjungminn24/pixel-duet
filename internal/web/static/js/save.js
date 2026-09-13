import { state, SPRITE_NAME } from "./state.js";
import { galleryPath, saveToGallery } from "./api.js";
import { showStatus } from "./status.js";
import { t } from "./i18n.js";

const folderField = document.getElementById("saveFolder");
const nameField = document.getElementById("saveName");
const saveBtn = document.getElementById("save");

let folderHandle = null; // FileSystemDirectoryHandle, per page session

async function chooseFolder() {
  if (!window.showDirectoryPicker) {
    showStatus(t("save.unsupported"));
    return;
  }
  try {
    folderHandle = await window.showDirectoryPicker({ id: "pixelduet-save", mode: "readwrite" });
    folderField.value = folderHandle.name;
    folderField.title = folderHandle.name;
    showStatus("");
  } catch (e) {
    if (e.name !== "AbortError") showStatus(t("save.chooseFailed", { message: e.message }), "orange");
  }
}

function resetFolder() {
  folderHandle = null;
  folderField.value = "";
  folderField.title = t("save.folder");
  showStatus("");
}

async function saveToFolder(name) {
  const fileName = name + ".txt";
  let exists = false;
  try {
    await folderHandle.getFileHandle(fileName);
    exists = true;
  } catch (e) {
    if (e.name !== "NotFoundError") throw e;
  }
  if (exists && !confirm(t("save.overwrite", { file: fileName }))) return null;
  if (!state.grid) throw new Error(t("save.notConnected"));
  const file = await folderHandle.getFileHandle(fileName, { create: true });
  const writer = await file.createWritable();
  try {
    await writer.write(state.grid);
    await writer.close();
  } catch (e) {
    await writer.abort().catch(() => {});
    throw e;
  }
  return t("save.written", { path: folderHandle.name + "/" + fileName });
}

async function save() {
  const name = nameField.value.trim();
  if (!SPRITE_NAME.test(name)) {
    showStatus(t("save.badName"), "orange");
    nameField.focus();
    return;
  }
  saveBtn.disabled = true;
  try {
    const report = folderHandle ? await saveToFolder(name) : await saveToGallery(name);
    if (report) showStatus(report, "green");
  } catch (e) {
    showStatus(t("save.failed", { message: e.message }), "orange");
  } finally {
    saveBtn.disabled = false;
  }
}

export function initSave() {
  document.getElementById("chooseFolder").onclick = chooseFolder;
  document.getElementById("resetFolder").onclick = resetFolder;
  saveBtn.onclick = save;
  galleryPath()
    .then((p) => { folderField.placeholder = p; })
    .catch((e) => showStatus(e.message, "orange"));
}
