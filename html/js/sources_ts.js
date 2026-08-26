"use strict";
var sourceFeedbackByKey = {};
var sourcePageFeedback = {};
var sourcePopupInvoker;
var sourcePopupFocusKey = "";
var sourcePopupFocusListenerAttached = false;
var sourcePopupShownListenerAttached = false;
function sourceRecord(value) {
    return value && typeof value == "object" && !Array.isArray(value) ? value : {};
}
function sourceString(value) {
    return typeof value == "string" ? value : value === undefined || value === null ? "" : String(value);
}
function sourceNumber(value) {
    var number = typeof value == "number" ? value : Number(value);
    return isFinite(number) ? number : 0;
}
function sourceLocationForDisplay(value) {
    var source = sourceString(value);
    if (!/^https?:/i.test(source.trim())) {
        return source;
    }
    try {
        var location = new URL(source.trim());
        if (!location.protocol || !location.host) {
            return "{{.sources.locationInvalidDisplay}}";
        }
        return location.protocol + "//" + location.host + location.pathname;
    }
    catch (_error) {
        return "{{.sources.locationInvalidDisplay}}";
    }
}
function sourceFeedbackKey(providerType, id) {
    return providerType + ":" + id;
}
function selectSourceList(server, destination) {
    var root = sourceRecord(server);
    var settings = sourceRecord(root.settings);
    var files = sourceRecord(settings.files);
    var types = destination == "playlist" ? ["m3u", "hdhr"] : ["xmltv"];
    var sources = [];
    types.forEach(providerType => {
        var configured = sourceRecord(files[providerType]);
        Object.keys(configured).forEach(id => {
            var source = sourceRecord(configured[id]);
            var compatibility = sourceRecord(source.compatibility);
            var rawAvailability = source["provider.availability"];
            var counts = [];
            if (providerType == "xmltv") {
                counts.push({ label: "{{.sources.counts.channels}}", value: sourceString(compatibility["xmltv.channels"] || 0) });
                counts.push({ label: "{{.sources.counts.programs}}", value: sourceString(compatibility["xmltv.programs"] || 0) });
            }
            else {
                counts.push({ label: "{{.sources.counts.streams}}", value: sourceString(compatibility.streams || 0) });
                counts.push({ label: "{{.sources.counts.tuners}}", value: sourceString(source.tuner || 0) });
                counts.push({ label: "{{.sources.counts.groupCoverage}}", value: sourceString(compatibility["group.title"] || 0) + "%" });
                counts.push({ label: "{{.sources.counts.tvgIDCoverage}}", value: sourceString(compatibility["tvg.id"] || 0) + "%" });
                counts.push({ label: "{{.sources.counts.uniqueIDCoverage}}", value: sourceString(compatibility["stream.id"] || 0) + "%" });
            }
            sources.push({
                id: id,
                providerType: providerType,
                name: sourceString(source.name) || (providerType == "xmltv" ? "XMLTV" : providerType.toUpperCase()),
                typeLabel: providerType == "m3u" ? "M3U" : providerType == "hdhr" ? "HDHomeRun" : "XMLTV",
                source: sourceLocationForDisplay(source["file.source"]),
                lastUpdate: sourceString(source["last.update"]),
                availability: sourceNumber(rawAvailability),
                availabilityKnown: rawAvailability !== undefined && rawAvailability !== null,
                counts: counts,
            });
        });
    });
    sources.sort((a, b) => {
        var byName = a.name.localeCompare(b.name);
        return byName == 0 ? a.id.localeCompare(b.id) : byName;
    });
    return sources;
}
function renderSourceManagementPage(destination, host) {
    host.innerHTML = "";
    var page = document.createElement("div");
    page.className = "tf-sources";
    page.setAttribute("data-source-destination", destination);
    var header = document.createElement("header");
    header.className = "tf-sources-header";
    var titleGroup = document.createElement("div");
    var title = document.createElement("h1");
    title.textContent = destination == "playlist" ? "{{.mainMenu.item.playlist}}" : "{{.mainMenu.item.xmltv}}";
    var purpose = document.createElement("p");
    purpose.textContent = destination == "playlist" ? "{{.sources.playlist.purpose}}" : "{{.sources.xmltv.purpose}}";
    titleGroup.appendChild(title);
    titleGroup.appendChild(purpose);
    header.appendChild(titleGroup);
    header.appendChild(createSourceAddButton(destination));
    page.appendChild(header);
    if (destination == "xmltv" && sourceString(sourceRecord(SERVER["clientInfo"]).epgSource) != "XEPG") {
        var modeNote = document.createElement("div");
        modeNote.className = "tf-source-mode-note";
        modeNote.textContent = "{{.sources.xmltv.pmsNote}}";
        var settingsAction = document.createElement("button");
        settingsAction.type = "button";
        settingsAction.textContent = "{{.sources.xmltv.reviewEpgSource}}";
        settingsAction.addEventListener("click", function () {
            openDestination("settings", true);
        });
        modeNote.appendChild(settingsAction);
        page.appendChild(modeNote);
    }
    page.appendChild(renderSourcePageFeedback(destination));
    var sources = selectSourceList(SERVER, destination);
    if (sources.length == 0) {
        page.appendChild(renderSourceEmptyState(destination));
    }
    else {
        var list = document.createElement("div");
        list.className = "tf-source-list";
        list.setAttribute("role", "list");
        sources.forEach(source => list.appendChild(renderSourceRow(source)));
        page.appendChild(list);
    }
    host.appendChild(page);
}
function createSourceAddButton(destination) {
    var button = document.createElement("button");
    button.type = "button";
    button.className = "tf-source-primary-action";
    button.setAttribute("data-source-focus-key", destination + ":add");
    button.textContent = destination == "playlist" ? "{{.sources.playlist.add}}" : "{{.sources.xmltv.add}}";
    button.addEventListener("click", function () {
        openSourcePopup(destination == "playlist" ? "playlist" : "xmltv", undefined, button);
    });
    return button;
}
function renderSourcePageFeedback(destination) {
    var region = document.createElement("div");
    region.id = "source-page-status";
    region.className = "tf-source-page-status";
    region.setAttribute("role", "status");
    region.setAttribute("aria-live", "polite");
    var feedback = sourcePageFeedback[destination];
    if (!feedback) {
        region.hidden = true;
        return region;
    }
    region.setAttribute("data-state", feedback.state);
    var message = document.createElement("span");
    message.textContent = feedback.message;
    region.appendChild(message);
    if (feedback.nextDestination && feedback.nextLabel) {
        var next = document.createElement("button");
        next.type = "button";
        next.textContent = feedback.nextLabel;
        next.addEventListener("click", function () {
            openDestination(feedback.nextDestination, true);
        });
        region.appendChild(next);
    }
    return region;
}
function renderSourceEmptyState(destination) {
    var empty = document.createElement("section");
    empty.className = "tf-source-empty";
    var title = document.createElement("h2");
    title.textContent = destination == "playlist" ? "{{.sources.playlist.emptyTitle}}" : "{{.sources.xmltv.emptyTitle}}";
    var explanation = document.createElement("p");
    explanation.textContent = destination == "playlist" ? "{{.sources.playlist.emptyDescription}}" : "{{.sources.xmltv.emptyDescription}}";
    empty.appendChild(title);
    empty.appendChild(explanation);
    empty.appendChild(createSourceAddButton(destination));
    return empty;
}
function renderSourceRow(source) {
    var row = document.createElement("article");
    row.className = "tf-source-row";
    row.setAttribute("role", "listitem");
    row.setAttribute("data-source-id", source.id);
    row.setAttribute("data-source-type", source.providerType);
    var identity = document.createElement("div");
    identity.className = "tf-source-identity";
    var name = document.createElement("h2");
    name.textContent = source.name;
    var type = document.createElement("p");
    type.className = "tf-source-type";
    type.textContent = source.typeLabel;
    var location = document.createElement("code");
    location.textContent = source.source || "{{.sources.locationUnavailable}}";
    identity.appendChild(name);
    identity.appendChild(type);
    identity.appendChild(location);
    row.appendChild(identity);
    var health = document.createElement("div");
    health.className = "tf-source-health";
    var availability = document.createElement("strong");
    availability.className = "tf-source-availability";
    if (!source.availabilityKnown) {
        availability.textContent = "{{.sources.availabilityUnknown}}";
        health.setAttribute("data-status", "unknown");
    }
    else if (source.availability > 0) {
        availability.textContent = "{{.sources.available}} · " + source.availability + "%";
        health.setAttribute("data-status", "ready");
    }
    else {
        availability.textContent = "{{.sources.unavailable}} · " + source.availability + "%";
        health.setAttribute("data-status", "unavailable");
    }
    var updated = document.createElement("span");
    updated.textContent = source.lastUpdate ? "{{.sources.lastUpdate}}: " + source.lastUpdate : "{{.sources.neverUpdated}}";
    health.appendChild(availability);
    health.appendChild(updated);
    row.appendChild(health);
    var counts = document.createElement("dl");
    counts.className = "tf-source-counts";
    source.counts.forEach(count => {
        var group = document.createElement("div");
        var term = document.createElement("dt");
        term.textContent = count.label;
        var value = document.createElement("dd");
        value.textContent = count.value;
        group.appendChild(term);
        group.appendChild(value);
        counts.appendChild(group);
    });
    row.appendChild(counts);
    var feedback = sourceFeedbackByKey[sourceFeedbackKey(source.providerType, source.id)];
    var feedbackElement = document.createElement("p");
    feedbackElement.className = "tf-source-feedback";
    feedbackElement.setAttribute("role", "status");
    feedbackElement.setAttribute("aria-live", "polite");
    if (feedback) {
        feedbackElement.textContent = feedback.message;
        feedbackElement.setAttribute("data-state", feedback.state);
    }
    else {
        feedbackElement.hidden = true;
    }
    row.appendChild(feedbackElement);
    var actions = document.createElement("div");
    actions.className = "tf-source-actions";
    actions.appendChild(createSourceRowAction("{{.sources.actions.edit}}", source.providerType + ":" + source.id + ":edit", function (invoker) {
        openSourcePopup(source.providerType, { id: source.id }, invoker);
    }));
    actions.appendChild(createSourceRowAction("{{.sources.actions.update}}", source.providerType + ":" + source.id + ":update", function (invoker) {
        invokeSourcePopupAction(source.providerType, source.id, "update", invoker);
    }));
    var remove = createSourceRowAction("{{.sources.actions.delete}}", source.providerType + ":" + source.id + ":delete", function (invoker) {
        invokeSourcePopupAction(source.providerType, source.id, "delete", invoker);
    });
    remove.className += " tf-source-delete-action";
    actions.appendChild(remove);
    row.appendChild(actions);
    return row;
}
function createSourceRowAction(label, focusKey, listener) {
    var button = document.createElement("button");
    button.type = "button";
    button.textContent = label;
    button.setAttribute("data-source-focus-key", focusKey);
    button.addEventListener("click", function () {
        listener(button);
    });
    return button;
}
function invokeSourcePopupAction(providerType, id, action, invoker) {
    openSourcePopup(providerType, { id: id }, invoker);
    savePopupData(providerType, id, action == "delete", action == "update" ? 1 : 0);
}
function openSourcePopup(dataType, element, invoker) {
    sourcePopupInvoker = invoker;
    sourcePopupFocusKey = invoker ? sourceString(invoker.getAttribute("data-source-focus-key")) : "";
    var modal = document.getElementById("popup");
    if (modal) {
        modal.setAttribute("role", "dialog");
        modal.setAttribute("aria-modal", "true");
    }
    if (modal && !sourcePopupFocusListenerAttached) {
        sourcePopupFocusListenerAttached = true;
        modal.addEventListener("hidden.bs.modal", function () {
            var focusTarget = sourcePopupInvoker && document.contains(sourcePopupInvoker) ? sourcePopupInvoker : sourcePopupReplacement(sourcePopupFocusKey);
            if (focusTarget) {
                focusTarget.focus();
            }
            sourcePopupInvoker = undefined;
            sourcePopupFocusKey = "";
        });
    }
    if (modal && !sourcePopupShownListenerAttached) {
        sourcePopupShownListenerAttached = true;
        modal.addEventListener("shown.bs.modal", function () {
            focusSourcePopupFirstControl();
        });
    }
    openPopUp(dataType, element);
}
function focusSourcePopupFirstControl() {
    var popup = document.getElementById("popup-custom");
    if (!popup || !popup.classList.contains("tf-source-popup")) {
        return;
    }
    var firstControl = popup.querySelector("input, select, button");
    if (firstControl) {
        firstControl.focus();
    }
}
function focusRenderedSourcePopup(modal) {
    if (modal && modal.classList.contains("show")) {
        window.setTimeout(focusSourcePopupFirstControl, 0);
    }
}
function sourcePopupReplacement(focusKey) {
    if (!focusKey) {
        return undefined;
    }
    var candidates = document.querySelectorAll("[data-source-focus-key]");
    for (var index = 0; index < candidates.length; index++) {
        var candidate = candidates[index];
        if (candidate.getAttribute("data-source-focus-key") == focusKey) {
            return candidate;
        }
    }
    return undefined;
}
function enhanceSourcePopup(dataType) {
    var popup = document.getElementById("popup-custom");
    if (!popup) {
        return;
    }
    popup.classList.remove("tf-source-popup");
    var sourcePopup = dataType == "playlist" || dataType == "m3u" || dataType == "hdhr" || dataType == "xmltv";
    var title = popup.querySelector("h3");
    var modal = document.getElementById("popup");
    if (modal) {
        if (title) {
            title.id = sourcePopup ? "source-popup-title" : "popup-title";
            modal.setAttribute("aria-labelledby", title.id);
        }
        else {
            modal.removeAttribute("aria-labelledby");
        }
    }
    if (!sourcePopup) {
        return;
    }
    popup.classList.add("tf-source-popup");
    if (dataType == "playlist") {
        focusRenderedSourcePopup(modal);
        return;
    }
    var fields = popup.querySelectorAll("input, select");
    Array.prototype.forEach.call(fields, function (field) {
        var row = field.closest("tr");
        var title = row ? row.querySelector("td:first-child") : undefined;
        if (title && title.textContent) {
            field.setAttribute("aria-label", title.textContent.replace(/:\s*$/, ""));
        }
    });
    var sourceInput = popup.querySelector('[name="file.source"]');
    if (sourceInput && sourceInput.parentElement) {
        var help = document.createElement("p");
        help.id = "source-location-help";
        help.className = "tf-source-field-help";
        help.textContent = dataType == "hdhr" ? "{{.sources.forms.hdhrHelp}}" : "{{.sources.forms.locationHelp}}";
        sourceInput.setAttribute("aria-describedby", help.id + " source-file-source-error");
        sourceInput.parentElement.appendChild(help);
    }
    ;
    ["name", "file.source"].forEach(fieldName => {
        var field = popup.querySelector('[name="' + fieldName + '"]');
        if (!field || !field.parentElement) {
            return;
        }
        field.setAttribute("aria-required", "true");
        var error = document.createElement("p");
        error.id = "source-" + fieldName.replace(".", "-") + "-error";
        error.className = "tf-source-field-error";
        error.hidden = true;
        field.parentElement.appendChild(error);
    });
    var status = document.createElement("p");
    status.id = "source-form-status";
    status.className = "tf-source-form-status";
    status.setAttribute("role", "status");
    status.setAttribute("aria-live", "polite");
    status.hidden = true;
    popup.appendChild(status);
    focusRenderedSourcePopup(modal);
}
function sourceLocationAccepted(value) {
    var source = sourceString(value).trim();
    if (!source) {
        return false;
    }
    var scheme = source.match(/^([a-z][a-z0-9+.-]*):\/\//i);
    return !scheme || scheme[1].toLowerCase() == "http" || scheme[1].toLowerCase() == "https";
}
function validateSourcePopup(dataType) {
    if (dataType != "m3u" && dataType != "hdhr" && dataType != "xmltv") {
        return true;
    }
    var popup = document.getElementById("popup-custom");
    if (!popup) {
        return false;
    }
    sourceClearFormErrors();
    var name = popup.querySelector('[name="name"]');
    var source = popup.querySelector('[name="file.source"]');
    if (!name || !name.value.trim()) {
        sourceSetFieldError(name, "source-name-error", "{{.sources.forms.nameRequired}}");
        return false;
    }
    if (!source || !source.value.trim()) {
        sourceSetFieldError(source, "source-file-source-error", "{{.sources.forms.locationRequired}}");
        return false;
    }
    if (dataType != "hdhr" && !sourceLocationAccepted(source.value)) {
        sourceSetFieldError(source, "source-file-source-error", "{{.sources.forms.locationInvalid}}");
        return false;
    }
    return true;
}
function sourceClearFormErrors() {
    var popup = document.getElementById("popup-custom");
    if (!popup) {
        return;
    }
    var invalid = popup.querySelectorAll('[aria-invalid="true"]');
    Array.prototype.forEach.call(invalid, function (field) {
        field.removeAttribute("aria-invalid");
    });
    var errors = popup.querySelectorAll(".tf-source-field-error");
    Array.prototype.forEach.call(errors, function (error) {
        error.textContent = "";
        error.hidden = true;
    });
}
function sourceSetFieldError(field, errorID, message) {
    if (field) {
        field.setAttribute("aria-invalid", "true");
    }
    var error = document.getElementById(errorID);
    if (error) {
        error.textContent = message;
        error.hidden = false;
    }
    if (field) {
        field.focus();
    }
}
function sourceSetFormStatus(message, state) {
    var status = document.getElementById("source-form-status");
    if (!status) {
        return;
    }
    status.textContent = message;
    status.setAttribute("data-state", state);
    status.hidden = false;
}
function beginSourceRequest(dataType, id, remove, option) {
    if (dataType != "m3u" && dataType != "hdhr" && dataType != "xmltv") {
        return;
    }
    var providerType = dataType;
    var message = remove == true ? "Deleting source…" : option == 1 ? "Updating source…" : id == "-" ? "Adding source…" : "Saving source…";
    var feedback = { state: "progress", message: message };
    sourceFeedbackByKey[sourceFeedbackKey(providerType, id)] = feedback;
    sourceSetFormStatus(message, "progress");
}
function sourceRequestDescriptor(command, data) {
    var commandMap = {
        saveFilesM3U: { providerType: "m3u", mode: "save" },
        updateFileM3U: { providerType: "m3u", mode: "update" },
        saveFilesHDHR: { providerType: "hdhr", mode: "save" },
        updateFileHDHR: { providerType: "hdhr", mode: "update" },
        saveFilesXMLTV: { providerType: "xmltv", mode: "save" },
        updateFileXMLTV: { providerType: "xmltv", mode: "update" },
    };
    var contract = commandMap[command];
    if (!contract) {
        return undefined;
    }
    var files = sourceRecord(sourceRecord(data).files);
    var providers = sourceRecord(files[contract.providerType]);
    var ids = Object.keys(providers);
    if (ids.length != 1) {
        return undefined;
    }
    var values = sourceRecord(providers[ids[0]]);
    return {
        providerType: contract.providerType,
        destination: contract.providerType == "xmltv" ? "xmltv" : "playlist",
        id: ids[0],
        mode: values.delete === true ? "delete" : contract.mode,
        values: values,
    };
}
function completeSourceRequest(command, data, response) {
    var request = sourceRequestDescriptor(command, data);
    if (!request) {
        return;
    }
    var root = sourceRecord(response);
    if (root.status !== true) {
        var error = root.status === false ? sourceString(root.err) || "Source request failed." : "{{.sources.responseInvalid}}";
        var failed = { state: "error", message: error };
        sourceFeedbackByKey[sourceFeedbackKey(request.providerType, request.id)] = failed;
        sourceSetFormStatus(error, "error");
        return;
    }
    if (request.mode == "delete") {
        delete sourceFeedbackByKey[sourceFeedbackKey(request.providerType, request.id)];
        sourcePageFeedback[request.destination] = { state: "success", message: "Source deleted." };
        return;
    }
    var responseFiles = sourceRecord(sourceRecord(sourceRecord(root.settings).files)[request.providerType]);
    var resolvedID = request.id;
    if (resolvedID == "-") {
        Object.keys(responseFiles).some(id => {
            var candidate = sourceRecord(responseFiles[id]);
            if (sourceString(candidate["file.source"]) == sourceString(request.values["file.source"]) &&
                sourceString(candidate.name) == sourceString(request.values.name)) {
                resolvedID = id;
                return true;
            }
            return false;
        });
    }
    var saved = sourceRecord(responseFiles[resolvedID]);
    var lastUpdate = sourceString(saved["last.update"]);
    var message = request.mode == "update" ? "Update completed" + (lastUpdate ? ". Last update: " + lastUpdate + "." : ".") : "Source saved.";
    sourceFeedbackByKey[sourceFeedbackKey(request.providerType, resolvedID)] = { state: "success", message: message };
    if (request.id == "-") {
        sourcePageFeedback[request.destination] = request.destination == "playlist" ? {
            state: "success", message: "Playlist added. Next, select imported streams with a Filter.",
            nextDestination: "filter", nextLabel: "Open Filter",
        } : {
            state: "success", message: "XMLTV source added. Next, review guide assignments in Mapping.",
            nextDestination: "mapping", nextLabel: "Open Mapping",
        };
    }
}
