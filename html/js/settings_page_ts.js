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
var settingsActiveSection = "general";
var settingsAdvancedOpen = {};
var settingsFeedback = { state: "idle", message: "" };
var SETTINGS_ADVANCED_KEYS = {
    general: ["epgCategories", "epgCategoriesColors", "dummy", "dummyChannel", "ignoreFilters", "api"],
    files: ["bindIpAddress", "httpThreadfinDomain", "forceHttps", "excludeStreamHttps", "httpsPort", "httpsThreadfinDomain", "enableNonAscii"],
    streaming: ["udpxy", "user.agent", "ffmpeg.options", "ffmpeg.forceHttp", "vlc.path", "vlc.options"],
};
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
            changes[name] = Number(value);
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
    settingsActiveSection = sectionKey;
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
        var field = control.closest(".tf-form-field");
        var labelCell = field ? field.querySelector("label") : null;
        if (!labelCell || !labelCell.textContent) {
            continue;
        }
        var key = String(control.name || "setting").replace(/[^A-Za-z0-9_-]+/g, "-");
        if (!labelCell.id)
            labelCell.id = "settings-field-label-" + key + "-" + index;
        if (!control.id)
            control.id = "settings-field-" + key + "-" + index;
        control.setAttribute("aria-labelledby", labelCell.id);
        labelCell.htmlFor = control.id;
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
    header.appendChild(headingGroup);
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
        button.setAttribute("aria-selected", definition.key == settingsActiveSection ? "true" : "false");
        button.setAttribute("tabindex", definition.key == settingsActiveSection ? "0" : "-1");
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
        panel.hidden = definition.key != settingsActiveSection;
        var panelHeading = document.createElement("h2");
        panelHeading.textContent = definition.label;
        panel.appendChild(panelHeading);
        if (definition.key == "authentication") {
            var gating = document.createElement("p");
            gating.className = "tf-settings-auth-gating";
            panel.appendChild(gating);
        }
        panelHeading.id = "settings-heading-" + definition.key;
        var fields = document.createElement("fieldset");
        fields.className = "tf-form-fields";
        fields.setAttribute("aria-labelledby", panelHeading.id);
        panel.appendChild(fields);
        var advancedKeys = SETTINGS_ADVANCED_KEYS[definition.key] || [];
        var advancedFields;
        if (advancedKeys.length) {
            var advanced = document.createElement("details");
            advanced.className = "tf-form-advanced";
            advanced.open = settingsAdvancedOpen[definition.key] === true;
            var summary = document.createElement("summary");
            summary.textContent = "Advanced " + definition.label.toLowerCase() + " settings";
            advanced.appendChild(summary);
            advancedFields = document.createElement("fieldset");
            advancedFields.className = "tf-form-fields";
            advancedFields.setAttribute("aria-label", summary.textContent);
            advanced.appendChild(advancedFields);
            advanced.addEventListener("toggle", () => { settingsAdvancedOpen[definition.key] = advanced.open; });
            panel.appendChild(advanced);
        }
        var renderer = new SettingsCategory();
        definition.settings.forEach(settingsKey => {
            var field = renderer.createSettings(settingsKey);
            field.setAttribute("data-setting-key", settingsKey);
            var help = renderer.createDescription(settingsKey);
            if (help.textContent.trim())
                appendUIFormHelp(field, help);
            var control = field.querySelector("input, select");
            var error = document.createElement("p");
            error.id = control.id + "-error";
            error.className = "tf-field-error";
            error.hidden = true;
            control.setAttribute("aria-describedby", [control.getAttribute("aria-describedby"), error.id].filter(Boolean).join(" "));
            field.appendChild(error);
            if (control.tagName == "INPUT" && SETTINGS_NUMERIC_KEYS[settingsKey]) {
                control.type = "number";
                control.step = settingsKey == "buffer.timeout" ? "any" : "1";
                control.inputMode = settingsKey == "buffer.timeout" ? "decimal" : "numeric";
                control.required = true;
            }
            control.addEventListener("input", () => clearSettingsFieldError(field, control));
            control.addEventListener("change", () => clearSettingsFieldError(field, control));
            (advancedKeys.indexOf(settingsKey) >= 0 ? advancedFields : fields).appendChild(field);
            if (settingsKey == "epgSource") {
                var preview = document.createElement("p");
                preview.className = "tf-settings-epg-preview";
                preview.setAttribute("aria-live", "polite");
                field.appendChild(preview);
                renderEPGSourcePreview(control.value, preview);
                control.addEventListener("change", () => renderEPGSourcePreview(control.value, preview));
            }
        });
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
    var footer = document.createElement("footer");
    footer.className = "tf-settings-save-bar";
    var status = document.createElement("p");
    status.id = "settings-form-status";
    status.setAttribute("role", "status");
    status.setAttribute("aria-live", "polite");
    var save = document.createElement("button");
    save.type = "button";
    save.className = "tf-primary-action";
    save.textContent = "Save settings";
    save.addEventListener("click", saveSettings);
    footer.appendChild(status);
    footer.appendChild(save);
    root.appendChild(footer);
    doc.appendChild(root);
    labelSettingsControls(root);
    var authWeb = root.querySelector('[name="authentication.web"]');
    if (authWeb) {
        authWeb.addEventListener("change", () => syncSettingsAuthenticationGating(root));
    }
    syncSettingsAuthenticationGating(root);
    renderSettingsFeedback();
}
function clearSettingsFieldError(field, control) {
    control.removeAttribute("aria-invalid");
    var error = field.querySelector(".tf-field-error");
    if (error) {
        error.textContent = "";
        error.hidden = true;
    }
    settingsFeedback = { state: "idle", message: "Unsaved changes." };
    renderSettingsFeedback();
}
function settingsInputError(name, value) {
    if (name == "update") {
        var times = value.split(",").map(time => time.replace(/ /g, "")).filter(Boolean);
        if (times.some(time => !/^(?:[01]\d|2[0-3])[0-5]\d$/.test(time))) {
            return "Use 24-hour times as HHMM, separated by commas (for example 0000,1200).";
        }
    }
    if (SETTINGS_NUMERIC_KEYS[name] && (!value.trim() || !Number.isFinite(Number(value)) ||
        (name != "buffer.timeout" && !Number.isInteger(Number(value))))) {
        return name == "buffer.timeout" ? "Enter a valid number." : "Enter a whole number.";
    }
    return "";
}
function validateSettingsForm(root) {
    var controls = root.querySelectorAll("input.changed, select.changed");
    for (var index = 0; index < controls.length; index++) {
        var control = controls[index];
        if (control.type == "checkbox" || control.disabled)
            continue;
        var message = settingsInputError(control.name, control.value);
        if (!message && !control.checkValidity())
            message = control.validationMessage;
        if (!message)
            continue;
        var field = control.closest(".tf-form-field");
        var error = field.querySelector(".tf-field-error");
        error.textContent = message;
        error.hidden = false;
        control.setAttribute("aria-invalid", "true");
        var panel = control.closest("[data-settings-panel]");
        activateSettingsSection(panel.getAttribute("data-settings-panel"));
        var details = control.closest("details");
        if (details)
            details.open = true;
        control.focus();
        return false;
    }
    return true;
}
function renderSettingsFeedback() {
    var root = document.querySelector(".tf-settings");
    if (!root)
        return;
    var status = root.querySelector("#settings-form-status");
    status.textContent = settingsFeedback.message;
    status.setAttribute("data-state", settingsFeedback.state);
    var pending = settingsFeedback.state == "progress";
    root.setAttribute("aria-busy", pending ? "true" : "false");
    root.querySelector(".tf-settings-save-bar button").disabled = pending;
    Array.prototype.forEach.call(root.querySelectorAll("fieldset"), (group) => { group.disabled = pending; });
}
function completeSettingsRequest(response, transportFailure = false) {
    settingsFeedback = response && response.status === true
        ? { state: "success", message: "Settings saved." }
        : { state: "error", message: transportFailure
                ? "Save result unclear. Reload Settings to check the saved values before retrying."
                : response && response.err ? String(response.err) : "Settings could not be saved. Your inputs are still available." };
    renderSettingsFeedback();
}
