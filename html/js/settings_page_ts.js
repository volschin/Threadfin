"use strict";
var SETTINGS_SECTIONS = [
    {
        key: "general",
        label: "General",
        settings: ["epgSource", "ThreadfinAutoUpdate", "ssdp", "tuner", "epgCategories", "epgCategoriesColors", "dummy", "dummyChannel", "ignoreFilters", "api"],
    },
    {
        key: "files",
        label: "Files",
        settings: ["update", "files.update", "temp.path", "cache.images", "bindIpAddress", "httpThreadfinDomain", "forceHttps", "excludeStreamHttps", "httpsPort", "httpsThreadfinDomain", "xepg.replace.missing.images", "xepg.replace.channel.title", "enableNonAscii"],
    },
    {
        key: "streaming",
        label: "Streaming",
        settings: ["udpxy", "buffer.size.kb", "buffer.timeout", "user.agent", "ffmpeg.path", "ffmpeg.options", "ffmpeg.forceHttp", "vlc.path", "vlc.options"],
    },
    { key: "backup", label: "Backup", settings: ["backup.path", "backup.keep"] },
    {
        key: "authentication",
        label: "Authentication",
        settings: ["authentication.web", "authentication.pms", "authentication.m3u", "authentication.xml", "authentication.api"],
    },
];
var SETTINGS_NUMERIC_KEYS = {
    "tuner": true,
    "backup.keep": true,
    "buffer.size.kb": true,
    "buffer.timeout": true,
    "httpsPort": true,
};
function serializeSettingsChanges(controls) {
    var changes = {};
    controls.forEach(control => {
        var classes = String(control.className || "").split(/\s+/);
        if (classes.indexOf("changed") == -1 || !control.name) {
            return;
        }
        var name = String(control.name);
        if (String(control.tagName).toUpperCase() == "INPUT" && control.type == "checkbox") {
            changes[name] = Boolean(control.checked);
            return;
        }
        var value = String(control.value == undefined ? "" : control.value);
        if (name == "update") {
            changes[name] = value.split(",").filter(entry => entry.length > 0);
            return;
        }
        if (SETTINGS_NUMERIC_KEYS[name]) {
            changes[name] = name == "buffer.timeout" ? parseFloat(value) : parseInt(value, 10);
            return;
        }
        changes[name] = value;
    });
    return changes;
}
function settingsEPGImpact(value) {
    if (value == "XEPG") {
        return "XEPG uses XMLTV guide sources and makes Mapping plus M3U and XMLTV outputs applicable after you save.";
    }
    return "PMS lets the connected client manage guide data; XMLTV sources, Mapping, and M3U/XMLTV guide outputs are not used by this mode.";
}
function renderEPGSourcePreview(value, target) {
    if (target) {
        target.textContent = settingsEPGImpact(value);
    }
}
function activateSettingsSection(sectionKey, focusActiveTab) {
    var root = document.querySelector(".tf-settings");
    if (!root) {
        return;
    }
    var panels = root.querySelectorAll("[data-settings-panel]");
    for (var i = 0; i < panels.length; i++) {
        var panel = panels[i];
        panel.hidden = panel.getAttribute("data-settings-panel") != sectionKey;
    }
    var tabs = root.querySelectorAll("[data-settings-tab]");
    for (var j = 0; j < tabs.length; j++) {
        var tab = tabs[j];
        var active = tab.getAttribute("data-settings-tab") == sectionKey;
        tab.setAttribute("aria-selected", active ? "true" : "false");
        tab.setAttribute("tabindex", active ? "0" : "-1");
        if (active && focusActiveTab) {
            tab.focus();
        }
    }
}
function handleSettingsTabKeydown(event, currentIndex) {
    var nextIndex = currentIndex;
    switch (event.key) {
        case "ArrowRight":
        case "ArrowDown":
            nextIndex = (currentIndex + 1) % SETTINGS_SECTIONS.length;
            break;
        case "ArrowLeft":
        case "ArrowUp":
            nextIndex = (currentIndex + SETTINGS_SECTIONS.length - 1) % SETTINGS_SECTIONS.length;
            break;
        case "Home":
            nextIndex = 0;
            break;
        case "End":
            nextIndex = SETTINGS_SECTIONS.length - 1;
            break;
        default:
            return;
    }
    event.preventDefault();
    activateSettingsSection(SETTINGS_SECTIONS[nextIndex].key, true);
}
function syncSettingsAuthenticationGating(root) {
    var web = root.querySelector('[name="authentication.web"]');
    var explanation = root.querySelector(".tf-settings-auth-gating");
    var enabled = web ? web.checked : false;
    if (explanation) {
        explanation.textContent = enabled
            ? "WEB access is enabled, so PMS, M3U, XML, and API permissions can be configured."
            : "WEB access gates user management. Saving WEB as disabled clears PMS, M3U, XML, and API authentication on the server.";
    }
    ;
    ["authentication.pms", "authentication.m3u", "authentication.xml", "authentication.api"].forEach(key => {
        var input = root.querySelector('[name="' + key + '"]');
        if (input) {
            input.disabled = !enabled;
        }
    });
}
function labelSettingsControls(root) {
    var controls = root.querySelectorAll("input, select, textarea");
    for (var index = 0; index < controls.length; index++) {
        var control = controls[index];
        if (String(control.type || "").toLowerCase() == "hidden") {
            continue;
        }
        var row = control.closest("tr");
        var labelCell = row ? row.querySelector("td:first-child") : null;
        if (!labelCell || !labelCell.textContent) {
            continue;
        }
        var key = String(control.name || "setting").replace(/[^A-Za-z0-9_-]+/g, "-");
        labelCell.id = "settings-field-label-" + key + "-" + index;
        control.id = "settings-field-" + key + "-" + index;
        control.setAttribute("aria-labelledby", labelCell.id);
    }
}
function renderSettingsPage(doc) {
    doc.innerHTML = "";
    var root = document.createElement("div");
    root.className = "tf-settings";
    var header = document.createElement("header");
    header.className = "tf-admin-header";
    var headingGroup = document.createElement("div");
    var heading = document.createElement("h1");
    heading.textContent = "Settings";
    var purpose = document.createElement("p");
    purpose.textContent = "Configure Threadfin ingestion, files, streaming, backups, and access without changing unrelated settings.";
    headingGroup.appendChild(heading);
    headingGroup.appendChild(purpose);
    var save = document.createElement("button");
    save.type = "button";
    save.className = "tf-primary-action";
    save.textContent = "Save settings";
    save.addEventListener("click", saveSettings);
    header.appendChild(headingGroup);
    header.appendChild(save);
    root.appendChild(header);
    var nav = document.createElement("div");
    nav.className = "tf-settings-nav";
    nav.setAttribute("role", "tablist");
    nav.setAttribute("aria-label", "Settings sections");
    SETTINGS_SECTIONS.forEach((definition, index) => {
        var button = document.createElement("button");
        button.type = "button";
        button.textContent = definition.label;
        button.setAttribute("role", "tab");
        button.id = "settings-tab-" + definition.key;
        button.setAttribute("data-settings-tab", definition.key);
        button.setAttribute("aria-controls", "settings-panel-" + definition.key);
        button.setAttribute("aria-selected", index == 0 ? "true" : "false");
        button.setAttribute("tabindex", index == 0 ? "0" : "-1");
        button.addEventListener("click", () => activateSettingsSection(definition.key));
        button.addEventListener("keydown", event => handleSettingsTabKeydown(event, index));
        nav.appendChild(button);
    });
    root.appendChild(nav);
    var settingsHost = document.createElement("div");
    settingsHost.id = "content_settings";
    settingsHost.className = "tf-settings-panels";
    SETTINGS_SECTIONS.forEach((definition, index) => {
        var panel = document.createElement("section");
        panel.id = "settings-panel-" + definition.key;
        panel.className = "tf-settings-panel";
        panel.setAttribute("role", "tabpanel");
        panel.setAttribute("aria-labelledby", "settings-tab-" + definition.key);
        panel.setAttribute("data-settings-panel", definition.key);
        panel.hidden = index != 0;
        var panelHeading = document.createElement("h2");
        panelHeading.textContent = definition.label;
        panel.appendChild(panelHeading);
        if (definition.key == "authentication") {
            var gating = document.createElement("p");
            gating.className = "tf-settings-auth-gating";
            panel.appendChild(gating);
        }
        var table = document.createElement("table");
        table.className = "tf-settings-table";
        var renderer = new SettingsCategory();
        definition.settings.forEach(settingsKey => {
            var setting = renderer.createSettings(settingsKey);
            setting.setAttribute("data-setting-key", settingsKey);
            table.appendChild(setting);
            table.appendChild(renderer.createDescription(settingsKey));
        });
        panel.appendChild(table);
        if (definition.key == "general") {
            var preview = document.createElement("p");
            preview.className = "tf-settings-epg-preview";
            preview.setAttribute("aria-live", "polite");
            panel.insertBefore(preview, table.nextSibling);
            var epgSource = table.querySelector('[name="epgSource"]');
            if (epgSource) {
                renderEPGSourcePreview(epgSource.value, preview);
                epgSource.addEventListener("change", () => renderEPGSourcePreview(epgSource.value, preview));
            }
        }
        if (definition.key == "backup") {
            var backupActions = document.createElement("div");
            backupActions.className = "tf-backup-actions";
            var backupButton = document.createElement("button");
            backupButton.type = "button";
            backupButton.textContent = "Download backup";
            backupButton.addEventListener("click", backup);
            var restoreButton = document.createElement("button");
            restoreButton.type = "button";
            restoreButton.className = "tf-destructive-action";
            restoreButton.textContent = "Restore backup…";
            restoreButton.addEventListener("click", restore);
            backupActions.appendChild(backupButton);
            backupActions.appendChild(restoreButton);
            panel.appendChild(backupActions);
        }
        settingsHost.appendChild(panel);
    });
    root.appendChild(settingsHost);
    doc.appendChild(root);
    labelSettingsControls(root);
    var authWeb = root.querySelector('[name="authentication.web"]');
    if (authWeb) {
        authWeb.addEventListener("change", () => syncSettingsAuthenticationGating(root));
    }
    syncSettingsAuthenticationGating(root);
}
