package generator

// appTemplate is the single-page shell. All content is rendered client-side by
// app.js from the JSON model embedded in #unifidoc-data.
const appTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=IBM+Plex+Sans:wght@400;500;600;700&family=IBM+Plex+Mono:wght@400;500;600&display=swap" rel="stylesheet">
    <link rel="stylesheet" href="assets/style.css">
</head>
<body class="theme-light">
    <div id="app"></div>
    <noscript>This documentation requires JavaScript to render.</noscript>
    <script type="application/json" id="unifidoc-data">{{ .Data }}</script>
    <script src="assets/app.js"></script>
</body>
</html>`

const appCSS = `
:root, .theme-light {
  --bg:#fcfcfd; --panel:#ffffff; --border:#e8e8ee; --border-soft:#f0f0f4;
  --text:#1a1a22; --muted:#6d6d7a; --faint:#9a9aa6;
  --accent:oklch(0.52 0.15 155); --accent-strong:oklch(0.44 0.15 155);
  --accent-soft:oklch(0.95 0.05 155); --hover:#f4f4f8;
  --code-bg:#111118; --code-panel:#16161e; --code-border:#26262f;
}
.theme-dark {
  --bg:#101014; --panel:#16161c; --border:#26262e; --border-soft:#1e1e26;
  --text:#ececf1; --muted:#93939f; --faint:#6a6a76;
  --accent:oklch(0.72 0.14 155); --accent-strong:oklch(0.8 0.12 155);
  --accent-soft:oklch(0.28 0.07 155); --hover:#1d1d25;
  --code-bg:#0c0c11; --code-panel:#111118; --code-border:#22222b;
}

* { margin:0; padding:0; box-sizing:border-box; }
html, body { margin:0; padding:0; }
body {
  font-family:'IBM Plex Sans', -apple-system, BlinkMacSystemFont, sans-serif;
  -webkit-font-smoothing:antialiased;
  background:var(--bg); color:var(--text);
  font-size:14px; line-height:1.55;
}
.mono { font-family:'IBM Plex Mono', ui-monospace, monospace; }
a { color:var(--accent); text-decoration:none; }
a:hover { color:var(--accent-strong); text-decoration:underline; }
::selection { background:var(--accent-soft); }
button { font-family:inherit; }

/* ---------- TOP BAR ---------- */
.topbar {
  position:sticky; top:0; z-index:40; display:flex; align-items:center; gap:20px;
  height:56px; padding:0 20px; background:var(--panel); border-bottom:1px solid var(--border);
}
.brand { display:flex; align-items:center; gap:10px; width:240px; flex-shrink:0; }
.brand-mark {
  width:26px; height:26px; border-radius:7px; background:var(--accent);
  display:grid; place-items:center; color:#fff;
  font-family:'IBM Plex Mono',monospace; font-weight:600; font-size:13px;
}
.brand-name { font-weight:600; font-size:15px; letter-spacing:-0.01em; }
.brand-ver {
  font-family:'IBM Plex Mono',monospace; font-size:11px; color:var(--muted);
  border:1px solid var(--border); border-radius:5px; padding:1px 6px;
}
.search-trigger {
  flex:1; max-width:420px; display:flex; align-items:center; gap:8px; height:34px;
  padding:0 12px; background:var(--bg); border:1px solid var(--border); border-radius:8px;
  color:var(--faint); font-size:13px; cursor:pointer; text-align:left;
}
.search-trigger:hover { border-color:var(--accent); }
.search-trigger .grow { flex:1; }
.kbd {
  font-family:'IBM Plex Mono',monospace; font-size:11px; border:1px solid var(--border);
  border-radius:4px; padding:1px 5px; color:var(--muted);
}
.spacer { flex:1; }
.topnav { display:flex; align-items:center; gap:18px; font-size:13.5px; color:var(--muted); }
.topnav a { color:var(--muted); }
.topnav a.active { color:var(--text); font-weight:500; }
.theme-btn {
  width:32px; height:32px; border-radius:8px; border:1px solid var(--border);
  background:var(--panel); color:var(--muted); cursor:pointer; font-size:14px;
}
.theme-btn:hover { background:var(--hover); }

/* ---------- LAYOUT ---------- */
.layout { display:grid; grid-template-columns:260px minmax(0,1fr) 420px; align-items:start; }

/* ---------- SIDEBAR ---------- */
.sidebar {
  position:sticky; top:56px; height:calc(100vh - 56px); overflow-y:auto;
  border-right:1px solid var(--border); padding:16px 12px 40px; background:var(--panel);
}
.group-tabs {
  display:flex; background:var(--bg); border:1px solid var(--border); border-radius:8px;
  padding:3px; margin:0 4px 16px; gap:2px;
}
.group-tab {
  flex:1; height:26px; border:none; border-radius:6px; font-size:12px; font-weight:500;
  cursor:pointer; background:transparent; color:var(--muted);
}
.group-tab.active { background:var(--panel); color:var(--text); box-shadow:0 1px 2px rgba(0,0,0,0.08); }
.nav-section { margin-bottom:4px; }
.nav-sec-btn {
  display:flex; align-items:center; gap:8px; width:100%; padding:6px 8px; border:none;
  background:transparent; cursor:pointer; border-radius:6px; color:var(--text);
}
.nav-sec-btn:hover { background:var(--hover); }
.nav-dot { width:8px; height:8px; border-radius:3px; flex-shrink:0; }
.nav-sec-title { font-size:13px; font-weight:600; flex:1; text-align:left; }
.nav-sec-tag { font-family:'IBM Plex Mono',monospace; font-size:10px; color:var(--faint); }
.nav-chevron { font-size:9px; color:var(--faint); transition:transform .15s; }
.nav-items { display:flex; flex-direction:column; gap:1px; padding:2px 0 6px 16px; }
.nav-item {
  display:flex; align-items:center; gap:8px; padding:5px 8px; border:none; text-align:left;
  cursor:pointer; font-size:13px; border-radius:6px; background:transparent;
  color:var(--muted); font-weight:400; width:100%;
}
.nav-item:hover { background:var(--hover); }
.nav-item.active { background:var(--accent-soft); color:var(--accent-strong); font-weight:600; }
.nav-item .label { overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.badge {
  font-family:'IBM Plex Mono',monospace; font-size:9.5px; font-weight:600;
  letter-spacing:0.02em; width:38px; flex-shrink:0;
}

/* ---------- CONTENT ---------- */
.content { padding:36px 44px 96px; max-width:760px; }
.breadcrumb {
  display:flex; align-items:center; gap:8px; font-size:12.5px; color:var(--faint);
  margin-bottom:14px; font-family:'IBM Plex Mono',monospace; flex-wrap:wrap;
}
.breadcrumb .cur { color:var(--muted); }
.title { margin:0 0 10px; font-size:28px; font-weight:700; letter-spacing:-0.02em; line-height:1.2; }
.title .dep { font-size:13px; font-weight:600; color:oklch(0.62 0.14 35); margin-left:10px; vertical-align:middle; }
.method-pill {
  display:inline-flex; align-items:center; gap:10px; margin:4px 0 20px; padding:7px 12px;
  background:var(--panel); border:1px solid var(--border); border-radius:8px;
  font-family:'IBM Plex Mono',monospace; font-size:13px;
}
.method-tag {
  font-weight:600; font-size:11px; letter-spacing:0.04em; color:#fff; border-radius:5px; padding:2px 7px;
}
.method-path { color:var(--muted); }
.desc { margin:0 0 28px; color:var(--muted); font-size:15px; max-width:620px; text-wrap:pretty; }
.auth-banner {
  display:flex; align-items:center; gap:10px; flex-wrap:wrap; padding:12px 14px;
  border:1px solid var(--border); border-radius:10px; background:var(--panel);
  margin-bottom:36px; font-size:13px; color:var(--muted);
}
.auth-dot { width:7px; height:7px; border-radius:50%; background:oklch(0.58 0.14 155); }
.auth-chip {
  font-family:'IBM Plex Mono',monospace; font-size:12px; background:var(--accent-soft);
  color:var(--accent-strong); border-radius:5px; padding:2px 7px;
}
.section-h { margin:0 0 4px; font-size:16px; font-weight:600; letter-spacing:-0.01em; }
.section-h.gap { margin-top:40px; }
.divider { height:1px; background:var(--border); margin:12px 0 4px; }
.divider.wide { margin:12px 0 16px; }
.param { padding:16px 0; border-bottom:1px solid var(--border-soft); }
.param-head { display:flex; align-items:baseline; gap:10px; margin-bottom:5px; flex-wrap:wrap; }
.param-name { font-family:'IBM Plex Mono',monospace; font-size:13.5px; font-weight:600; }
.param-type { font-family:'IBM Plex Mono',monospace; font-size:12px; color:var(--accent); }
.param-in { font-family:'IBM Plex Mono',monospace; font-size:11px; color:var(--faint); }
.param-req { font-family:'IBM Plex Mono',monospace; font-size:11px; }
.param-req.required { color:oklch(0.6 0.15 25); }
.param-req.optional { color:var(--faint); }
.param-desc { margin:0; color:var(--muted); font-size:13.5px; max-width:600px; }
.returns-prose { margin:0 0 20px; color:var(--muted); font-size:13.5px; max-width:620px; }
.error-list { display:flex; flex-direction:column; gap:8px; margin-bottom:44px; }
.error-card {
  display:flex; align-items:baseline; gap:14px; padding:10px 14px; border:1px solid var(--border);
  border-radius:8px; background:var(--panel); font-size:13px;
}
.error-code { font-family:'IBM Plex Mono',monospace; font-weight:600; font-size:12.5px; width:32px; }
.error-name { font-family:'IBM Plex Mono',monospace; font-size:12.5px; color:var(--text); }
.error-desc { color:var(--muted); }
.prevnext { display:flex; justify-content:space-between; margin-top:48px; gap:12px; }
.navcard {
  flex:1; padding:14px 16px; border:1px solid var(--border); border-radius:10px;
  color:var(--text); background:var(--panel);
}
.navcard:hover { border-color:var(--accent); text-decoration:none; }
.navcard.next { text-align:right; }
.navcard .dir { font-size:11px; color:var(--faint); margin-bottom:2px; }
.navcard .lbl { font-size:13.5px; font-weight:500; }
.empty-state { padding:80px 0; text-align:center; color:var(--faint); }

/* ---------- CODE PANEL ---------- */
.codepanel {
  position:sticky; top:56px; height:calc(100vh - 56px); overflow-y:auto;
  background:var(--code-bg); padding:24px 24px 48px; display:flex; flex-direction:column; gap:20px;
}
.code-card {
  border:1px solid var(--code-border); border-radius:12px; background:var(--code-panel);
  overflow:hidden; flex-shrink:0;
}
.code-tabs { display:flex; align-items:center; gap:2px; padding:8px 10px; border-bottom:1px solid var(--code-border); }
.lang-tab {
  height:26px; padding:0 12px; border:none; border-radius:6px; font-family:'IBM Plex Mono',monospace;
  font-size:12px; cursor:pointer; background:transparent; color:#8b8b99;
}
.lang-tab:hover { color:#e6e6ef; }
.lang-tab.active { background:#26262f; color:#e6e6ef; }
.copy-btn {
  height:26px; padding:0 10px; border:1px solid var(--code-border); border-radius:6px;
  background:transparent; color:#8b8b99; font-family:'IBM Plex Mono',monospace; font-size:11px; cursor:pointer;
}
.copy-btn:hover { color:#e6e6ef; }
pre.code {
  margin:0; padding:16px 18px; font-family:'IBM Plex Mono',monospace; font-size:12.5px;
  line-height:1.7; overflow-x:auto; color:#c9c9d6; white-space:pre;
}
.resp-head { display:flex; align-items:center; gap:8px; padding:9px 14px; border-bottom:1px solid var(--code-border); }
.resp-label { font-family:'IBM Plex Mono',monospace; font-size:11px; color:#8b8b99; letter-spacing:0.05em; }
.resp-status { font-family:'IBM Plex Mono',monospace; font-size:11px; color:oklch(0.72 0.14 155); }

/* ---------- SEARCH MODAL ---------- */
.modal-overlay {
  position:fixed; inset:0; z-index:100; background:rgba(10,10,16,0.5); backdrop-filter:blur(3px);
  display:flex; justify-content:center; padding-top:12vh;
}
.modal {
  width:600px; max-width:92vw; max-height:420px; background:var(--panel); border:1px solid var(--border);
  border-radius:14px; box-shadow:0 24px 64px rgba(0,0,0,0.3); overflow:hidden;
  display:flex; flex-direction:column; align-self:flex-start;
}
.modal-search { display:flex; align-items:center; gap:10px; padding:14px 16px; border-bottom:1px solid var(--border); }
.modal-input { flex:1; border:none; outline:none; background:transparent; font-family:inherit; font-size:15px; color:var(--text); }
.modal-results { overflow-y:auto; padding:8px; }
.result-item {
  display:flex; align-items:center; gap:10px; width:100%; padding:9px 10px; border:none;
  background:transparent; cursor:pointer; text-align:left; border-radius:8px; color:var(--text);
}
.result-item:hover, .result-item.active { background:var(--hover); }
.result-dot { width:7px; height:7px; border-radius:2.5px; flex-shrink:0; }
.result-badge { font-family:'IBM Plex Mono',monospace; font-size:9.5px; font-weight:600; width:44px; flex-shrink:0; }
.result-label { font-size:13.5px; flex:1; }
.result-section { font-size:11.5px; color:var(--faint); }
.modal-empty { padding:24px; text-align:center; color:var(--faint); font-size:13px; }
.modal-foot {
  display:flex; gap:14px; padding:9px 16px; border-top:1px solid var(--border);
  font-size:11px; color:var(--faint); font-family:'IBM Plex Mono',monospace;
}

/* ---------- RESPONSIVE ---------- */
@media (max-width:1200px) {
  .layout { grid-template-columns:240px minmax(0,1fr); }
  .codepanel { grid-column:1 / -1; position:static; height:auto; }
}
@media (max-width:820px) {
  .layout { grid-template-columns:1fr; }
  .sidebar { display:none; }
  .brand { width:auto; }
  .topnav { display:none; }
  .content { padding:28px 20px 80px; }
}
`

const appJS = `
(function () {
  "use strict";
  var root = document.getElementById('app');
  var dataEl = document.getElementById('unifidoc-data');
  var data = {};
  try { data = JSON.parse(dataEl.textContent); } catch (e) { data = { project:{}, protocols:[] }; }

  var PROTOCOL_META = {
    openapi:   { label:'REST API',  tag:'openapi', color:'oklch(0.62 0.14 155)' },
    api:       { label:'REST API',  tag:'api',     color:'oklch(0.62 0.14 155)' },
    grpc:      { label:'gRPC',      tag:'proto',   color:'oklch(0.62 0.14 220)' },
    websocket: { label:'WebSocket', tag:'ws',      color:'oklch(0.62 0.14 268)' },
    webhook:   { label:'Webhooks',  tag:'yaml',    color:'oklch(0.62 0.14 35)' },
    asyncapi:  { label:'AsyncAPI',  tag:'async',   color:'oklch(0.62 0.14 90)' },
    events:    { label:'Events',    tag:'events',  color:'oklch(0.62 0.14 320)' },
    mcp:       { label:'MCP',       tag:'mcp',     color:'oklch(0.62 0.14 190)' },
    custom:    { label:'Custom',    tag:'custom',  color:'oklch(0.62 0.14 268)' }
  };
  var LANG_ORDER = ['curl', 'grpcurl', 'wscat', 'javascript', 'python', 'go', 'json', 'payload'];
  var LANG_LABEL = { curl:'curl', grpcurl:'grpcurl', wscat:'wscat', javascript:'js', python:'python', go:'go', json:'json-rpc', payload:'payload' };

  function meta(id) { return PROTOCOL_META[id] || { label:id, tag:id, color:'oklch(0.6 0.1 268)' }; }

  function esc(s) {
    return String(s == null ? '' : s)
      .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;');
  }

  function badgeFor(method) {
    var m = (method || '').toUpperCase();
    var map = { DELETE:'DEL', RECEIVE:'RECV', BIDIRECTIONAL:'BIDI', RESOURCE:'RES', EVENT:'EVT',
      PUBLISH:'PUB', SUBSCRIBE:'SUB' };
    if (map[m]) return map[m];
    if (m.length > 5) return m.slice(0, 4);
    return m || 'GET';
  }

  function methodColor(method) {
    var m = (method || '').toUpperCase();
    if (m === 'GET') return 'oklch(0.62 0.12 220)';
    if (m === 'POST') return 'oklch(0.58 0.14 155)';
    if (m === 'PUT' || m === 'PATCH') return 'oklch(0.62 0.14 90)';
    if (m === 'DELETE') return 'oklch(0.6 0.15 25)';
    if (m === 'SEND' || m === 'RECEIVE' || m === 'BIDIRECTIONAL') return 'oklch(0.65 0.1 268)';
    if (m === 'RPC' || m === 'MSG') return 'oklch(0.62 0.12 220)';
    if (m === 'EVENT') return 'oklch(0.62 0.12 320)';
    if (m === 'TOOL' || m === 'RESOURCE' || m === 'PROMPT') return 'oklch(0.62 0.12 190)';
    if (m === 'PUB' || m === 'SUB' || m === 'PUBLISH' || m === 'SUBSCRIBE') return 'oklch(0.62 0.11 90)';
    return 'oklch(0.55 0.02 268)';
  }

  function statusColor(status) {
    var c = (status || '')[0];
    if (c === '4') return 'oklch(0.62 0.14 90)';
    if (c === '5') return 'oklch(0.6 0.15 25)';
    return 'oklch(0.62 0.14 155)';
  }

  // Flatten resources with a back-reference to their protocol.
  var flat = [];
  data.protocols.forEach(function (p) {
    (p.resources || []).forEach(function (r) { flat.push({ p: p, r: r }); });
  });

  var state = {
    group: 'protocol',
    lang: null,
    active: null,
    closed: {},
    searchOpen: false,
    query: '',
    searchIndex: 0
  };

  function byId(id) {
    for (var i = 0; i < flat.length; i++) { if (flat[i].r.id === id) return flat[i]; }
    return null;
  }

  function sections() {
    if (state.group === 'protocol') {
      return data.protocols.map(function (p) {
        var m = meta(p.id);
        return {
          key: p.id,
          title: p.title || m.label,
          tag: m.tag,
          color: m.color,
          items: (p.resources || []).map(function (r) { return { p: p, r: r }; })
        };
      });
    }
    // Group by first tag across all protocols.
    var order = [];
    var groups = {};
    flat.forEach(function (e) {
      var t = (e.r.tags && e.r.tags[0]) || 'General';
      if (!groups[t]) { groups[t] = []; order.push(t); }
      groups[t].push(e);
    });
    var palette = ['oklch(0.62 0.14 155)', 'oklch(0.62 0.14 220)', 'oklch(0.62 0.14 90)',
      'oklch(0.62 0.14 320)', 'oklch(0.62 0.14 35)', 'oklch(0.62 0.14 190)', 'oklch(0.62 0.14 268)'];
    return order.map(function (t, i) {
      return { key: t, title: t, tag: String(groups[t].length), color: palette[i % palette.length], items: groups[t] };
    });
  }

  function availableLangs(r) {
    var code = r.code || {};
    var langs = [];
    LANG_ORDER.forEach(function (l) { if (code[l]) langs.push(l); });
    Object.keys(code).forEach(function (l) { if (langs.indexOf(l) === -1) langs.push(l); });
    return langs;
  }

  function currentEntry() { return byId(state.active) || flat[0] || null; }

  // ---------- RENDER ----------
  function render() {
    if (!flat.length) { root.innerHTML = topbar() + '<div class="empty-state">No protocols configured.</div>'; wire(); return; }
    var entry = currentEntry();
    state.active = entry.r.id;
    var langs = availableLangs(entry.r);
    if (langs.indexOf(state.lang) === -1) state.lang = langs[0] || null;

    root.innerHTML =
      topbar() +
      '<div class="layout">' +
        sidebar() +
        content(entry) +
        codepanel(entry, langs) +
      '</div>' +
      (state.searchOpen ? modal() : '');
    wire();
  }

  function topbar() {
    var p = data.project || {};
    var ver = p.version ? '<span class="brand-ver">v' + esc(p.version) + '</span>' : '';
    return '' +
      '<header class="topbar">' +
        '<div class="brand">' +
          '<div class="brand-mark">' + esc((p.name || 'u').trim().charAt(0).toLowerCase() || 'u') + '</div>' +
          '<span class="brand-name">' + esc(p.name || 'unifidoc') + '</span>' + ver +
        '</div>' +
        '<button class="search-trigger" data-act="search">' +
          '<span>⌕</span><span class="grow">Search endpoints, events, types…</span>' +
          '<span class="kbd">⌘K</span>' +
        '</button>' +
        '<div class="spacer"></div>' +
        '<nav class="topnav"><a class="active">API Reference</a></nav>' +
        '<button class="theme-btn" data-act="theme" title="Toggle theme">' + (isDark() ? '☀' : '☾') + '</button>' +
      '</header>';
  }

  function sidebar() {
    var html = '<aside class="sidebar">' +
      '<div class="group-tabs">' +
        '<button class="group-tab ' + (state.group === 'protocol' ? 'active' : '') + '" data-group="protocol">By protocol</button>' +
        '<button class="group-tab ' + (state.group === 'resource' ? 'active' : '') + '" data-group="resource">By resource</button>' +
      '</div>';
    sections().forEach(function (sec) {
      var open = !state.closed[sec.key];
      html += '<div class="nav-section">' +
        '<button class="nav-sec-btn" data-toggle="' + esc(sec.key) + '">' +
          '<span class="nav-dot" style="background:' + sec.color + '"></span>' +
          '<span class="nav-sec-title">' + esc(sec.title) + '</span>' +
          '<span class="nav-sec-tag">' + esc(sec.tag) + '</span>' +
          '<span class="nav-chevron" style="transform:rotate(' + (open ? 90 : 0) + 'deg)">▶</span>' +
        '</button>';
      if (open) {
        html += '<div class="nav-items">';
        sec.items.forEach(function (e) {
          var active = e.r.id === state.active;
          html += '<button class="nav-item ' + (active ? 'active' : '') + '" data-goto="' + esc(e.r.id) + '">' +
            '<span class="badge" style="color:' + methodColor(e.r.method) + '">' + esc(badgeFor(e.r.method)) + '</span>' +
            '<span class="label">' + esc(e.r.name || e.r.path) + '</span>' +
          '</button>';
        });
        html += '</div>';
      }
      html += '</div>';
    });
    return html + '</aside>';
  }

  function content(entry) {
    var r = entry.r, p = entry.p, m = meta(p.id);
    var crumbTag = (r.tags && r.tags[0]) ? '<span>' + esc(r.tags[0]) + '</span><span>/</span>' : '';
    var html = '<main class="content">' +
      '<div class="breadcrumb"><span>' + esc(m.label) + '</span><span>/</span>' + crumbTag +
        '<span class="cur">' + esc(r.name || r.path) + '</span></div>' +
      '<h1 class="title">' + esc(r.name || r.path) +
        (r.deprecated ? '<span class="dep">deprecated</span>' : '') + '</h1>';

    if (r.method || r.path) {
      html += '<div class="method-pill">' +
        (r.method ? '<span class="method-tag" style="background:' + methodColor(r.method) + '">' + esc(r.method.toUpperCase()) + '</span>' : '') +
        (r.path ? '<span class="method-path">' + esc(r.path) + '</span>' : '') +
      '</div>';
    }
    if (r.description) html += '<p class="desc">' + esc(r.description) + '</p>';

    html += authBanner(r);
    html += paramSection('Body parameters', r.body);
    html += paramSection('Parameters', r.params);
    html += returnsSection(r);
    html += prevNext(entry);
    return html + '</main>';
  }

  function authBanner(r) {
    if (!r.security || !r.security.length) return '';
    var chips = r.security.map(function (a) {
      var s = '<span class="auth-chip">' + esc(a.name) + '</span>';
      if (a.scopes && a.scopes.length) {
        s += '<span>scope</span>' + a.scopes.map(function (sc) { return '<span class="auth-chip">' + esc(sc) + '</span>'; }).join('');
      }
      return s;
    }).join('');
    return '<div class="auth-banner"><span class="auth-dot"></span><span>Requires</span>' + chips + '</div>';
  }

  function paramSection(title, params) {
    if (!params || !params.length) return '';
    var html = '<h2 class="section-h">' + esc(title) + '</h2><div class="divider"></div>';
    params.forEach(function (pr) {
      var reqCls = pr.required ? 'required' : 'optional';
      var reqTxt = pr.required ? 'required' : 'optional';
      html += '<div class="param">' +
        '<div class="param-head">' +
          '<span class="param-name">' + esc(pr.name) + '</span>' +
          (pr.type ? '<span class="param-type">' + esc(pr.type) + '</span>' : '') +
          (pr.in ? '<span class="param-in">' + esc(pr.in) + '</span>' : '') +
          '<span class="param-req ' + reqCls + '">' + reqTxt + '</span>' +
        '</div>' +
        (pr.description ? '<p class="param-desc">' + esc(pr.description) + '</p>' : '') +
      '</div>';
    });
    return html;
  }

  function returnsSection(r) {
    if (!r.responses || !r.responses.length) return '';
    var success = null, errors = [];
    r.responses.forEach(function (resp) {
      var c = (resp.status || '')[0];
      if (c === '2' || c === '3') { if (!success) success = resp; }
      else if (c === '4' || c === '5') errors.push(resp);
    });
    var html = '<h2 class="section-h gap">Returns</h2><div class="divider wide"></div>';
    if (success) {
      html += '<p class="returns-prose">Returns <span class="mono">' + esc(success.status) + '</span> — ' + esc(success.description || 'Success.') + '</p>';
    }
    if (errors.length) {
      html += '<div class="error-list">';
      errors.forEach(function (e) {
        html += '<div class="error-card">' +
          '<span class="error-code" style="color:' + statusColor(e.status) + '">' + esc(e.status) + '</span>' +
          '<span class="error-desc">' + esc(e.description || '') + '</span>' +
        '</div>';
      });
      html += '</div>';
    }
    return html;
  }

  function prevNext(entry) {
    var idx = flat.indexOf(entry);
    var prev = idx > 0 ? flat[idx - 1] : null;
    var next = idx >= 0 && idx < flat.length - 1 ? flat[idx + 1] : null;
    if (!prev && !next) return '';
    var html = '<div class="prevnext">';
    html += prev
      ? '<a class="navcard" data-goto="' + esc(prev.r.id) + '" href="#' + esc(prev.r.id) + '"><div class="dir">← Previous</div><div class="lbl">' + esc(prev.r.name || prev.r.path) + '</div></a>'
      : '<span></span>';
    html += next
      ? '<a class="navcard next" data-goto="' + esc(next.r.id) + '" href="#' + esc(next.r.id) + '"><div class="dir">Next →</div><div class="lbl">' + esc(next.r.name || next.r.path) + '</div></a>'
      : '<span></span>';
    return html + '</div>';
  }

  function codepanel(entry, langs) {
    var r = entry.r;
    var code = (r.code && state.lang && r.code[state.lang]) || '';
    var tabs = langs.map(function (l) {
      return '<button class="lang-tab ' + (l === state.lang ? 'active' : '') + '" data-lang="' + esc(l) + '">' + esc(LANG_LABEL[l] || l) + '</button>';
    }).join('');
    var reqCard = langs.length ? '' +
      '<div class="code-card">' +
        '<div class="code-tabs">' + tabs + '<div class="grow" style="flex:1"></div>' +
          '<button class="copy-btn" data-act="copy">copy</button>' +
        '</div>' +
        '<pre class="code" id="req-code">' + esc(code) + '</pre>' +
      '</div>' : '';
    var respCard = r.responseExample ? '' +
      '<div class="code-card">' +
        '<div class="resp-head"><span class="resp-label">RESPONSE</span>' +
          (r.responseStatus ? '<span class="resp-status">' + esc(r.responseStatus) + '</span>' : '') +
        '</div>' +
        '<pre class="code">' + esc(r.responseExample) + '</pre>' +
      '</div>' : '';
    return '<aside class="codepanel">' + reqCard + respCard + '</aside>';
  }

  function modal() {
    var results = searchResults();
    var body;
    if (state.query && !results.length) {
      body = '<div class="modal-empty">No results for “' + esc(state.query) + '”</div>';
    } else {
      body = results.map(function (e, i) {
        var m = meta(e.p.id);
        return '<button class="result-item ' + (i === state.searchIndex ? 'active' : '') + '" data-goto="' + esc(e.r.id) + '">' +
          '<span class="result-dot" style="background:' + m.color + '"></span>' +
          '<span class="result-badge" style="color:' + methodColor(e.r.method) + '">' + esc(badgeFor(e.r.method)) + '</span>' +
          '<span class="result-label">' + esc(e.r.name || e.r.path) + '</span>' +
          '<span class="result-section">' + esc(m.label) + '</span>' +
        '</button>';
      }).join('');
    }
    return '<div class="modal-overlay" data-act="overlay">' +
      '<div class="modal" data-stop="1">' +
        '<div class="modal-search"><span style="color:var(--faint)">⌕</span>' +
          '<input class="modal-input" id="search-input" placeholder="Search across all protocols…" value="' + esc(state.query) + '">' +
          '<span class="kbd">ESC</span>' +
        '</div>' +
        '<div class="modal-results">' + body + '</div>' +
        '<div class="modal-foot"><span>↑↓ navigate</span><span>↵ open</span><span>esc close</span></div>' +
      '</div>' +
    '</div>';
  }

  function searchResults() {
    var q = state.query.trim().toLowerCase();
    var list = flat;
    if (q) {
      list = flat.filter(function (e) {
        var hay = (e.r.name + ' ' + e.r.method + ' ' + e.r.path + ' ' + meta(e.p.id).label + ' ' + (e.r.tags || []).join(' ')).toLowerCase();
        return hay.indexOf(q) !== -1;
      });
    }
    return list.slice(0, 8);
  }

  // ---------- INTERACTION ----------
  function goto(id) {
    state.active = id;
    if (history.replaceState) history.replaceState(null, '', '#' + id);
    else location.hash = id;
    state.searchOpen = false;
    render();
    var c = document.querySelector('.content');
    if (c) c.scrollIntoView({ block: 'start' });
  }

  function wire() {
    root.querySelectorAll('[data-goto]').forEach(function (el) {
      el.addEventListener('click', function (ev) { ev.preventDefault(); goto(el.getAttribute('data-goto')); });
    });
    root.querySelectorAll('[data-group]').forEach(function (el) {
      el.addEventListener('click', function () { state.group = el.getAttribute('data-group'); render(); });
    });
    root.querySelectorAll('[data-toggle]').forEach(function (el) {
      el.addEventListener('click', function () {
        var k = el.getAttribute('data-toggle');
        state.closed[k] = !state.closed[k];
        render();
      });
    });
    root.querySelectorAll('[data-lang]').forEach(function (el) {
      el.addEventListener('click', function () { state.lang = el.getAttribute('data-lang'); render(); });
    });
    var searchBtn = root.querySelector('[data-act="search"]');
    if (searchBtn) searchBtn.addEventListener('click', openSearch);
    var themeBtn = root.querySelector('[data-act="theme"]');
    if (themeBtn) themeBtn.addEventListener('click', toggleTheme);
    var copyBtn = root.querySelector('[data-act="copy"]');
    if (copyBtn) copyBtn.addEventListener('click', function () {
      var pre = document.getElementById('req-code');
      if (pre && navigator.clipboard) {
        navigator.clipboard.writeText(pre.textContent);
        copyBtn.textContent = 'copied ✓';
        setTimeout(function () { copyBtn.textContent = 'copy'; }, 1500);
      }
    });
    var overlay = root.querySelector('[data-act="overlay"]');
    if (overlay) overlay.addEventListener('click', function (ev) {
      if (ev.target === overlay) closeSearch();
    });
    var input = document.getElementById('search-input');
    if (input) {
      input.focus();
      input.addEventListener('input', function () { state.query = input.value; state.searchIndex = 0; render(); });
    }
  }

  function openSearch() { state.searchOpen = true; state.query = ''; state.searchIndex = 0; render(); }
  function closeSearch() { state.searchOpen = false; render(); }

  // ---------- THEME ----------
  function isDark() { return document.body.classList.contains('theme-dark'); }
  function applyTheme(dark) {
    document.body.classList.toggle('theme-dark', dark);
    document.body.classList.toggle('theme-light', !dark);
    try { localStorage.setItem('unifidoc-theme', dark ? 'dark' : 'light'); } catch (e) {}
  }
  function toggleTheme() { applyTheme(!isDark()); render(); }

  // ---------- KEYBOARD ----------
  window.addEventListener('keydown', function (e) {
    if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') { e.preventDefault(); openSearch(); return; }
    if (!state.searchOpen) return;
    if (e.key === 'Escape') { closeSearch(); }
    else if (e.key === 'ArrowDown') { e.preventDefault(); moveSearch(1); }
    else if (e.key === 'ArrowUp') { e.preventDefault(); moveSearch(-1); }
    else if (e.key === 'Enter') {
      var results = searchResults();
      if (results[state.searchIndex]) goto(results[state.searchIndex].r.id);
    }
  });
  function moveSearch(delta) {
    var n = searchResults().length;
    if (!n) return;
    state.searchIndex = (state.searchIndex + delta + n) % n;
    render();
    var input = document.getElementById('search-input');
    if (input) input.focus();
  }

  // ---------- INIT ----------
  (function init() {
    var stored;
    try { stored = localStorage.getItem('unifidoc-theme'); } catch (e) {}
    if (stored) applyTheme(stored === 'dark');
    var hash = (location.hash || '').replace(/^#/, '');
    state.active = (hash && byId(hash)) ? hash : (flat[0] ? flat[0].r.id : null);
    render();
  })();
})();
`
