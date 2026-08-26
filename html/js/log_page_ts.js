"use strict";
var LOG_VIEW_STATE = {
    search: "",
    severities: { DEBUG: false, WARNING: false, ERROR: false },
};
function logTextForDisplay(entry) {
    return String(entry == undefined ? "" : entry).replace(/&nbsp;/g, " ");
}
function logEntrySeverity(entry) {
    var text = String(entry || "");
    if (text.indexOf("[ERROR]") != -1) {
        return "ERROR";
    }
    if (text.indexOf("[WARNING]") != -1) {
        return "WARNING";
    }
    if (text.indexOf("[DEBUG]") != -1) {
        return "DEBUG";
    }
    return "";
}
function logEntryLayer(entry) {
    var match = String(entry || "").match(/(?:\[Threadfin\]|\[DEBUG\])\s+(?:\[(?:WARNING|ERROR)\]\s+)?([A-Za-z][A-Za-z0-9 /_-]{0,39}):(?:&nbsp;|\s)/);
    return match ? match[1].trim() : "";
}
function logEntryMatches(entry) {
    var search = LOG_VIEW_STATE.search.trim().toLocaleLowerCase();
    if (search && logTextForDisplay(entry).toLocaleLowerCase().indexOf(search) == -1) {
        return false;
    }
    var active = Object.keys(LOG_VIEW_STATE.severities).filter(marker => LOG_VIEW_STATE.severities[marker]);
    return active.length == 0 || active.indexOf(logEntrySeverity(entry)) != -1;
}
function createLogEntryElement(entry) {
    var text = logTextForDisplay(entry);
    var element = document.createElement("article");
    element.className = "tf-log-entry";
    var severity = logEntrySeverity(entry);
    if (severity) {
        element.setAttribute("data-severity", severity);
    }
    var layer = logEntryLayer(entry);
    if (layer) {
        var layerLabel = document.createElement("span");
        layerLabel.className = "tf-log-layer";
        layerLabel.textContent = layer;
        element.appendChild(layerLabel);
    }
    var pre = document.createElement("pre");
    pre.textContent = text;
    element.appendChild(pre);
    element.queryText = text;
    return element;
}
function renderLogEntryList(host, entries) {
    var visible = (entries || []).filter(logEntryMatches).map(createLogEntryElement);
    host.replaceChildren.apply(host, visible);
    if (visible.length == 0) {
        var empty = document.createElement("p");
        empty.className = "tf-log-empty";
        empty.textContent = "No log entries match the current filters.";
        host.appendChild(empty);
    }
}
function renderLogPage(doc) {
    doc.innerHTML = "";
    var root = document.createElement("div");
    root.id = "content_log";
    root.className = "tf-log";
    var header = document.createElement("header");
    header.className = "tf-admin-header";
    var headingGroup = document.createElement("div");
    var heading = document.createElement("h1");
    heading.textContent = "Log";
    var purpose = document.createElement("p");
    purpose.textContent = "Search Threadfin's current diagnostic record or limit it to response-backed DEBUG, WARNING, and ERROR markers.";
    headingGroup.appendChild(heading);
    headingGroup.appendChild(purpose);
    header.appendChild(headingGroup);
    root.appendChild(header);
    var controls = document.createElement("div");
    controls.className = "tf-log-controls";
    var searchLabel = document.createElement("label");
    searchLabel.textContent = "Search log";
    var search = document.createElement("input");
    search.type = "search";
    search.id = "log-search";
    search.value = LOG_VIEW_STATE.search;
    search.autocomplete = "off";
    search.addEventListener("input", () => {
        LOG_VIEW_STATE.search = search.value;
        showLogs(false);
    });
    searchLabel.appendChild(search);
    controls.appendChild(searchLabel);
    var severityGroup = document.createElement("fieldset");
    var legend = document.createElement("legend");
    legend.textContent = "Severity markers";
    severityGroup.appendChild(legend);
    ["DEBUG", "WARNING", "ERROR"].forEach(marker => {
        var label = document.createElement("label");
        var input = document.createElement("input");
        input.type = "checkbox";
        input.checked = LOG_VIEW_STATE.severities[marker];
        input.addEventListener("change", () => {
            LOG_VIEW_STATE.severities[marker] = input.checked;
            showLogs(false);
        });
        label.appendChild(input);
        label.appendChild(document.createTextNode(marker));
        severityGroup.appendChild(label);
    });
    controls.appendChild(severityGroup);
    root.appendChild(controls);
    var wrapper = document.createElement("div");
    wrapper.id = "box-wrapper";
    wrapper.className = "tf-log-scroll";
    var entries = document.createElement("div");
    entries.id = "content_log_entries";
    entries.className = "tf-log-entries";
    wrapper.appendChild(entries);
    root.appendChild(wrapper);
    var danger = document.createElement("section");
    danger.className = "tf-log-danger-zone";
    var dangerHeading = document.createElement("h2");
    dangerHeading.textContent = "Reset log";
    var dangerText = document.createElement("p");
    dangerText.textContent = "Permanently remove all current in-memory log entries. Filtering does not remove entries.";
    var reset = document.createElement("button");
    reset.type = "button";
    reset.className = "tf-destructive-action";
    reset.textContent = "Reset logs…";
    reset.addEventListener("click", resetLogs);
    danger.appendChild(dangerHeading);
    danger.appendChild(dangerText);
    danger.appendChild(reset);
    root.appendChild(danger);
    doc.appendChild(root);
    showLogs(true);
}
