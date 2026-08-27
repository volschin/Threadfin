"use strict";
class Log {
    createLog(entry) {
        return createLogEntryElement(entry);
    }
}
function showLogs(bottom) {
    var logs = SERVER["log"] && Array.isArray(SERVER["log"]["log"]) ? SERVER["log"]["log"] : [];
    var div = document.getElementById("content_log_entries");
    if (!div) {
        return;
    }
    renderLogEntryList(div, logs);
    setTimeout(function () {
        if (bottom == true) {
            var wrapper = document.getElementById("box-wrapper");
            wrapper.scrollTop = wrapper.scrollHeight;
        }
    }, 10);
}
function resetLogs() {
    if (!confirm("Reset all current log entries? This permanently removes them and cannot be undone.")) {
        return;
    }
    var cmd = "resetLogs";
    var data = new Object();
    var server = new Server(cmd);
    server.request(data);
}
