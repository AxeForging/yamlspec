package services

import (
	"encoding/json"
	"html"
	"strings"

	"github.com/AxeForging/yamlspec/domain"
)

// HTMLFormatter outputs a self-contained interactive report.
type HTMLFormatter struct{}

// NewHTMLFormatter creates a new HTMLFormatter.
func NewHTMLFormatter() *HTMLFormatter {
	return &HTMLFormatter{}
}

func (f *HTMLFormatter) Format(result *domain.SuiteResult) ([]byte, error) {
	payload, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	var title string
	if result.Summary.Success {
		title = "yamlspec standards report - passed"
	} else {
		title = "yamlspec standards report - failed"
	}

	page := strings.Replace(htmlReportTemplate, "{{TITLE}}", html.EscapeString(title), 1)
	page = strings.Replace(page, "{{DATA}}", string(payload), 1)
	page = strings.Replace(page, "{{GENERATED}}", html.EscapeString(result.Summary.Timestamp.Format("2006-01-02 15:04:05 MST")), 1)

	return []byte(page), nil
}

const htmlReportTemplate = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{TITLE}}</title>
<style>
:root {
  color-scheme: dark;
  --ink: oklch(97% 0.006 255);
  --muted: oklch(67% 0.04 260);
  --dim: oklch(48% 0.035 260);
  --line: oklch(25% 0.035 260);
  --paper: oklch(12% 0.035 263);
  --paper-deep: oklch(8% 0.035 263);
  --panel: oklch(18% 0.035 260);
  --panel-strong: oklch(22% 0.04 260);
  --steel: oklch(77% 0.04 260);
  --blue: oklch(61% 0.18 255);
  --orange: oklch(67% 0.2 39);
  --green: oklch(74% 0.16 154);
  --red: oklch(68% 0.2 30);
  --amber: oklch(78% 0.16 78);
  --space-xs: 4px;
  --space-sm: 8px;
  --space-md: 12px;
  --space-lg: 16px;
  --space-xl: 24px;
  --space-2xl: 32px;
  --space-3xl: 48px;
  --radius: 6px;
}
* { box-sizing: border-box; }
body {
  margin: 0;
  background:
    radial-gradient(circle at 15% 8%, color-mix(in oklch, var(--orange) 10%, transparent), transparent 28%),
    radial-gradient(circle at 82% 18%, color-mix(in oklch, var(--blue) 9%, transparent), transparent 32%),
    linear-gradient(180deg, var(--paper-deep), var(--paper));
  color: var(--ink);
  font: 15px/1.55 ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
}
button, input, select { font: inherit; }
.shell { min-height: 100vh; }
.mast {
  padding: clamp(24px, 4vw, 44px) clamp(16px, 4vw, 44px) var(--space-2xl);
  border-bottom: 1px solid color-mix(in oklch, var(--line), transparent 30%);
  background:
    linear-gradient(135deg, color-mix(in oklch, var(--orange) 8%, transparent), transparent 44%),
    color-mix(in oklch, var(--paper-deep) 82%, transparent);
}
.mast-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.4fr) minmax(280px, 0.6fr);
  gap: var(--space-2xl);
  align-items: end;
  max-width: 1240px;
  margin: 0 auto;
}
.eyebrow {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 7px 12px;
  border: 1px solid var(--line);
  border-radius: 999px;
  background: var(--panel);
  color: var(--steel);
  font-size: 12px;
  font-weight: 800;
  letter-spacing: .12em;
  text-transform: uppercase;
}
.eyebrow::before {
  content: "";
  width: 8px;
  height: 8px;
  border-radius: 999px;
  background: var(--orange);
  box-shadow: 0 0 16px color-mix(in oklch, var(--orange), transparent 25%);
}
h1 {
  margin: var(--space-sm) 0 0;
  max-width: 760px;
  font-size: clamp(34px, 5vw, 60px);
  line-height: 1;
  letter-spacing: 0;
}
.subtitle {
  max-width: 70ch;
  margin: var(--space-lg) 0 0;
  color: var(--muted);
  font-size: 16px;
}
.summary-card {
  display: grid;
  gap: var(--space-lg);
  padding: var(--space-xl);
  border: 1px solid var(--line);
  border-radius: 14px;
  background: color-mix(in oklch, var(--panel) 86%, transparent);
  box-shadow: 0 24px 80px color-mix(in oklch, var(--paper-deep), transparent 24%);
}
.verdict {
  display: inline-flex;
  width: fit-content;
  align-items: center;
  gap: var(--space-sm);
  padding: 6px 10px;
  border: 1px solid currentColor;
  border-radius: 999px;
  color: var(--green);
  font-size: 12px;
  font-weight: 800;
  letter-spacing: .08em;
  text-transform: uppercase;
}
.verdict.fail { color: var(--red); }
.metrics {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-md);
}
.metric {
  padding: var(--space-md);
  border: 1px solid var(--line);
  border-radius: var(--radius);
  background: var(--paper-deep);
}
.metric strong {
  display: block;
  font-size: 26px;
  line-height: 1;
}
.metric span {
  color: var(--muted);
  font-size: 12px;
  font-weight: 700;
  letter-spacing: .06em;
  text-transform: uppercase;
}
.content {
  max-width: 1240px;
  margin: 0 auto;
  padding: var(--space-2xl) clamp(16px, 4vw, 44px) var(--space-3xl);
}
.toolbar {
  position: sticky;
  top: 0;
  z-index: 5;
  display: grid;
  grid-template-columns: minmax(220px, 1fr) repeat(2, minmax(150px, 210px)) auto auto auto;
  gap: var(--space-md);
  align-items: center;
  padding: var(--space-lg) 0;
  background: color-mix(in oklch, var(--paper-deep) 86%, transparent);
  backdrop-filter: blur(14px);
}
.control {
  width: 100%;
  min-height: 42px;
  border: 1px solid var(--line);
  border-radius: var(--radius);
  background: var(--panel);
  color: var(--ink);
  padding: 0 12px;
}
.button {
  min-height: 42px;
  border: 1px solid color-mix(in oklch, var(--orange), var(--ink) 12%);
  border-radius: var(--radius);
  background: var(--orange);
  color: var(--ink);
  padding: 0 14px;
  cursor: pointer;
  font-weight: 800;
}
.button.secondary {
  border-color: var(--line);
  background: var(--panel);
  color: var(--steel);
}
.failure-panel {
  display: none;
  gap: var(--space-md);
  margin-bottom: var(--space-lg);
  padding: var(--space-lg);
  border: 1px solid color-mix(in oklch, var(--red), var(--line) 55%);
  border-radius: 16px;
  background:
    linear-gradient(135deg, color-mix(in oklch, var(--red) 10%, transparent), transparent 44%),
    var(--paper-deep);
}
.failure-panel.visible { display: grid; }
.failure-panel h2 {
  margin: 0;
  font-size: 16px;
}
.failure-list {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: var(--space-sm);
}
.failure-item {
  display: grid;
  gap: 4px;
  padding: var(--space-md);
  border: 1px solid var(--line);
  border-radius: 10px;
  background: var(--panel);
  color: var(--ink);
  text-align: left;
  cursor: pointer;
}
.failure-item strong { color: var(--red); }
.failure-item span {
  color: var(--muted);
  font-size: 12px;
  overflow-wrap: anywhere;
}
.results {
  display: grid;
  gap: var(--space-lg);
}
.spec {
  border: 1px solid var(--line);
  border-radius: 16px;
  background: var(--panel);
  overflow: clip;
}
.spec[open] { background: color-mix(in oklch, var(--panel) 84%, var(--paper-deep)); }
.spec > summary {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) minmax(120px, auto);
  gap: var(--space-md);
  align-items: center;
  padding: var(--space-lg);
  cursor: pointer;
}
.status {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 66px;
  padding: 4px 8px;
  border-radius: 999px;
  border: 1px solid currentColor;
  font-size: 11px;
  font-weight: 900;
  letter-spacing: .08em;
  text-transform: uppercase;
  align-self: start;
}
.passed { color: var(--green); }
.failed, .error { color: var(--red); }
.skipped { color: var(--amber); }
.spec-title { min-width: 0; }
.spec-title strong { display: block; font-size: 18px; }
.tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 6px;
}
.tag {
  padding: 2px 7px;
  border: 1px solid var(--line);
  border-radius: 999px;
  color: var(--steel);
  font-size: 12px;
  background: var(--paper-deep);
}
.duration {
  color: var(--muted);
  font-size: 13px;
  text-align: right;
}
.spec-body {
  display: grid;
  gap: var(--space-lg);
  padding: 0 var(--space-lg) var(--space-lg);
}
.describe {
  display: grid;
  gap: var(--space-md);
  padding: var(--space-lg);
  border: 1px solid var(--line);
  border-radius: 12px;
  background: var(--paper-deep);
}
.describe h2 {
  margin: 0;
  font-size: 15px;
}
.selector {
  width: fit-content;
  max-width: 100%;
  overflow-wrap: anywhere;
  padding: 5px 8px;
  border-radius: var(--radius);
  background: var(--panel-strong);
  color: var(--steel);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
}
.assertions {
  display: grid;
  gap: var(--space-sm);
}
.assertion {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: var(--space-md);
  padding: var(--space-md);
  border: 1px solid var(--line);
  border-radius: var(--radius);
  background: var(--panel);
}
.assertion.failed, .assertion.error {
  border-color: color-mix(in oklch, var(--red), var(--line) 55%);
  background: color-mix(in oklch, var(--red) 7%, var(--panel));
}
.assertion-main { min-width: 0; }
.assertion-title { font-weight: 700; }
.assertion-meta {
  margin-top: 4px;
  color: var(--muted);
  font-size: 13px;
  overflow-wrap: anywhere;
}
.error-box {
  margin-top: var(--space-sm);
  padding: var(--space-md);
  border: 1px solid color-mix(in oklch, var(--red), var(--line) 55%);
  border-radius: var(--radius);
  background: color-mix(in oklch, var(--red) 12%, var(--paper-deep));
  color: var(--red);
  overflow-wrap: anywhere;
}
.evidence-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: var(--space-md);
}
.evidence {
  border: 1px solid var(--line);
  border-radius: 12px;
  background: var(--paper-deep);
  overflow: clip;
}
.evidence-group {
  display: grid;
  gap: var(--space-md);
}
.evidence-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-md);
}
.evidence-heading h2 {
  margin: 0;
  font-size: 15px;
}
.evidence-heading span {
  color: var(--muted);
  font-size: 12px;
}
.evidence summary {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-md);
  padding: var(--space-md) var(--space-lg);
  color: var(--ink);
  cursor: pointer;
  font-weight: 800;
}
.evidence summary span {
  min-width: 0;
  color: var(--muted);
  font-size: 12px;
  font-weight: 700;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.code {
  margin: 0;
  max-height: 460px;
  overflow: auto;
  padding: var(--space-lg);
  border-top: 1px solid var(--line);
  color: var(--steel);
  background: oklch(7.5% 0.035 263);
  font: 12px/1.6 ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  tab-size: 2;
  white-space: pre;
}
.empty {
  padding: var(--space-2xl);
  border: 1px dashed var(--line);
  border-radius: var(--radius);
  color: var(--muted);
  text-align: center;
}
.footer {
  margin-top: var(--space-2xl);
  color: var(--muted);
  font-size: 13px;
}
@media (max-width: 860px) {
  .mast-grid, .toolbar { grid-template-columns: 1fr; }
  .spec > summary { grid-template-columns: 1fr; }
  .metrics { grid-template-columns: 1fr 1fr; }
  .duration { text-align: left; }
}
</style>
</head>
<body>
<main class="shell">
  <header class="mast">
    <div class="mast-grid">
      <section>
        <div class="eyebrow">AxeForge / yamlspec</div>
        <h1>Standards report for infrastructure confidence.</h1>
        <p class="subtitle">A portable CI artifact for reviewing rendered YAML, specs, selectors, assertions, and failure details in one place.</p>
      </section>
      <aside class="summary-card" id="summary"></aside>
    </div>
  </header>
  <section class="content">
    <div class="toolbar">
      <input class="control" id="search" type="search" placeholder="Search specs, selectors, assertions, errors">
      <select class="control" id="statusFilter" aria-label="Filter by status">
        <option value="all">All statuses</option>
        <option value="failed">Failed</option>
        <option value="error">Errors</option>
        <option value="passed">Passed</option>
        <option value="skipped">Skipped</option>
      </select>
      <select class="control" id="tagFilter" aria-label="Filter by tag">
        <option value="all">All tags</option>
      </select>
      <button class="button" id="failedOnly" type="button">Failures only</button>
      <button class="button secondary" id="expandAll" type="button">Expand all</button>
      <button class="button secondary" id="collapseAll" type="button">Collapse all</button>
    </div>
    <section class="failure-panel" id="failurePanel"></section>
    <section class="results" id="results"></section>
    <div class="footer">Generated {{GENERATED}}. Single-file report: HTML, CSS, JavaScript, and run data are bundled in this document.</div>
  </section>
</main>
<script id="yamlspec-data" type="application/json">{{DATA}}</script>
<script>
(function () {
  var data = JSON.parse(document.getElementById("yamlspec-data").textContent);
  var search = document.getElementById("search");
  var statusFilter = document.getElementById("statusFilter");
  var tagFilter = document.getElementById("tagFilter");
  var results = document.getElementById("results");
  var failurePanel = document.getElementById("failurePanel");
  var failedOnly = document.getElementById("failedOnly");
  var expandAll = document.getElementById("expandAll");
  var collapseAll = document.getElementById("collapseAll");

  function fmtDuration(ns) {
    var seconds = Number(ns || 0) / 1000000000;
    if (seconds < 1) return Math.round(seconds * 1000) + "ms";
    return seconds.toFixed(2) + "s";
  }

  function text(value) {
    if (value === null || value === undefined) return "";
    if (typeof value === "string") return value;
    return JSON.stringify(value);
  }

  function append(parent, tag, className, value) {
    var el = document.createElement(tag);
    if (className) el.className = className;
    if (value !== undefined) el.textContent = value;
    parent.appendChild(el);
    return el;
  }

  function shortPath(path) {
    if (!path) return "";
    var normalized = String(path).replace(/\\/g, "/");
    var marker = "/integration/";
    var markerIndex = normalized.indexOf(marker);
    if (markerIndex >= 0) return normalized.slice(markerIndex + 1);
    var parts = normalized.split("/").filter(Boolean);
    if (parts.length <= 4) return normalized;
    return "..." + parts.slice(-4).join("/");
  }

  function countAssertions(spec) {
    var total = 0;
    (spec.describes || []).forEach(function (desc) {
      total += (desc.assertions || []).length;
    });
    return total;
  }

  function countFailures(spec) {
    var total = 0;
    (spec.describes || []).forEach(function (desc) {
      (desc.assertions || []).forEach(function (a) {
        if (a.status !== "passed") total += 1;
      });
    });
    if (spec.error) total += 1;
    return total;
  }

  function addEvidence(parent, title, detail, content) {
    if (!content) return;
    var box = append(parent, "details", "evidence");
    var summary = append(box, "summary", "");
    append(summary, "strong", "", title);
    if (detail) append(summary, "span", "", shortPath(detail));
    append(box, "pre", "code", content);
  }

  function searchable(spec) {
    var parts = [spec.name, spec.status, spec.error, spec.source_file, spec.source_content, (spec.tags || []).join(" ")];
    (spec.manifests || []).forEach(function (manifest) {
      parts.push(manifest.path, manifest.content, String(manifest.documents || ""));
    });
    (spec.describes || []).forEach(function (desc) {
      parts.push(desc.name, desc.select);
      (desc.assertions || []).forEach(function (a) {
        parts.push(a.should, a.status, a.error, text(a.expected), text(a.actual), String(a.source_line || ""));
      });
    });
    return parts.join(" ").toLowerCase();
  }

  function populateSummary() {
    var s = data.summary || {};
    var root = document.getElementById("summary");
    root.innerHTML = "";
    append(root, "div", "verdict " + (s.success ? "" : "fail"), s.success ? "Passed" : "Needs attention");
    var metrics = append(root, "div", "metrics");
    [["Specs", s.total_specs], ["Assertions", s.total_assertions], ["Failed", s.failed_assertions], ["Duration", fmtDuration(s.duration)]].forEach(function (item) {
      var metric = append(metrics, "div", "metric");
      append(metric, "strong", "", String(item[1] || 0));
      append(metric, "span", "", item[0]);
    });
  }

  function renderFailurePanel() {
    var failures = [];
    (data.specs || []).forEach(function (spec, specIndex) {
      (spec.describes || []).forEach(function (desc) {
        (desc.assertions || []).forEach(function (a) {
          if (a.status !== "passed") {
            failures.push({ spec: spec, specIndex: specIndex, desc: desc, assertion: a });
          }
        });
      });
      if (spec.error) {
        failures.push({ spec: spec, specIndex: specIndex, desc: null, assertion: { should: "spec execution", error: spec.error } });
      }
    });

    failurePanel.innerHTML = "";
    failurePanel.className = failures.length ? "failure-panel visible" : "failure-panel";
    if (!failures.length) return;
    append(failurePanel, "h2", "", "Failure focus");
    var list = append(failurePanel, "div", "failure-list");
    failures.slice(0, 12).forEach(function (failure) {
      var item = append(list, "button", "failure-item");
      item.type = "button";
      item.setAttribute("data-spec-id", "spec-" + failure.specIndex);
      append(item, "strong", "", failure.spec.name + " / " + failure.assertion.should);
      append(item, "span", "", failure.assertion.error || "Open spec for details");
      if (failure.assertion.source_line) append(item, "span", "", shortPath(failure.spec.source_file) + ":" + failure.assertion.source_line);
    });
    if (failures.length > 12) {
      append(list, "div", "failure-item", String(failures.length - 12) + " more failures. Use the Failed filter to review all.");
    }
  }

  function populateTags() {
    var seen = {};
    (data.specs || []).forEach(function (spec) {
      (spec.tags || []).forEach(function (tag) { seen[tag] = true; });
    });
    Object.keys(seen).sort().forEach(function (tag) {
      var option = document.createElement("option");
      option.value = tag;
      option.textContent = tag;
      tagFilter.appendChild(option);
    });
  }

  function render() {
    var q = search.value.trim().toLowerCase();
    var status = statusFilter.value;
    var tag = tagFilter.value;
    var specs = (data.specs || []).filter(function (spec) {
      if (status !== "all" && spec.status !== status) return false;
      if (tag !== "all" && (spec.tags || []).indexOf(tag) === -1) return false;
      if (q && searchable(spec).indexOf(q) === -1) return false;
      return true;
    });

    results.innerHTML = "";
    if (specs.length === 0) {
      append(results, "div", "empty", "No specs match the current filters.");
      return;
    }

    specs.forEach(function (spec) {
      var specIndex = (data.specs || []).indexOf(spec);
      var details = append(results, "details", "spec");
      details.id = "spec-" + specIndex;
      if (spec.status !== "passed") details.open = true;
      var summary = append(details, "summary", "");
      append(summary, "span", "status " + spec.status, spec.status);
      var title = append(summary, "span", "spec-title");
      append(title, "strong", "", spec.name || "Unnamed spec");
      if (spec.tags && spec.tags.length) {
        var tags = append(title, "span", "tags");
        spec.tags.forEach(function (tagName) { append(tags, "span", "tag", tagName); });
      }
      var failed = countFailures(spec);
      var total = countAssertions(spec);
      append(summary, "span", "duration", (failed ? failed + " failed / " : "") + total + " assertions · " + fmtDuration(spec.duration));

      var body = append(details, "div", "spec-body");
      if (spec.source_file) append(body, "div", "selector", shortPath(spec.source_file));
      if (spec.error) {
        append(body, "div", "error-box", spec.error);
      }

      (spec.describes || []).forEach(function (desc) {
        var block = append(body, "section", "describe");
        append(block, "h2", "", desc.name || "Describe");
        if (desc.select) append(block, "div", "selector", "select: " + desc.select);
        var assertions = append(block, "div", "assertions");
        (desc.assertions || []).forEach(function (a) {
          var row = append(assertions, "article", "assertion " + a.status);
          append(row, "span", "status " + a.status, a.status);
          var main = append(row, "div", "assertion-main");
          append(main, "div", "assertion-title", a.should || "Assertion");
          var meta = [];
          if (a.source_line) meta.push("line " + a.source_line);
          if (a.actual !== undefined) meta.push("actual: " + text(a.actual));
          if (a.expected !== undefined) meta.push("expected: " + text(a.expected));
          if (meta.length) append(main, "div", "assertion-meta", meta.join(" | "));
          if (a.error) append(main, "div", "error-box", a.error);
        });
      });

      var evidenceGroup = append(body, "section", "evidence-group");
      var evidenceHeading = append(evidenceGroup, "div", "evidence-heading");
      append(evidenceHeading, "h2", "", "Evidence");
      append(evidenceHeading, "span", "", "Original spec and rendered YAML used for this run");
      var evidence = append(evidenceGroup, "div", "evidence-grid");
      addEvidence(evidence, "Spec YAML", spec.source_file || "spec.yaml", spec.source_content);
      (spec.manifests || []).forEach(function (manifest, index) {
        var label = manifest.path || ("manifest " + (index + 1));
        var docs = manifest.documents ? " · " + manifest.documents + " document" + (manifest.documents === 1 ? "" : "s") : "";
        addEvidence(evidence, "Rendered manifest", label + docs, manifest.content);
      });
    });
  }

  failedOnly.addEventListener("click", function () {
    statusFilter.value = "failed";
    tagFilter.value = "all";
    search.value = "";
    render();
  });
  expandAll.addEventListener("click", function () {
    document.querySelectorAll(".spec").forEach(function (el) { el.open = true; });
  });
  collapseAll.addEventListener("click", function () {
    document.querySelectorAll(".spec").forEach(function (el) { el.open = false; });
  });
  failurePanel.addEventListener("click", function (event) {
    var item = event.target.closest("[data-spec-id]");
    if (!item) return;
    var target = document.getElementById(item.getAttribute("data-spec-id"));
    if (!target) return;
    target.open = true;
    target.scrollIntoView({ behavior: "smooth", block: "start" });
  });
  [search, statusFilter, tagFilter].forEach(function (el) { el.addEventListener("input", render); });

  populateSummary();
  populateTags();
  renderFailurePanel();
  render();
}());
</script>
</body>
</html>`
