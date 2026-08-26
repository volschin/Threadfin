function renderActivity(server: any): void {
  refreshActivityOperationalState(server)
  if (activityHasStreamPreview(server) && typeof showPreview == "function") {
    showPreview(true)
  }
}

function refreshActivityOperationalState(server: any): void {
  var activity = selectActivityState(server)
  setActivityText("playlist-connection-information", activityCapacityLabel("Playlist source connections", activity.playlists))
  setActivityText("client-connection-information", activityCapacityLabel("Client connections", activity.clients))
}

function activityCapacityLabel(label: string, capacity: OverviewCapacityState): string {
  return label + ": " + capacity.active + " / " + capacity.total
}

function setActivityText(id: string, value: string): void {
  var element = document.getElementById(id)
  if (element && element.textContent != value) {
    element.textContent = value
  }
}

function activityHasStreamPreview(server: any): boolean {
  var root = overviewRecord(server)
  var data = overviewRecord(root.data)
  var preview = overviewRecord(data.StreamPreviewUI)
  return Array.isArray(preview.activeStreams) && Array.isArray(preview.inactiveStreams)
}
