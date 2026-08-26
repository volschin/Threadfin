"use strict";
function overviewRecord(value) {
    if (value && typeof value == "object" && !Array.isArray(value)) {
        return value;
    }
    return {};
}
function overviewNumber(value) {
    var parsed = typeof value == "number" ? value : Number(value);
    return isFinite(parsed) && parsed > 0 ? Math.floor(parsed) : 0;
}
function overviewString(value) {
    return typeof value == "string" ? value.trim() : "";
}
function overviewObjectCount(value) {
    return Object.keys(overviewRecord(value)).length;
}
function selectPlaylistCount(server) {
    var settings = overviewRecord(overviewRecord(server).settings);
    var files = overviewRecord(settings.files);
    return overviewObjectCount(files.m3u) + overviewObjectCount(files.hdhr);
}
function selectSelectedStreamCount(server) {
    var clientInfo = overviewRecord(overviewRecord(server).clientInfo);
    var streamCount = overviewString(clientInfo.streams).match(/^\s*(\d+)/);
    if (streamCount) {
        return overviewNumber(streamCount[1]);
    }
    var data = overviewRecord(overviewRecord(server).data);
    var preview = overviewRecord(data.StreamPreviewUI);
    return Array.isArray(preview.activeStreams) ? preview.activeStreams.length : 0;
}
function selectXMLTVState(server) {
    var root = overviewRecord(server);
    var settings = overviewRecord(root.settings);
    var clientInfo = overviewRecord(root.clientInfo);
    var files = overviewRecord(settings.files);
    var sources = overviewRecord(files.xmltv);
    var applicable = overviewString(settings.epgSource || clientInfo.epgSource).toUpperCase() == "XEPG";
    var ready = false;
    Object.keys(sources).forEach(key => {
        var source = overviewRecord(sources[key]);
        var compatibility = overviewRecord(source.compatibility);
        if (overviewNumber(source["provider.availability"]) > 0 && overviewNumber(compatibility["xmltv.channels"]) > 0) {
            ready = true;
        }
    });
    return {
        applicable: applicable,
        ready: applicable && ready,
        sourceCount: Object.keys(sources).length,
    };
}
function selectMappingState(server) {
    var root = overviewRecord(server);
    var xepg = overviewRecord(root.xepg);
    var mappings = overviewRecord(xepg.epgMapping);
    var activeCount = 0;
    var unresolvedCount = 0;
    Object.keys(mappings).forEach(key => {
        var mapping = overviewRecord(mappings[key]);
        if (mapping["x-active"] === true && mapping["x-hide-channel"] !== true) {
            activeCount++;
        }
        else {
            unresolvedCount++;
        }
    });
    return { activeCount: activeCount, unresolvedCount: unresolvedCount };
}
function overviewEndpointAvailable(value, endpoint) {
    var address = overviewString(value);
    if (!address) {
        return false;
    }
    if (endpoint == "dvr") {
        return !/\s/.test(address) && address.indexOf(":") > 0;
    }
    return /^https?:\/\//i.test(address);
}
function selectOutputState(server, selectedStreamCount, xmltv, mapping) {
    var root = overviewRecord(server);
    var clientInfo = overviewRecord(root.clientInfo);
    var dvrAvailable = selectedStreamCount > 0 && overviewEndpointAvailable(clientInfo.DVR, "dvr");
    var m3uAvailable = xmltv.applicable && mapping.activeCount > 0 && overviewEndpointAvailable(clientInfo["m3u-url"], "m3u");
    var xmltvAvailable = xmltv.applicable && mapping.activeCount > 0 && overviewEndpointAvailable(clientInfo["xepg-url"], "xmltv");
    var ready = dvrAvailable;
    if (xmltv.applicable) {
        ready = ready && mapping.activeCount > 0 && m3uAvailable && xmltvAvailable;
    }
    return {
        ready: ready,
        endpoints: [
            {
                key: "dvr",
                label: "DVR address",
                value: dvrAvailable ? overviewString(clientInfo.DVR) : "",
                available: dvrAvailable,
                explanation: dvrAvailable ? "Use this address when a client discovers Threadfin as a tuner." : selectedStreamCount == 0 ? "The DVR address becomes usable after streams are selected." : "The DVR address is not available yet.",
            },
            {
                key: "m3u",
                label: "M3U URL",
                value: m3uAvailable ? overviewString(clientInfo["m3u-url"]) : "",
                available: m3uAvailable,
                explanation: !xmltv.applicable ? "M3U output requires XEPG mode." : mapping.activeCount == 0 ? "M3U output becomes usable after channels are active in Mapping." : m3uAvailable ? "Threadfin's generated channel playlist." : "The M3U output is not available yet.",
            },
            {
                key: "xmltv",
                label: "XMLTV URL",
                value: xmltvAvailable ? overviewString(clientInfo["xepg-url"]) : "",
                available: xmltvAvailable,
                explanation: !xmltv.applicable ? "Guide data is managed by the client in PMS mode." : mapping.activeCount == 0 ? "XMLTV output becomes usable after channels are active in Mapping." : xmltvAvailable ? "Threadfin's generated guide data." : "The XMLTV output is not available yet.",
            },
        ],
    };
}
function selectActivityState(server) {
    var clientInfo = overviewRecord(overviewRecord(server).clientInfo);
    return {
        activeStreams: overviewNumber(clientInfo.activeClients),
        clients: {
            active: overviewNumber(clientInfo.activeClients),
            total: overviewNumber(clientInfo.totalClients),
        },
        playlists: {
            active: overviewNumber(clientInfo.activePlaylist),
            total: overviewNumber(clientInfo.totalPlaylist),
        },
    };
}
function selectAttentionState(server) {
    var root = overviewRecord(server);
    var log = overviewRecord(root.log);
    var clientInfo = overviewRecord(root.clientInfo);
    return {
        errors: overviewNumber(log.errors !== undefined ? log.errors : clientInfo.errors),
        warnings: overviewNumber(log.warnings !== undefined ? log.warnings : clientInfo.warnings),
    };
}
function selectSourceState(server) {
    var root = overviewRecord(server);
    var settings = overviewRecord(root.settings);
    var files = overviewRecord(settings.files);
    var sources = [];
    function appendSources(fileType, kind) {
        var configured = overviewRecord(files[fileType]);
        Object.keys(configured).forEach(key => {
            var source = overviewRecord(configured[key]);
            var rawAvailability = source["provider.availability"];
            var availability = overviewNumber(rawAvailability);
            sources.push({
                id: key,
                name: overviewString(source.name) || kind,
                kind: kind,
                lastUpdate: overviewString(source["last.update"]),
                availability: availability,
                status: rawAvailability === undefined || rawAvailability === null ? "unknown" : availability > 0 ? "ready" : "unavailable",
            });
        });
    }
    appendSources("m3u", "Playlist");
    appendSources("hdhr", "HDHomeRun");
    appendSources("xmltv", "XMLTV");
    sources.sort((a, b) => {
        if (a.lastUpdate == b.lastUpdate) {
            return a.name.localeCompare(b.name);
        }
        return a.lastUpdate < b.lastUpdate ? 1 : -1;
    });
    return sources;
}
function selectOverviewStages(state) {
    var playlistSources = state.sources.filter(source => source.kind != "XMLTV");
    var playlistsReady = playlistSources.some(source => source.status == "ready");
    var stages = [];
    stages.push(state.playlistCount == 0 ? {
        key: "playlist", label: "Playlist", status: "empty", summary: "No playlist configured",
        explanation: "Add an M3U playlist or HDHomeRun source to begin.",
        action: { label: "Add playlist", destination: "playlist" },
    } : playlistsReady ? {
        key: "playlist", label: "Playlist", status: "ready", summary: state.playlistCount + (state.playlistCount == 1 ? " source ready" : " sources ready"),
        explanation: "Threadfin has a reachable channel source.",
        action: { label: "View playlist", destination: "playlist" },
    } : {
        key: "playlist", label: "Playlist", status: "attention", summary: "Source needs attention",
        explanation: "The configured channel source is not currently available.",
        action: { label: "Review playlist", destination: "playlist" },
    });
    stages.push(state.playlistCount == 0 ? {
        key: "filter", label: "Filter", status: "waiting", summary: "Waiting for playlist",
        explanation: "Filters select which imported streams continue to Mapping.",
        action: { label: "View filters", destination: "filter" },
    } : state.selectedStreamCount > 0 ? {
        key: "filter", label: "Filter", status: "ready", summary: state.selectedStreamCount + " selected",
        explanation: "These imported streams continue through the signal path.",
        action: { label: "Review filters", destination: "filter" },
    } : {
        key: "filter", label: "Filter", status: "attention", summary: "No streams selected",
        explanation: "Create a filter to select imported streams for Mapping.",
        action: { label: "Create filter", destination: "filter" },
    });
    stages.push(!state.xmltv.applicable ? {
        key: "xmltv", label: "XMLTV", status: "managed", summary: "Managed by client",
        explanation: "PMS mode leaves guide data management to the connected client.",
        action: { label: "Review EPG source", destination: "settings" },
    } : state.xmltv.ready ? {
        key: "xmltv", label: "XMLTV", status: "ready", summary: state.xmltv.sourceCount + (state.xmltv.sourceCount == 1 ? " guide ready" : " guides ready"),
        explanation: "Guide channels are available to XEPG.",
        action: { label: "View XMLTV", destination: "xmltv" },
    } : {
        key: "xmltv", label: "XMLTV", status: "attention", summary: "Guide not ready",
        explanation: "XEPG needs an available XMLTV guide with channels.",
        action: { label: "Add XMLTV", destination: "xmltv" },
    });
    stages.push(!state.xmltv.applicable ? {
        key: "mapping", label: "Mapping", status: "managed", summary: "Managed by client",
        explanation: "PMS mode does not use Threadfin's XEPG Mapping workflow.",
        action: { label: "Review EPG source", destination: "settings" },
    } : state.selectedStreamCount == 0 ? {
        key: "mapping", label: "Mapping", status: "waiting", summary: "Waiting for selected streams",
        explanation: "Selected streams appear here after the Filter stage.",
        action: { label: "View Mapping", destination: "mapping" },
    } : state.mapping.unresolvedCount > 0 ? {
        key: "mapping", label: "Mapping", status: "attention", summary: state.mapping.unresolvedCount + " need attention",
        explanation: state.mapping.activeCount + " active; unresolved channels are not part of generated outputs.",
        action: { label: "Review " + state.mapping.unresolvedCount + " channels", destination: "mapping" },
    } : state.mapping.activeCount > 0 ? {
        key: "mapping", label: "Mapping", status: "ready", summary: state.mapping.activeCount + " active",
        explanation: "Active mappings are included in Threadfin's generated lineup.",
        action: { label: "View Mapping", destination: "mapping" },
    } : {
        key: "mapping", label: "Mapping", status: "attention", summary: "No active mappings",
        explanation: "Activate and map channels before using XEPG outputs.",
        action: { label: "Review Mapping", destination: "mapping" },
    });
    stages.push(state.outputs.ready ? {
        key: "outputs", label: "Outputs", status: "ready", summary: "Threadfin output ready",
        explanation: "Endpoints are ready to copy into a supported client; client setup is not observable here.",
        action: { label: "View output endpoints", destination: "connections" },
    } : {
        key: "outputs", label: "Outputs", status: "waiting", summary: "Waiting for lineup",
        explanation: "Complete the applicable stages before connecting a client.",
        action: { label: "View Connections", destination: "connections" },
    });
    return stages;
}
function selectOverviewState(server) {
    var playlistCount = selectPlaylistCount(server);
    var selectedStreamCount = selectSelectedStreamCount(server);
    var xmltv = selectXMLTVState(server);
    var mapping = selectMappingState(server);
    var outputs = selectOutputState(server, selectedStreamCount, xmltv, mapping);
    var stateWithoutStages = {
        playlistCount: playlistCount,
        selectedStreamCount: selectedStreamCount,
        xmltv: xmltv,
        mapping: mapping,
        outputs: outputs,
        activity: selectActivityState(server),
        attention: selectAttentionState(server),
        sources: selectSourceState(server),
    };
    var state = stateWithoutStages;
    state.stages = selectOverviewStages(stateWithoutStages);
    return state;
}
