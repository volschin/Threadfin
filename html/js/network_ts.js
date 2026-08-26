"use strict";
class Server {
    constructor(cmd) {
        this.cmd = cmd;
    }
    request(data) {
        if (SERVER_CONNECTION == true) {
            console.warn("WebSocket request skipped because another request is active:", this.cmd);
            completeTask5RequestFailure(this.cmd, data, "{{.sources.requestBusy}}");
            if (typeof completeMappingRequest == "function") {
                completeMappingRequest(this.cmd, data, { status: false, err: "{{.sources.requestBusy}}" }, "busy");
            }
            return;
        }
        SERVER_CONNECTION = true;
        if (this.cmd != "updateLog") {
            // showElement("loading", true)
            UNDO = new Object();
        }
        switch (window.location.protocol) {
            case "http:":
                this.protocol = "ws://";
                break;
            case "https:":
                this.protocol = "wss://";
                break;
        }
        var url = this.protocol + window.location.hostname + ":" + window.location.port + "/data/" + "?Token=" + getCookie("Token");
        var command = this.cmd;
        data["cmd"] = command;
        var ws = new WebSocket(url);
        var requestSettled = false;
        var settleTransportFailure = function () {
            if (requestSettled) {
                return;
            }
            requestSettled = true;
            SERVER_CONNECTION = false;
            showElement("loading", false);
            console.warn("WebSocket request failed:", command);
            completeTask5RequestFailure(command, data, "{{.sources.transportError}}");
            if (typeof completeMappingRequest == "function") {
                completeMappingRequest(command, data, { status: false, err: "{{.sources.transportError}}" }, "transport");
            }
        };
        ws.onopen = function () {
            WS_AVAILABLE = true;
            console.log("WebSocket request opened:", command);
            this.send(JSON.stringify(data));
        };
        ws.onerror = function (e) {
            settleTransportFailure();
            if (WS_AVAILABLE == false) {
                alert("No websocket connection to Threadfin could be established. Check your network configuration.");
            }
        };
        ws.onclose = function () {
            settleTransportFailure();
        };
        ws.onmessage = function (e) {
            var response;
            try {
                response = JSON.parse(e.data);
            }
            catch (_error) {
                settleTransportFailure();
                return;
            }
            requestSettled = true;
            SERVER_CONNECTION = false;
            showElement("loading", false);
            console.log("WebSocket response received:", command);
            if (typeof completeSourceRequest == "function") {
                completeSourceRequest(data["cmd"], data, response);
            }
            if (typeof completeFilterRequest == "function") {
                completeFilterRequest(data["cmd"], data, response);
            }
            if (data["cmd"] == "saveWizard" && typeof completeConfigurationWizardRequest == "function") {
                completeConfigurationWizardRequest(response);
            }
            if (typeof completeMappingRequest == "function") {
                completeMappingRequest(data["cmd"], data, response);
            }
            var responseIsObject = response && typeof response == "object" && !Array.isArray(response);
            if (!responseIsObject || response["status"] !== true) {
                if (data["cmd"] == "saveEpgMapping") {
                    if (responseIsObject && response["xepg"] && response["xepg"]["epgMapping"] && SERVER["xepg"]) {
                        SERVER["xepg"]["epgMapping"] = response["xepg"]["epgMapping"];
                    }
                    return;
                }
                if (responseIsObject && response["status"] === false) {
                    alert(response["err"] || "{{.sources.responseInvalid}}");
                    if (response.hasOwnProperty("reload")) {
                        location.reload();
                    }
                }
                else {
                    alert("{{.sources.responseInvalid}}");
                }
                return;
            }
            if (response.hasOwnProperty("token")) {
                document.cookie = "Token=" + response["token"];
            }
            if (response.hasOwnProperty("probeInfo")) {
                if (document.getElementById("probeDetails")) {
                    if (response["probeInfo"]["resolution"] !== undefined) {
                        document.getElementById("probeDetails").innerHTML = "<p>Resolution: <span class='text-primary'>" + response["probeInfo"]["resolution"] + "</span></p><p>Frame Rate: <span class='text-primary'>" + response["probeInfo"]["frameRate"] + " FPS</span></p><p>Audio: <span class='text-primary'>" + response["probeInfo"]["audioChannel"] + "</span></p>";
                    }
                }
            }
            if (response.hasOwnProperty("logoURL")) {
                var div = document.getElementById("channel-icon");
                div.value = response["logoURL"];
                div.className = "changed";
                return;
            }
            switch (data["cmd"]) {
                case "updateLog":
                    mergeUpdateLogResponse(response);
                    refreshOverviewOperationalState(SERVER);
                    refreshActivityOperationalState(SERVER);
                    if (document.getElementById("content_log")) {
                        showLogs(false);
                    }
                    return;
                    break;
                default:
                    SERVER = new Object();
                    SERVER = response;
                    break;
            }
            if (response.hasOwnProperty("openMenu")) {
                openLegacyMenu(response["openMenu"]);
                showElement("popup", false);
            }
            if (response.hasOwnProperty("openLink")) {
                window.location = response["openLink"];
            }
            if (response.hasOwnProperty("alert")) {
                alert(response["alert"]);
            }
            if (response.hasOwnProperty("reload")) {
                if (data["cmd"] == "saveWizard" && typeof completeConfigurationWizard == "function") {
                    completeConfigurationWizard();
                    return;
                }
                location.reload();
            }
            if (response.hasOwnProperty("wizard")) {
                createLayout();
                showConfigurationWizard(response["wizard"]);
                return;
            }
            createLayout();
        };
    }
}
function completeTask5RequestFailure(command, data, message) {
    var response = { status: false, err: message };
    if (typeof completeSourceRequest == "function") {
        completeSourceRequest(command, data, response);
    }
    if (typeof completeFilterRequest == "function") {
        completeFilterRequest(command, data, response);
    }
    if (command == "saveWizard" && typeof completeConfigurationWizardRequest == "function") {
        completeConfigurationWizardRequest(response);
    }
}
function mergeUpdateLogResponse(response) {
    var server = overviewRecord(SERVER);
    var update = overviewRecord(response);
    mergeUpdateLogRecord(server, "clientInfo", update.clientInfo);
    mergeUpdateLogRecord(server, "log", update.log);
}
function mergeUpdateLogRecord(server, key, updateValue) {
    var existing = overviewRecord(server[key]);
    if (server[key] !== existing) {
        server[key] = existing;
    }
    var update = overviewRecord(updateValue);
    Object.keys(update).forEach(updateKey => {
        existing[updateKey] = update[updateKey];
    });
}
function getCookie(name) {
    var value = "; " + document.cookie;
    var parts = value.split("; " + name + "=");
    if (parts.length == 2)
        return parts.pop().split(";").shift();
}
