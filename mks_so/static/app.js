
// Use Wasm if not running on localhost:8080 (or if explicitly enabled, but for GitHub Pages we assume Wasm)
// Or better: try to fetch /process, if fails, use Wasm.
// Actually, for GitHub Pages migration, we can just default to Wasm if the file exists or just always try to load it.
// Let's implement a Wasm loader.

const go = new Go();
let wasmReady = false;

// Try to load Wasm
WebAssembly.instantiateStreaming(fetch("mks.wasm?v=" + new Date().getTime()), go.importObject).then((result) => {
    go.run(result.instance);
    console.log("Wasm loaded");
    wasmReady = true;
    // Provide data from Wasm to frontend
    initializeDataFromWasm();
}).catch((err) => {
    console.log("Wasm not loaded (maybe running on Go server):", err);
    // Fallback to fetch API if Wasm fails (e.g. running on local server without wasm build)
    initializeDataFromFetch();
});

// Configure Marked
marked.setOptions({
    highlight: function (code, lang) {
        const language = hljs.getLanguage(lang) ? lang : 'plaintext';
        return hljs.highlight(code, { language }).value;
    },
    langPrefix: 'hljs language-'
});

function initializeDataFromWasm() {
    // Wasm functions: getRules(), getTests() return JSON strings
    if (typeof getRules !== 'function') {
        setTimeout(initializeDataFromWasm, 100);
        return;
    }

    try {
        const rules = JSON.parse(getRules());
        const tests = JSON.parse(getTests());
        // Version might be string or need parsing? It's just a string.
        if (typeof getVersion === 'function') {
            const verInfo = getVersion();
            // Handle both new object format and potential legacy string
            if (verInfo && typeof verInfo === 'object') {
                const verEl = document.getElementById('app-version');
                verEl.textContent = verInfo.version;
                if (verInfo.last_build) {
                    const buildEl = document.getElementById('app-build-time');
                    const buildContainer = document.getElementById('build-info-container');
                    if (buildEl && buildContainer) {
                        buildEl.textContent = verInfo.last_build;
                        buildContainer.style.display = "inline";
                    }
                    verEl.title = "Built: " + verInfo.last_build;
                    verEl.style.cursor = "help";
                }
            } else {
                document.getElementById('app-version').textContent = verInfo;
            }
        }
        setupUI(tests, rules);
    } catch (e) {
        console.error("Error loading data from Wasm:", e);
    }
}

function initializeDataFromFetch() {
    Promise.all([
        fetch('/tests').then(r => r.json()),
        fetch('/rules').then(r => r.json()),
        fetch('/version').then(r => r.json())
    ]).then(([tests, rules, verInfo]) => {
        if (verInfo && verInfo.version) {
            const verEl = document.getElementById('app-version');
            verEl.textContent = "Version: " + verInfo.version;
            if (verInfo.last_build) {
                verEl.title = "Built: " + verInfo.last_build;
                verEl.style.cursor = "help";
            }
        } else {
            document.getElementById('app-version').textContent = "1.0.0 (Server)";
        }
        setupUI(tests, rules);
    }).catch(e => console.error("Fetch failed:", e));
}

// Set Year
const yearEl = document.getElementById('year');
if (yearEl) yearEl.textContent = new Date().getFullYear();

function setupUI(tests, rules) {
    const select = document.getElementById('test-chooser');
    window.tests = tests; // Store globally
    window.rules = rules; // Store globally

    // Create Rule Lookup Map
    const ruleMap = {};
    rules.forEach(p => {
        ruleMap[p.id] = p.description;
    });

    // Sort tests by ID ascending
    tests.sort((a, b) => a.id - b.id);

    // Clear options first
    select.innerHTML = '<option value="">Select a Test Case...</option>';

    tests.forEach((test, index) => {
        const opt = document.createElement('option');
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

// Load Doc
// On GitHub Pages, we are likely at /mks_sql_parser/
// file is at /mks_sql_parser/doc/reference_guide.md if copied?
// Wait, the deploy action only copies 'mks_so/cmd/server/static'.
// Does 'static' contain 'doc'? No.
// We need to copy 'doc' folder to 'static' in the build script!

const docBaseUrl = "";

function loadDoc(filename) {
    const url = (docBaseUrl && !filename.includes('/') ? docBaseUrl + '/' : '') + filename;

    // Handle fallback for local dev vs production paths if needed
    // Currently simplified: try direct fetch
    fetch(url).then(r => {
        if (!r.ok && filename === "reference_guide.md" && !url.includes("doc/")) {
            // Fallback check for local dev structure: try "doc/reference_guide.md"
            return fetch("doc/reference_guide.md"); // Hard fallback
        }
        return r;
    })
        .then(r => {
            if (!r.ok) throw new Error("Document not found: " + url);
            return r.text();
        })
        .then(text => {
            const html = marked.parse(text);
            const docContent = document.getElementById('doc-content');

            let navHtml = '';
            if (filename !== 'reference_guide.md') {
                navHtml = `<p><a href="#" onclick="event.preventDefault(); loadDoc('reference_guide.md'); return false;">&larr; Back to Reference Guide</a></p><hr>`;
            }

            docContent.innerHTML = navHtml + html;
            docContent.scrollTop = 0; // consistent view

            // Style adjustments
            docContent.style.whiteSpace = 'normal';
            docContent.style.fontFamily = 'inherit';

            // Re-bind links
            bindDocLinks();
        })
        .catch(err => {
            const docContent = document.getElementById('doc-content');
            if (docContent) docContent.innerHTML = `<p style="color:red">Error loading documentation: ${err.message}</p>`;
        });
}

function bindDocLinks() {
    const docContent = document.getElementById('doc-content');
    if (!docContent) return;
    const links = docContent.querySelectorAll('a');
    links.forEach(link => {
        const href = link.getAttribute('href');
        if (href && href.endsWith('.md')) {
            link.onclick = function (e) {
                e.preventDefault();
                loadDoc(href);
            };
        }
    });
}

// Initial Load
loadDoc("reference_guide.md");


// Initial Theme Check
updateHighlightTheme();

// document.addEventListener('DOMContentLoaded', () => { ... }); // Removed, we call init above

function getTestDescription(test) {
    return `Test Case ${test.id}`;
}

function loadTest(index) {
    if (index === "" || index === null) return;
    const test = window.tests[index];
    if (test) {
        document.getElementById('sql-text').value = test.text;
        document.getElementById('json-input').value = JSON.stringify(test.input, null, 2);
        document.getElementById('result-output').textContent = "";
    }
}

async function runProcess() {
    const sql = document.getElementById('sql-text').value;
    const input = document.getElementById('json-input').value;
    const outputDiv = document.getElementById('result-output');
    const minify = document.getElementById('minify-btn').classList.contains('active');

    outputDiv.textContent = "Processing...";
    outputDiv.className = "output-area";

    if (wasmReady && typeof processSql === 'function') {
        try {
            // Wasm execution
            // processSql returns string result directly
            const result = processSql(sql, input, minify);
            outputDiv.textContent = result;
        } catch (e) {
            outputDiv.textContent = "Wasm Error: " + e;
            outputDiv.classList.add("error-text");
        }
    } else {
        // Server fallback
        try {
            const response = await fetch('/process', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ sql, input, minify })
            });

            const data = await response.json();
            if (data.error) {
                outputDiv.textContent = "Error: " + data.error;
                outputDiv.classList.add("error-text");
            } else {
                outputDiv.textContent = data.result;
            }
        } catch (e) {
            outputDiv.textContent = "Network Error: " + e.message;
            outputDiv.classList.add("error-text");
        }
    }
}

function toggleMinify() {
    const btn = document.getElementById('minify-btn');
    btn.classList.toggle('active');
}

// Initial Theme Check
// Check system preference
if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
    document.body.classList.replace('light-mode', 'dark-mode');
    document.querySelector('#theme-toggle i').classList.replace('fa-moon', 'fa-lightbulb');
}
updateHighlightTheme();

// Listen for system changes
if (window.matchMedia) {
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', event => {
        const body = document.body;
        const icon = document.querySelector('#theme-toggle i');
        // Only update if user hasn't toggled manually?
        // For simplicity, let system override or just sync.
        // Let's just sync to system change.
        if (event.matches) {
            if (body.classList.contains('light-mode')) {
                body.classList.replace('light-mode', 'dark-mode');
                icon.classList.replace('fa-moon', 'fa-lightbulb');
            }
        } else {
            if (body.classList.contains('dark-mode')) {
                body.classList.replace('dark-mode', 'light-mode');
                icon.classList.replace('fa-lightbulb', 'fa-moon');
            }
        }
        updateHighlightTheme();
    });
}

function toggleTheme() {
    const body = document.body;
    const icon = document.querySelector('#theme-toggle i');
    if (body.classList.contains('light-mode')) {
        body.classList.replace('light-mode', 'dark-mode');
        icon.classList.replace('fa-moon', 'fa-lightbulb');
    } else {
        body.classList.replace('dark-mode', 'light-mode');
        icon.classList.replace('fa-lightbulb', 'fa-moon');
    }
    updateHighlightTheme();
}

function updateHighlightTheme() {
    // Check if highlight.js is loaded
    if (typeof hljs === 'undefined') return;

    const isDark = document.body.classList.contains('dark-mode');
    const themeLink = document.getElementById('hljs-theme');
    if (themeLink) {
        if (isDark) {
            themeLink.href = "https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/github-dark.min.css";
        } else {
            themeLink.href = "https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/github.min.css";
        }
    }
}

function toggleDoc() {
    const modal = document.getElementById('doc-modal');
    if (modal) modal.style.display = modal.style.display === 'block' ? 'none' : 'block';
}

function toggleHelp() {
    const modal = document.getElementById('help-modal');
    if (modal) modal.style.display = modal.style.display === 'block' ? 'none' : 'block';
}

function toggleFold(paneId) {
    const pane = document.getElementById(paneId);
    if (!pane) return;
    pane.classList.toggle('collapsed');
    // Change icon?
    const btn = pane.querySelector('.fold-btn i');
    if (btn) {
        if (pane.classList.contains('collapsed')) {
            btn.classList.replace('fa-minus', 'fa-plus');
        } else {
            btn.classList.replace('fa-plus', 'fa-minus');
        }
    }
}

// Global click to close modals
window.onclick = function (event) {
    const docModal = document.getElementById('doc-modal');
    const helpModal = document.getElementById('help-modal');
    if (event.target == docModal) {
        docModal.style.display = "none";
    }
    if (event.target == helpModal) {
        helpModal.style.display = "none";
    }
}

/* --- New UX Features --- */

// 1. File Upload & Drag-n-Drop
function setupDragDrop(elementId) {
    const dropZone = document.getElementById(elementId);
    if (!dropZone) return;

    ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
        dropZone.addEventListener(eventName, preventDefaults, false);
    });

    function preventDefaults(e) {
        e.preventDefault();
        e.stopPropagation();
    }

    ['dragenter', 'dragover'].forEach(eventName => {
        dropZone.addEventListener(eventName, highlight, false);
    });

    ['dragleave', 'drop'].forEach(eventName => {
        dropZone.addEventListener(eventName, unhighlight, false);
    });

    function highlight(e) {
        dropZone.classList.add('drag-over');
    }

    function unhighlight(e) {
        dropZone.classList.remove('drag-over');
    }

    dropZone.addEventListener('drop', handleDrop, false);

    function handleDrop(e) {
        const dt = e.dataTransfer;
        const files = dt.files;
        if (files.length > 0) {
            processFile(files[0], elementId);
        }
    }
}

// Initialize Drag & Drop
setupDragDrop('sql-text');
setupDragDrop('json-input');

function loadFile(input, targetId) {
    if (input.files && input.files[0]) {
        processFile(input.files[0], targetId);
        // Reset input so same file can be selected again if needed
        input.value = '';
    }
}

function processFile(file, targetId) {
    const reader = new FileReader();
    reader.onload = function (e) {
        const text = e.target.result;

        // Check if file is mks_sql_ins wrapper
        if (text.includes("select mks_sql_ins(")) {
            if (typeof MksSqlInsParser !== 'undefined') {
                const parser = new MksSqlInsParser();
                const parsed = parser.parse(text);
                if (parsed) {
                    if (parsed.sql) {
                        document.getElementById('sql-text').value = parsed.sql;
                    }
                    if (parsed.info && parsed.info.default !== undefined) {
                        let val = parsed.info.default;
                        if (typeof val === 'object') {
                            document.getElementById('json-input').value = JSON.stringify(val, null, 2);
                        } else {
                            document.getElementById('json-input').value = val;
                        }
                        validateJSON();
                    }
                    return;
                }
            } else {
                console.warn("MksSqlInsParser not loaded");
            }
        }

        const target = document.getElementById(targetId);
        target.value = text;

        // If JSON input, trigger validation
        if (targetId === 'json-input') {
            validateJSON();
        }
    };
    reader.readAsText(file);
}

// 2. JSON Validation
function validateJSON() {
    const input = document.getElementById('json-input');
    const val = input.value.trim();
    if (!val) {
        input.classList.remove('invalid-json');
        return;
    }
    try {
        JSON.parse(val);
        input.classList.remove('invalid-json');
    } catch (e) {
        input.classList.add('invalid-json');
    }
}

function copyToClipboard(elementId) {
    const el = document.getElementById(elementId);
    if (!el) return;
    let text = "";
    if (el.tagName === 'TEXTAREA' || el.tagName === 'INPUT') {
        text = el.value;
    } else {
        text = el.textContent;
    }

    navigator.clipboard.writeText(text).then(() => {
        // Show feedback?
        // Simple console log for now or change icon briefly
        console.log("Copied to clipboard");
    }).catch(err => {
        console.error('Failed to copy: ', err);
    });
}

function downloadResult() {
    const text = document.getElementById('result-output').textContent;
    const blob = new Blob([text], { type: 'text/plain' });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = "result.sql";
    document.body.appendChild(a);
    a.click();
    window.URL.revokeObjectURL(url);
    document.body.removeChild(a);
}

function saveState() {
    // Optional: save to localStorage?
    // Not strictly required but good for UX
}
