function renderConnections(server: any): void {
  var host = document.getElementById("connections-content")
  if (!host) {
    return
  }

  var endpoints = selectOverviewState(server).outputs.endpoints
  while (host.firstChild) {
    host.removeChild(host.firstChild)
  }

  var connections = document.createElement("div")
  connections.className = "tf-connections"

  var header = document.createElement("header")
  header.className = "tf-operations-header"
  var title = document.createElement("h1")
  title.textContent = "Connections"
  var introduction = document.createElement("p")
  introduction.textContent = "Copy an available Threadfin endpoint into a supported client. Client setup and success are verified in that client, not by Threadfin."
  header.appendChild(title)
  header.appendChild(introduction)
  connections.appendChild(header)

  var guidance = document.createElement("section")
  guidance.className = "tf-operations-panel tf-connections-guidance"
  var guidanceHeading = document.createElement("h2")
  guidanceHeading.textContent = "Choose an endpoint"
  var guidanceText = document.createElement("p")
  guidanceText.textContent = "Use the DVR address for tuner discovery. Use M3U for the generated channel playlist and XMLTV for generated guide data when those outputs are available."
  guidance.appendChild(guidanceHeading)
  guidance.appendChild(guidanceText)
  connections.appendChild(guidance)

  var endpointSection = document.createElement("section")
  endpointSection.className = "tf-operations-panel"
  var endpointHeading = document.createElement("h2")
  endpointHeading.textContent = "Threadfin endpoints"
  endpointSection.appendChild(endpointHeading)
  var endpointList = document.createElement("ul")
  endpointList.className = "tf-connections-endpoints"
  endpoints.forEach(endpoint => endpointList.appendChild(createConnectionsEndpoint(endpoint)))
  endpointSection.appendChild(endpointList)
  connections.appendChild(endpointSection)

  var clientNotes = document.createElement("section")
  clientNotes.className = "tf-operations-panel tf-client-notes"
  var clientNotesHeading = document.createElement("h2")
  clientNotesHeading.textContent = "Client-specific notes"
  var clientNotesIntroduction = document.createElement("p")
  clientNotesIntroduction.textContent = "These notes identify where to start. Complete and verify configuration in the selected client."
  clientNotes.appendChild(clientNotesHeading)
  clientNotes.appendChild(clientNotesIntroduction)
  clientNotes.appendChild(createConnectionsClientNote("Plex", "Use the DVR address when adding Threadfin as a tuner. Complete the remaining guide and channel setup in Plex."))
  clientNotes.appendChild(createConnectionsClientNote("Jellyfin", "Choose the DVR address or the available M3U and XMLTV endpoints for the setup flow supported by your Jellyfin version. Verify the result in Jellyfin."))
  clientNotes.appendChild(createConnectionsClientNote("Emby", "Choose the DVR address or the available M3U and XMLTV endpoints for the setup flow supported by your Emby version. Verify the result in Emby."))
  connections.appendChild(clientNotes)

  var copyStatus = document.createElement("p")
  copyStatus.id = "connections-copy-status"
  copyStatus.className = "tf-overview-status"
  copyStatus.setAttribute("role", "status")
  copyStatus.setAttribute("aria-live", "polite")
  copyStatus.setAttribute("aria-atomic", "true")
  connections.appendChild(copyStatus)

  host.appendChild(connections)
}

function createConnectionsEndpoint(endpoint: OverviewEndpointState): HTMLLIElement {
  var item = document.createElement("li")
  item.className = "tf-connection-endpoint"
  item.setAttribute("data-endpoint", endpoint.key)
  item.setAttribute("data-available", String(endpoint.available))

  var heading = document.createElement("h3")
  heading.textContent = endpoint.label
  item.appendChild(heading)

  if (endpoint.available) {
    var value = document.createElement("code")
    value.className = "tf-endpoint-value"
    value.textContent = endpoint.value
    item.appendChild(value)

    var explanation = document.createElement("p")
    explanation.textContent = endpoint.explanation
    item.appendChild(explanation)

    var copy = document.createElement("button")
    copy.type = "button"
    copy.className = "tf-overview-action tf-copy-action"
    copy.textContent = "Copy " + endpoint.label
    bindOverviewCopyAction(copy, endpoint.value, endpoint.label, "connections-copy-status")
    item.appendChild(copy)
  } else {
    var unavailable = document.createElement("p")
    unavailable.className = "tf-endpoint-unavailable"
    unavailable.textContent = "Unavailable — " + endpoint.explanation
    item.appendChild(unavailable)
  }

  return item
}

function createConnectionsClientNote(client: string, note: string): HTMLElement {
  var details = document.createElement("details")
  var summary = document.createElement("summary")
  summary.textContent = client
  var explanation = document.createElement("p")
  explanation.textContent = note
  details.appendChild(summary)
  details.appendChild(explanation)
  return details
}
