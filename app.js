// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE.
const go = new Go();
let wasmReady = false;

// Try to load Wasm
WebAssembly.instantiateStreaming(
  fetch("mks.wasm?v=" + new Date().getTime()),
  go.importObject
)
  .then((result) => {
    go.run(result.instance);
    console.log("Wasm loaded");
    wasmReady = true;
    // Provide data from Wasm to frontend
    initializeDataFromWasm();
  })
  .catch((err) => {
    console.log("Wasm not loaded (maybe running on Go server):", err);
    // Fallback to fetch API if Wasm fails (e.g. running on local server without wasm build)
    initializeDataFromFetch();
  });

// Configure Marked
marked.setOptions({
  highlight: function (code, lang) {
    const language = hljs.getLanguage(lang) ? lang : "plaintext";
    return hljs.highlight(code, { language }).value;
  },
  langPrefix: "hljs language-",
});

var sqlEditor, jsonEditor;

function getAceTheme() {
  const isDark = document.body.classList.contains("dark-mode");
  return isDark ? "ace/theme/tomorrow_night" : "ace/theme/tomorrow";
}

function initAceEditors() {
  const sqlContainer = document.getElementById("sql-editor");
  const jsonContainer = document.getElementById("json-editor");
  const sqlTextarea = document.getElementById("sql-text");
  const jsonTextarea = document.getElementById("json-input");

  if (sqlContainer && sqlTextarea) {
    sqlEditor = ace.edit("sql-editor");
    sqlEditor.setTheme(getAceTheme());
    sqlEditor.session.setMode("ace/mode/sql");
    sqlEditor.setShowPrintMargin(false);
    sqlEditor.setOptions({
      fontSize: "0.9rem",
      fontFamily:
        "'JetBrains Mono', 'Fira Code', 'Consolas', 'Monaco', 'Courier New', monospace",
      tabSize: 2,
      useSoftTabs: true,
      enableBasicAutocompletion: true,
      enableLiveAutocompletion: true,
    });
    sqlEditor.setValue(sqlTextarea.value, -1);
    sqlEditor.on("change", () => {
      sqlTextarea.value = sqlEditor.getValue();
      saveState();
    });
  }

  if (jsonContainer && jsonTextarea) {
    jsonEditor = ace.edit("json-editor");
    jsonEditor.setTheme(getAceTheme());
    jsonEditor.session.setMode("ace/mode/json");
    jsonEditor.setShowPrintMargin(false);
    jsonEditor.setOptions({
      fontSize: "0.9rem",
      fontFamily:
        "'JetBrains Mono', 'Fira Code', 'Consolas', 'Monaco', 'Courier New', monospace",
      tabSize: 2,
      useSoftTabs: true,
      enableBasicAutocompletion: true,
      enableLiveAutocompletion: true,
    });
    jsonEditor.setValue(jsonTextarea.value, -1);
    jsonEditor.on("change", () => {
      jsonTextarea.value = jsonEditor.getValue();
      validateJSON();
      saveState();
    });
  }
}

function initializeDataFromWasm() {
  if (typeof getRules !== "function") {
    setTimeout(initializeDataFromWasm, 100);
    return;
  }

  try {
    const rules = JSON.parse(getRules());
    const tests = JSON.parse(getTests());
    if (typeof getVersion === "function") {
      const verInfo = getVersion();
      if (verInfo && typeof verInfo === "object") {
        const verEl = document.getElementById("app-version");
        verEl.textContent = verInfo.version;
        if (verInfo.last_build) {
          const buildEl = document.getElementById("app-build-time");
          const buildContainer = document.getElementById(
            "build-info-container"
          );
          if (buildEl && buildContainer) {
            buildEl.textContent = verInfo.last_build;
            buildContainer.style.display = "inline";
          }
          verEl.title = "Built: " + verInfo.last_build;
          verEl.style.cursor = "help";
        }
      } else {
        document.getElementById("app-version").textContent = verInfo;
      }
    }
    setupUI(tests, rules);
  } catch (e) {
    console.error("Error loading data from Wasm:", e);
  }
}

function initializeDataFromFetch() {
  Promise.all([
    fetch("/tests").then((r) => r.json()),
    fetch("/rules").then((r) => r.json()),
    fetch("/version").then((r) => r.json()),
  ])
    .then(([tests, rules, verInfo]) => {
      if (verInfo && verInfo.version) {
        const verEl = document.getElementById("app-version");
        verEl.textContent = "Version: " + verInfo.version;
        if (verInfo.last_build) {
          verEl.title = "Built: " + verInfo.last_build;
          verEl.style.cursor = "help";
        }
      } else {
        document.getElementById("app-version").textContent = "1.0.0 (Server)";
      }
      setupUI(tests, rules);
    })
    .catch((e) => console.error("Fetch failed:", e));
}

// Set Year
const yearEl = document.getElementById("year");
if (yearEl) yearEl.textContent = new Date().getFullYear();

function initDatabaseChooser() {
  restoreManualConnectionData();
  fetch("/databases")
    .then((r) => r.json())
    .then((dbs) => {
      const select = document.getElementById("db-chooser");
      if (!select) return;
      window.predefinedDBs = dbs;
      // Keep first two options (placeholder and manual)
      select.innerHTML =
        '<option value="">DB Connection...</option><option value="manual">Manual Connection...</option>';
      dbs.forEach((db) => {
        const opt = document.createElement("option");
        opt.value = db.name;
        opt.textContent = db.name;
        select.appendChild(opt);
      });
    })
    .catch((e) => console.error("Failed to load databases:", e));
}

function connectToPredefined(dbName) {
  if (dbName === "") return;
  if (dbName === "manual") {
    toggleDBModal();
    document.getElementById("db-chooser").value = "";
    return;
  }

  const db = window.predefinedDBs.find((d) => d.name === dbName);
  if (!db) return;

  performConnect(db);
}

function connectManual(event) {
  event.preventDefault();
  saveManualConnectionData();
  const cfg = {
    name: "Manual",
    host: document.getElementById("db-host").value,
    port: document.getElementById("db-port").value,
    user: document.getElementById("db-user").value,
    password: document.getElementById("db-pass").value,
    database: document.getElementById("db-name").value,
    schema: document.getElementById("db-schema").value,
  };

  performConnect(cfg);
  toggleDBModal();
}

async function performConnect(cfg) {
  const select = document.getElementById("db-chooser");
  const disconnectBtn = document.getElementById("db-disconnect");

  try {
    const response = await fetch("/connect", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(cfg),
    });

    const data = await response.json();
    if (data && data.error) {
      logError("Connection failed: " + data.error);
      updateDBUI(false, null, true);
    } else {
      updateDBUI(true, cfg.name);
    }
  } catch (e) {
    logError("Network error: " + e.message);
    updateDBUI(false, null, true);
  }
}

async function disconnectDB() {
  try {
    await fetch("/disconnect", { method: "POST" });
    updateDBUI(false);
  } catch (e) {
    console.error("Disconnect failed:", e);
  }
}

function updateDBUI(connected, name, failed = false) {
  const select = document.getElementById("db-chooser");
  const group = select.closest(".db-connection-group");
  const disconnectBtn = document.getElementById("db-disconnect");
  const runQueryBtn = document.getElementById("run-query-btn");
  const sqlResultPane = document.getElementById("pane-sql-result");
  const gutter2 = document.getElementById("gutter-h2");
  const logsPane = document.getElementById("pane-logs");
  const gutterLogs = document.getElementById("gutter-logs");

  // Clear previous states
  if (group) {
    group.classList.remove("connected-success", "connected-fail");
  }

  if (connected) {
    select.classList.add("connected");
    if (group) group.classList.add("connected-success");
    disconnectBtn.classList.remove("hidden");
    if (runQueryBtn) runQueryBtn.classList.remove("hidden");
    if (sqlResultPane) sqlResultPane.classList.remove("hidden");
    if (gutter2) gutter2.classList.remove("hidden");
    if (logsPane) logsPane.classList.remove("hidden");
    if (gutterLogs) gutterLogs.classList.remove("hidden");

    // If it was manual, we might want to show "Manual" or the DB name
    if (name) {
      // Find or create option for the connected DB if not "Manual" or already there
      let opt = Array.from(select.options).find((o) => o.textContent === name);
      if (!opt) {
        opt = document.createElement("option");
        opt.value = name;
        opt.textContent = name;
        select.appendChild(opt);
      }
      select.value = opt.value;
    }
  } else {
    select.classList.remove("connected");
    if (failed && group) {
      group.classList.add("connected-fail");
    }
    disconnectBtn.classList.add("hidden");
    if (runQueryBtn) runQueryBtn.classList.add("hidden");
    if (sqlResultPane) sqlResultPane.classList.add("hidden");
    if (gutter2) gutter2.classList.add("hidden");
    if (logsPane) logsPane.classList.add("hidden");
    if (gutterLogs) gutterLogs.classList.add("hidden");
    select.value = "";
  }
}

function saveManualConnectionData() {
  const data = {
    host: document.getElementById("db-host").value,
    port: document.getElementById("db-port").value,
    user: document.getElementById("db-user").value,
    database: document.getElementById("db-name").value,
    schema: document.getElementById("db-schema").value,
  };
  localStorage.setItem("mks_manual_db_conn", JSON.stringify(data));
}

function restoreManualConnectionData() {
  const saved = localStorage.getItem("mks_manual_db_conn");
  if (saved) {
    try {
      const data = JSON.parse(saved);
      const fields = {
        "db-host": "host",
        "db-port": "port",
        "db-user": "user",
        "db-name": "database",
        "db-schema": "schema",
      };
      for (const [id, key] of Object.entries(fields)) {
        const el = document.getElementById(id);
        if (el) el.value = data[key] || "";
      }
    } catch (e) {
      console.error("Failed to restore manual connection data:", e);
    }
  }
}

function toggleDBModal() {
  const modal = document.getElementById("db-modal");
  if (modal)
    modal.style.display = modal.style.display === "block" ? "none" : "block";
}

function setupUI(tests, rules) {
  const select = document.getElementById("test-chooser");
  window.tests = tests; // Store globally
  window.rules = rules; // Store globally

  const ruleMap = {};
  rules.forEach((p) => {
    ruleMap[p.id] = p.description;
  });

  tests.sort((a, b) => a.id - b.id);
  select.innerHTML = '<option value="">Select a Test Case...</option>';

  tests.forEach((test, index) => {
    const opt = document.createElement("option");
    opt.value = index;

    let desc = "Unknown Rule";
    if (test.description && test.description.length > 0) {
      desc = test.description;
    } else if (ruleMap[test.id]) {
      desc = ruleMap[test.id];
    }

    if (desc.length > 120) desc = desc.substring(0, 117) + "...";

    opt.textContent = `${test.id} - ${desc}`;
    select.appendChild(opt);
  });
}

function loadDoc(filename) {
  const url = filename;
  fetch(url)
    .then((r) => {
      if (!r.ok && filename === "reference_guide.md" && !url.includes("doc/")) {
        return fetch("doc/reference_guide.md");
      }
      return r;
    })
    .then((r) => {
      if (!r.ok) throw new Error("Document not found: " + url);
      return r.text();
    })
    .then((text) => {
      const html = marked.parse(text);
      const docContent = document.getElementById("doc-content");

      let navHtml = "";
      if (filename !== "reference_guide.md") {
        navHtml = `<p><a href="#" onclick="event.preventDefault(); loadDoc('reference_guide.md'); return false;">&larr; Back to Reference Guide</a></p><hr>`;
      }

      docContent.innerHTML = navHtml + html;
      docContent.scrollTop = 0;
      docContent.style.whiteSpace = "normal";
      docContent.style.fontFamily = "inherit";
      bindDocLinks();
    })
    .catch((err) => {
      const docContent = document.getElementById("doc-content");
      if (docContent)
        docContent.innerHTML = `<p style="color:red">Error loading documentation: ${err.message}</p>`;
    });
}

function bindDocLinks() {
  const docContent = document.getElementById("doc-content");
  if (!docContent) return;
  const links = docContent.querySelectorAll("a");
  links.forEach((link) => {
    const href = link.getAttribute("href");
    if (href && href.endsWith(".md")) {
      link.onclick = function (e) {
        e.preventDefault();
        loadDoc(href);
      };
    }
  });
}

function loadTest(index) {
  if (index === "" || index === null) return;
  const test = window.tests[index];
  if (test) {
    if (sqlEditor) {
      sqlEditor.setValue(test.text, -1);
    } else {
      document.getElementById("sql-text").value = test.text;
    }

    let jsonVal = JSON.stringify(test.input, null, 2);
    if (jsonEditor) {
      jsonEditor.setValue(jsonVal, -1);
    } else {
      document.getElementById("json-input").value = jsonVal;
    }
    document.getElementById("result-output").textContent = "";
    validateJSON();
    saveState();
  }
}

async function runProcess() {
  const sql = sqlEditor
    ? sqlEditor.getValue()
    : document.getElementById("sql-text").value;
  const input = jsonEditor
    ? jsonEditor.getValue()
    : document.getElementById("json-input").value;
  const outputDiv = document.getElementById("result-output");
  const minify = document
    .getElementById("minify-btn")
    .classList.contains("active");

  logEntry("--- Processing SQL ---", "info");
  logEntry(sql, "sql");
  outputDiv.textContent = "Processing...";
  outputDiv.className = "output-area";

  if (wasmReady && typeof processSql === "function") {
    try {
      const result = processSql(sql, input, minify);
      outputDiv.textContent = result;
      return result;
    } catch (e) {
      const msg = "Wasm Error: " + e;
      outputDiv.textContent = msg;
      outputDiv.classList.add("error-text");
      logError(msg);
      return null;
    }
  } else {
    try {
      const response = await fetch("/process", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ sql, input, minify }),
      });

      const data = await response.json();
      if (data && data.error) {
        const msg = "Process Error: " + data.error;
        outputDiv.textContent = msg;
        outputDiv.classList.add("error-text");
        logError(msg);
        return null;
      } else {
        outputDiv.textContent = data.result;
        return data.result;
      }
    } catch (e) {
      const msg = "Network Error: " + e.message;
      outputDiv.textContent = msg;
      outputDiv.classList.add("error-text");
      logError(msg);
      return null;
    }
  }
}

async function runProcessAndQuery() {
  const parsedSql = await runProcess();
  if (!parsedSql) return;

  logEntry("--- Executing Query ---", "info");
  logEntry(parsedSql, "sql");

  const statusEl = document.getElementById("query-status");
  const tableHead = document.getElementById("sql-table-head");
  const tableBody = document.getElementById("sql-table-body");

  statusEl.textContent = "Executing query...";
  tableHead.innerHTML = "";
  tableBody.innerHTML = "";

  try {
    const input = jsonEditor.getValue();
    const response = await fetch("/query", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        sql: parsedSql,
        params: JSON.parse(input),
      }),
    });

    const data = await response.json();

    // Log substituted SQL if available (even on error)
    if (data && data.substituted_sql) {
      logEntry("--- Executing Query ---", "info");
      logEntry(data.substituted_sql, "sql");
    }

    if (data && data.error) {
      const msg = "Query Error: " + data.error;
      statusEl.textContent = msg;
      statusEl.style.color = "var(--danger)";
      logError(msg);
    } else if (data) {
      displayQueryResult(data.rows || []);
    }
  } catch (e) {
    const msg = "Network error: " + e.message;
    statusEl.textContent = msg;
    statusEl.style.color = "var(--danger)";
    logError(msg);
  }
}

function displayQueryResult(data) {
  const statusEl = document.getElementById("query-status");
  const tableHead = document.getElementById("sql-table-head");
  const tableBody = document.getElementById("sql-table-body");

  if (!Array.isArray(data) || data.length === 0) {
    statusEl.textContent = "No rows returned.";
    statusEl.style.color = "var(--text-color)";
    return;
  }

  statusEl.textContent = `${data.length} rows returned.`;
  statusEl.style.color = "var(--success)";

  const columns = Object.keys(data[0]);

  // Headers
  columns.forEach((col) => {
    const th = document.createElement("th");
    th.textContent = col;
    tableHead.appendChild(th);
  });

  // Body
  data.forEach((row) => {
    const tr = document.createElement("tr");
    columns.forEach((col) => {
      const td = document.createElement("td");
      let val = row[col];
      if (typeof val === "object" && val !== null) {
        val = JSON.stringify(val);
      }
      td.textContent = val;
      tr.appendChild(td);
    });
    tableBody.appendChild(tr);
  });
}

function logEntry(content, type = "sql") {
  const logOutput = document.getElementById("log-output");
  if (!logOutput) return;

  const entry = document.createElement("div");
  entry.className = "log-entry";
  if (type === "error") entry.classList.add("error");

  const time = document.createElement("span");
  time.className = "log-time";
  time.textContent = new Date().toLocaleTimeString();

  const bodyContent = document.createElement("span");
  bodyContent.className = "log-sql";
  bodyContent.textContent = content;

  entry.appendChild(time);
  entry.appendChild(bodyContent);
  logOutput.insertBefore(entry, logOutput.firstChild);

  // Auto-expand logs pane if it's hidden or collapsed
  const pane = document.getElementById("pane-logs");
  const gutter = document.getElementById("gutter-logs");
  if (pane && pane.classList.contains("hidden")) {
    pane.classList.remove("hidden");
    if (gutter) gutter.classList.remove("hidden");
  }
  if (pane && pane.classList.contains("collapsed")) {
    toggleFold("pane-logs");
  }
}

function logError(msg) {
  console.error(msg);
  logEntry(msg, "error");
}

function clearLogs() {
  const logOutput = document.getElementById("log-output");
  if (logOutput) logOutput.innerHTML = "";
}

function toggleMinify() {
  const btn = document.getElementById("minify-btn");
  btn.classList.toggle("active");
}

function toggleTheme() {
  const body = document.body;
  const icon = document.querySelector("#theme-toggle i");
  if (body.classList.contains("light-mode")) {
    body.classList.replace("light-mode", "dark-mode");
    icon.classList.replace("fa-moon", "fa-lightbulb");
  } else {
    body.classList.replace("dark-mode", "light-mode");
    icon.classList.replace("fa-lightbulb", "fa-moon");
  }
  updateHighlightTheme();
  if (sqlEditor) sqlEditor.setTheme(getAceTheme());
  if (jsonEditor) jsonEditor.setTheme(getAceTheme());
}

function updateHighlightTheme() {
  if (typeof hljs === "undefined") return;
  const isDark = document.body.classList.contains("dark-mode");
  const themeLink = document.getElementById("hljs-theme");
  if (themeLink) {
    themeLink.href = isDark
      ? "https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/github-dark.min.css"
      : "https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/github.min.css";
  }
}

function toggleDoc() {
  const modal = document.getElementById("doc-modal");
  if (modal)
    modal.style.display = modal.style.display === "block" ? "none" : "block";
}

function toggleHelp() {
  const modal = document.getElementById("help-modal");
  if (modal)
    modal.style.display = modal.style.display === "block" ? "none" : "block";
}

function toggleFold(paneId) {
  const pane = document.getElementById(paneId);
  if (!pane) return;
  pane.classList.toggle("collapsed");
  const btn = pane.querySelector(".fold-btn i");
  if (btn) {
    if (pane.classList.contains("collapsed")) {
      btn.classList.replace("fa-minus", "fa-plus");
    } else {
      btn.classList.replace("fa-plus", "fa-minus");
    }
  }
  // Resize editors when panes fold/unfold
  if (sqlEditor) sqlEditor.resize();
  if (jsonEditor) jsonEditor.resize();
}

window.onclick = function (event) {
  const docModal = document.getElementById("doc-modal");
  const helpModal = document.getElementById("help-modal");
  const dbModal = document.getElementById("db-modal");
  if (event.target == docModal) docModal.style.display = "none";
  if (event.target == helpModal) helpModal.style.display = "none";
  if (event.target == dbModal) dbModal.style.display = "none";
};

function setupDragDrop(elementId) {
  const dropZone = document.getElementById(
    elementId === "sql-text" ? "pane-sql" : "pane-json"
  );
  if (!dropZone) return;

  ["dragenter", "dragover", "dragleave", "drop"].forEach((eventName) => {
    dropZone.addEventListener(
      eventName,
      (e) => {
        e.preventDefault();
        e.stopPropagation();
      },
      false
    );
  });

  ["dragenter", "dragover"].forEach((eventName) => {
    dropZone.addEventListener(
      eventName,
      () => dropZone.classList.add("drag-over"),
      false
    );
  });

  ["dragleave", "drop"].forEach((eventName) => {
    dropZone.addEventListener(
      eventName,
      () => dropZone.classList.remove("drag-over"),
      false
    );
  });

  dropZone.addEventListener(
    "drop",
    (e) => {
      const files = e.dataTransfer.files;
      if (files.length > 0) processFile(files[0], elementId);
    },
    false
  );
}

function loadFile(input, targetId) {
  if (input.files && input.files[0]) {
    processFile(input.files[0], targetId);
    input.value = "";
  }
}

function processFile(file, targetId) {
  const reader = new FileReader();
  reader.onload = function (e) {
    const text = e.target.result;
    if (text.includes("select mks_sql_ins(")) {
      if (typeof MksSqlInsParser !== "undefined") {
        const parser = new MksSqlInsParser();
        const parsed = parser.parse(text);
        if (parsed) {
          if (parsed.sql) {
            if (sqlEditor) sqlEditor.setValue(parsed.sql, -1);
            else document.getElementById("sql-text").value = parsed.sql;
          }
          if (parsed.info && parsed.info.default !== undefined) {
            let val = parsed.info.default;
            let jsonStr =
              typeof val === "object" ? JSON.stringify(val, null, 2) : val;
            if (jsonEditor) jsonEditor.setValue(jsonStr, -1);
            else document.getElementById("json-input").value = jsonStr;
            validateJSON();
          }
          saveState();
          return;
        }
      }
    }

    if (targetId === "sql-text" && sqlEditor) sqlEditor.setValue(text, -1);
    else if (targetId === "json-input" && jsonEditor)
      jsonEditor.setValue(text, -1);
    else {
      const target = document.getElementById(targetId);
      target.value = text;
    }

    if (targetId === "json-input") validateJSON();
    saveState();
  };
  reader.readAsText(file);
}

function validateJSON() {
  const val = jsonEditor
    ? jsonEditor.getValue().trim()
    : document.getElementById("json-input").value.trim();
  const pane = document.getElementById("pane-json");
  if (!val) {
    if (pane) pane.style.borderColor = "var(--border-color)";
    return;
  }
  try {
    JSON.parse(val);
    if (pane) pane.style.borderColor = "var(--success)";
  } catch (e) {
    if (pane) pane.style.borderColor = "var(--danger)";
    // Don't log to window while typing, only console
    console.warn("JSON Parse Error:", e.message);
  }
}

function formatJson() {
  if (!jsonEditor) return;
  const val = jsonEditor.getValue().trim();
  if (!val) return;
  try {
    const obj = JSON.parse(val);
    jsonEditor.setValue(JSON.stringify(obj, null, 2), -1);
  } catch (e) {
    console.warn("Invalid JSON for formatting");
  }
}

function copyToClipboard(elementId) {
  let text = "";
  if (elementId === "sql-text" && sqlEditor) text = sqlEditor.getValue();
  else if (elementId === "json-input" && jsonEditor)
    text = jsonEditor.getValue();
  else {
    const el = document.getElementById(elementId);
    if (!el) return;
    text =
      el.tagName === "TEXTAREA" || el.tagName === "INPUT"
        ? el.value
        : el.textContent;
  }
  navigator.clipboard
    .writeText(text)
    .catch((err) => console.error("Failed to copy: ", err));
}

function downloadResult() {
  const text = document.getElementById("result-output").textContent;
  const blob = new Blob([text], { type: "text/plain" });
  const url = window.URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = "result.sql";
  document.body.appendChild(a);
  a.click();
  window.URL.revokeObjectURL(url);
  document.body.removeChild(a);
}

function saveState() {
  const sql = sqlEditor
    ? sqlEditor.getValue()
    : document.getElementById("sql-text").value;
  const json = jsonEditor
    ? jsonEditor.getValue()
    : document.getElementById("json-input").value;
  localStorage.setItem("mks_workspace_sql", sql);
  localStorage.setItem("mks_workspace_json", json);
}

let stateRestored = false;
function restoreState() {
  if (stateRestored) return;
  const savedSql = localStorage.getItem("mks_workspace_sql");
  const savedJson = localStorage.getItem("mks_workspace_json");
  if (savedSql !== null) {
    if (sqlEditor) sqlEditor.setValue(savedSql, -1);
    else document.getElementById("sql-text").value = savedSql;
  }
  if (savedJson !== null) {
    if (jsonEditor) jsonEditor.setValue(savedJson, -1);
    else document.getElementById("json-input").value = savedJson;
    validateJSON();
  }
  stateRestored = true;
}

function initGutter(gutterId, primaryPaneId, secondaryPaneId, isVertical) {
  const gutter = document.getElementById(gutterId);
  const primaryPane = document.getElementById(primaryPaneId);
  const secondaryPane = secondaryPaneId
    ? document.getElementById(secondaryPaneId)
    : null;
  if (!gutter || !primaryPane) return;

  let startPos, startDim;

  gutter.addEventListener("mousedown", function (e) {
    startPos = isVertical ? e.clientX : e.clientY;
    const rect = primaryPane.getBoundingClientRect();
    startDim = isVertical ? rect.width : rect.height;

    document.body.classList.add("resizing");
    document.body.style.cursor = isVertical ? "col-resize" : "row-resize";
    e.preventDefault();

    const onMove = function (e) {
      const currentPos = isVertical ? e.clientX : e.clientY;
      const delta = currentPos - startPos;
      const newDim = startDim + delta;

      if (newDim > 40) {
        primaryPane.style.flex = `0 0 ${newDim}px`;
        if (secondaryPane) secondaryPane.style.flex = "1 1 auto";
        if (sqlEditor) sqlEditor.resize();
        if (jsonEditor) jsonEditor.resize();
      }
    };

    const onUp = function () {
      document.body.classList.remove("resizing");
      document.body.style.cursor = "default";
      document.removeEventListener("mousemove", onMove);
      document.removeEventListener("mouseup", onUp);
    };

    document.addEventListener("mousemove", onMove);
    document.addEventListener("mouseup", onUp);
  });
}

function setupResizing() {
  initGutter("gutter-v", "left-col", "right-col", true);
  initGutter("gutter-h", "pane-sql", "pane-json", false);
  initGutter("gutter-h2", "pane-result", "pane-sql-result", false);
  initGutter("gutter-logs", "main-row", "pane-logs", false);
}

// Initialization
document.addEventListener("DOMContentLoaded", () => {
  initAceEditors();
  restoreState();
  setupResizing();
  setupDragDrop("sql-text");
  setupDragDrop("json-input");
  loadDoc("reference_guide.md");
  updateHighlightTheme();
  initDatabaseChooser();
});
