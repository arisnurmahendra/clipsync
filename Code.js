// Code.gs

const PROP_KEY = 'CLIP_VALUE';

// ── Normalisasi teks ──────────────────────────────────────────────
function normalizeText(text) {
  return text
    .replace(/[\r\n]+/g, ' ')  // newline → spasi
    .replace(/  +/g, ' ')      // collapse multiple space
    .trim();                    // trim awal & akhir
}

// ── Router utama ──────────────────────────────────────────────────
function doGet(e) {
  if (e.parameter.mode === 'api') {
    return handleApiGet();
  }
  return HtmlService
    .createHtmlOutputFromFile('index')
    .setTitle('Clipboard Relay')
    .addMetaTag('viewport', 'width=device-width, initial-scale=1')
    .setFaviconUrl('https://i.ibb.co.com/v64QWfz8/task-list.png')
    .setXFrameOptionsMode(HtmlService.XFrameOptionsMode.ALLOWALL);
}

function doPost(e) {
  const raw  = (e.parameter.text || '').toString();
  const text = normalizeText(raw);

  if (text === '') {
    return jsonResponse({ ok: false, error: 'empty text after normalization' });
  }

  PropertiesService.getScriptProperties().setProperty(PROP_KEY, text);
  return jsonResponse({ ok: true, stored: text });
}

// ── API handler ───────────────────────────────────────────────────
function handleApiGet() {
  const text = PropertiesService.getScriptProperties().getProperty(PROP_KEY) || '';
  return jsonResponse({ ok: true, text: text });
}

// ── Dipanggil dari HtmlService (client-side google.script.run) ────
function storeClip(text) {
  const normalized = normalizeText(text);
  if (normalized === '') {
    return { ok: false, error: 'empty text after normalization' };
  }
  PropertiesService.getScriptProperties().setProperty(PROP_KEY, normalized);
  return { ok: true, stored: normalized };
}

function clearClip() {
  PropertiesService.getScriptProperties().setProperty(PROP_KEY, '');
  return { ok: true };
}

function previewClip() {
  return PropertiesService.getScriptProperties().getProperty(PROP_KEY) || '';
}

// ── Helper ────────────────────────────────────────────────────────
function jsonResponse(obj) {
  return ContentService
    .createTextOutput(JSON.stringify(obj))
    .setMimeType(ContentService.MimeType.JSON);
}