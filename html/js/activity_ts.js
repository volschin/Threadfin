"use strict";
function renderActivity(server) {
    refreshActivityOperationalState(server);
    if (activityHasStreamPreview(server) && typeof showPreview == "function") {
        showPreview(true);
    }
}
function refreshActivityOperationalState(server) {
    var activity = selectActivityState(server);
    setActivityText("playlist-connection-information", activityCapacityLabel("Playlist source connections", activity.playlists));
    setActivityText("client-connection-information", activityCapacityLabel("Client connections", activity.clients));
}
function activityCapacityLabel(label, capacity) {
    return label + ": " + capacity.active + " / " + capacity.total;
}
function setActivityText(id, value) {
    var element = document.getElementById(id);
    if (element && element.textContent != value) {
        element.textContent = value;
    }
}
function activityHasStreamPreview(server) {
    var root = overviewRecord(server);
    var data = overviewRecord(root.data);
    var preview = overviewRecord(data.StreamPreviewUI);
    return Array.isArray(preview.activeStreams) && Array.isArray(preview.inactiveStreams);
}
