"use strict";
function renderOverview(server) {
    var host = document.getElementById("overview-content");
    if (!host) {
        return;
    }
    var state = selectOverviewState(server);
    while (host.firstChild) {
        host.removeChild(host.firstChild);
    }
    var overview = document.createElement("div");
    overview.className = "tf-overview";
    var header = document.createElement("header");
    header.className = "tf-overview-header";
    var title = document.createElement("h1");
    title.textContent = "Overview";
    var introduction = document.createElement("p");
    introduction.textContent = "Follow the Threadfin signal path from imported sources to generated outputs.";
    header.appendChild(title);
    header.appendChild(introduction);
    overview.appendChild(header);
    var pathHeading = document.createElement("h2");
    pathHeading.textContent = "Signal path";
    overview.appendChild(pathHeading);
    overview.appendChild(createOverviewSignalPath(state.stages));
    var panels = document.createElement("div");
    panels.className = "tf-overview-panels";
    panels.appendChild(createOverviewActivityPanel(state));
    panels.appendChild(createOverviewAttentionPanel(state));
    panels.appendChild(createOverviewEndpointsPanel(state.outputs.endpoints));
    panels.appendChild(createOverviewSourcesPanel(state.sources));
    overview.appendChild(panels);
    var copyStatus = document.createElement("p");
    copyStatus.id = "overview-copy-status";
    copyStatus.className = "tf-overview-status";
    copyStatus.setAttribute("role", "status");
    copyStatus.setAttribute("aria-live", "polite");
    copyStatus.setAttribute("aria-atomic", "true");
    overview.appendChild(copyStatus);
    host.appendChild(overview);
}
function createOverviewSignalPath(stages) {
    var path = document.createElement("ol");
    path.className = "tf-signal-path";
    path.setAttribute("aria-label", "Threadfin signal path");
    stages.forEach(stage => {
        var item = document.createElement("li");
        item.className = "tf-signal-stage";
        item.setAttribute("data-stage", stage.key);
        item.setAttribute("data-status", stage.status);
        var status = document.createElement("p");
        status.className = "tf-signal-stage-status";
        status.textContent = overviewStatusLabel(stage.status);
        var heading = document.createElement("h3");
        heading.textContent = stage.label;
        var summary = document.createElement("p");
        summary.className = "tf-signal-stage-summary";
        summary.textContent = stage.summary;
        var explanation = document.createElement("p");
        explanation.className = "tf-signal-stage-explanation";
        explanation.textContent = stage.explanation;
        var action = document.createElement("button");
        action.type = "button";
        action.className = "tf-overview-action";
        action.textContent = stage.action.label;
        action.addEventListener("click", function () {
            openDestination(stage.action.destination, true);
        });
        item.appendChild(status);
        item.appendChild(heading);
        item.appendChild(summary);
        item.appendChild(explanation);
        item.appendChild(action);
        path.appendChild(item);
    });
    return path;
}
function overviewStatusLabel(status) {
    switch (status) {
        case "ready":
            return "Ready";
        case "attention":
            return "Needs attention";
        case "empty":
            return "Not configured";
        case "managed":
            return "Managed by client";
        default:
            return "Waiting";
    }
}
function createOverviewPanel(title, className) {
    var panel = document.createElement("section");
    panel.className = "tf-overview-panel " + className;
    var heading = document.createElement("h2");
    heading.textContent = title;
    panel.appendChild(heading);
    return panel;
}
function createOverviewActivityPanel(state) {
    var panel = createOverviewPanel("Current activity", "tf-overview-activity");
    var metrics = document.createElement("dl");
    metrics.className = "tf-overview-metrics";
    appendOverviewMetric(metrics, "Active streams", "overview-active-streams", String(state.activity.activeStreams));
    appendOverviewMetric(metrics, "Client capacity", "overview-client-capacity", overviewCapacityLabel(state.activity.clients));
    appendOverviewMetric(metrics, "Source capacity", "overview-playlist-capacity", overviewCapacityLabel(state.activity.playlists));
    panel.appendChild(metrics);
    var explanation = document.createElement("p");
    explanation.className = "tf-overview-panel-note";
    explanation.textContent = state.activity.activeStreams == 0 ? "No client streams are active now." : "Threadfin currently observes active client streams.";
    panel.appendChild(explanation);
    return panel;
}
function appendOverviewMetric(list, label, id, value) {
    var item = document.createElement("div");
    var term = document.createElement("dt");
    term.textContent = label;
    var description = document.createElement("dd");
    description.id = id;
    description.textContent = value;
    item.appendChild(term);
    item.appendChild(description);
    list.appendChild(item);
}
function overviewCapacityLabel(capacity) {
    return capacity.active + " / " + capacity.total;
}
function createOverviewAttentionPanel(state) {
    var panel = createOverviewPanel("Attention required", "tf-overview-attention");
    var counts = document.createElement("dl");
    counts.className = "tf-overview-metrics";
    appendOverviewMetric(counts, "Errors", "overview-errors", String(state.attention.errors));
    appendOverviewMetric(counts, "Warnings", "overview-warnings", String(state.attention.warnings));
    panel.appendChild(counts);
    var explanation = document.createElement("p");
    explanation.className = "tf-overview-panel-note";
    explanation.id = "overview-attention-summary";
    explanation.textContent = overviewAttentionSummary(state.attention);
    panel.appendChild(explanation);
    var action = document.createElement("button");
    action.type = "button";
    action.className = "tf-overview-action";
    action.textContent = "View Log";
    action.addEventListener("click", function () {
        openDestination("log", true);
    });
    panel.appendChild(action);
    return panel;
}
function overviewAttentionSummary(attention) {
    if (attention.errors == 0 && attention.warnings == 0) {
        return "Threadfin reports no current errors or warnings.";
    }
    return "Threadfin reports " + attention.errors + (attention.errors == 1 ? " error and " : " errors and ") + attention.warnings + (attention.warnings == 1 ? " warning." : " warnings.");
}
function createOverviewEndpointsPanel(endpoints) {
    var panel = createOverviewPanel("Output endpoints", "tf-overview-endpoints");
    var list = document.createElement("ul");
    list.className = "tf-endpoint-list";
    endpoints.forEach(endpoint => {
        var item = document.createElement("li");
        item.className = "tf-endpoint";
        var heading = document.createElement("h3");
        heading.textContent = endpoint.label;
        var value = document.createElement("code");
        value.textContent = endpoint.available ? endpoint.value : "Unavailable";
        var explanation = document.createElement("p");
        explanation.textContent = endpoint.explanation;
        item.appendChild(heading);
        item.appendChild(value);
        item.appendChild(explanation);
        if (endpoint.available) {
            var copy = document.createElement("button");
            copy.type = "button";
            copy.className = "tf-overview-action tf-copy-action";
            copy.textContent = "Copy " + endpoint.label;
            bindOverviewCopyAction(copy, endpoint.value, endpoint.label);
            item.appendChild(copy);
        }
        list.appendChild(item);
    });
    panel.appendChild(list);
    return panel;
}
function bindOverviewCopyAction(button, value, label) {
    if (typeof ClipboardJS != "undefined") {
        var helper = new ClipboardJS(button, {
            text: function () { return value; },
        });
        helper.on("success", function (event) {
            if (event && event.clearSelection) {
                event.clearSelection();
            }
            announceOverviewCopyStatus(label + " copied.");
        });
        helper.on("error", function () {
            copyOverviewWithFallback(value, label);
        });
        return;
    }
    button.addEventListener("click", function () {
        copyOverviewWithFallback(value, label);
    });
}
function copyOverviewWithFallback(value, label) {
    if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(value).then(function () {
            announceOverviewCopyStatus(label + " copied.");
        }).catch(function () {
            copyOverviewWithSelection(value, label);
        });
        return;
    }
    copyOverviewWithSelection(value, label);
}
function copyOverviewWithSelection(value, label) {
    var temporary = document.createElement("textarea");
    temporary.value = value;
    temporary.setAttribute("readonly", "");
    temporary.style.position = "fixed";
    temporary.style.opacity = "0";
    document.body.appendChild(temporary);
    temporary.select();
    var copied = false;
    try {
        copied = document.execCommand("copy");
    }
    catch (_) {
        copied = false;
    }
    document.body.removeChild(temporary);
    announceOverviewCopyStatus(copied ? label + " copied." : "Copy failed. Select and copy the endpoint manually.");
}
function announceOverviewCopyStatus(message) {
    var status = document.getElementById("overview-copy-status");
    if (status) {
        status.textContent = message;
    }
}
function createOverviewSourcesPanel(sources) {
    var panel = createOverviewPanel("Recent source state", "tf-overview-sources");
    if (sources.length == 0) {
        var empty = document.createElement("p");
        empty.className = "tf-overview-panel-note";
        empty.textContent = "No Playlist or XMLTV sources are configured yet.";
        panel.appendChild(empty);
        return panel;
    }
    var list = document.createElement("ul");
    list.className = "tf-source-state-list";
    sources.forEach(source => {
        var item = document.createElement("li");
        item.className = "tf-source-state";
        item.setAttribute("data-status", source.status);
        var heading = document.createElement("h3");
        heading.textContent = source.name;
        var kind = document.createElement("p");
        kind.className = "tf-source-state-kind";
        kind.textContent = source.kind + " · " + overviewSourceStatusLabel(source);
        var updated = document.createElement("p");
        updated.textContent = source.lastUpdate ? "Last update: " + source.lastUpdate : "No successful update recorded";
        item.appendChild(heading);
        item.appendChild(kind);
        item.appendChild(updated);
        list.appendChild(item);
    });
    panel.appendChild(list);
    return panel;
}
function overviewSourceStatusLabel(source) {
    if (source.status == "unknown") {
        return "Availability unknown";
    }
    if (source.status == "unavailable") {
        return "Unavailable";
    }
    return source.availability + "% available";
}
function refreshOverviewOperationalState(response) {
    var activity = selectActivityState(response);
    var attention = selectAttentionState(response);
    setOverviewOperationalText("overview-active-streams", String(activity.activeStreams));
    setOverviewOperationalText("overview-client-capacity", overviewCapacityLabel(activity.clients));
    setOverviewOperationalText("overview-playlist-capacity", overviewCapacityLabel(activity.playlists));
    setOverviewOperationalText("overview-errors", String(attention.errors));
    setOverviewOperationalText("overview-warnings", String(attention.warnings));
    setOverviewOperationalText("overview-attention-summary", overviewAttentionSummary(attention));
}
function setOverviewOperationalText(id, value) {
    var element = document.getElementById(id);
    if (element && element.textContent != value) {
        element.textContent = value;
    }
}
