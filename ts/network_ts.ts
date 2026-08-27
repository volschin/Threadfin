type QueuedThreadfinRequest = {
  command: string
  data: any
  requestId: string
  sent: boolean
  settled: boolean
  timeoutId: any
}

class ThreadfinConnection {
  socket: WebSocket = null
  queue: QueuedThreadfinRequest[] = []
  active: QueuedThreadfinRequest = null
  nextRequestId: number = 1
  policyRejected: boolean = false
  reconnectTimeoutId: any = null
  preOpenFailures: number = 0
  transportFailureInProgress: boolean = false

  enqueue(command: string, data: any): void {
    var requestId = "request-" + this.nextRequestId
    this.nextRequestId += 1
    data["requestId"] = requestId
    var request: QueuedThreadfinRequest = {
      command: command,
      data: data,
      requestId: requestId,
      sent: false,
      settled: false,
      timeoutId: null,
    }
    if (this.policyRejected) {
      request.settled = true
      completeThreadfinRequestFailure(request, "{{.sources.transportError}}", "transport")
      return
    }
    this.queue.push(request)
    this.pump()
  }

  connect(): void {
    if (this.policyRejected || this.socket !== null || this.reconnectTimeoutId !== null || (this.active === null && this.queue.length == 0)) {
      return
    }
    var protocol = window.location.protocol == "https:" ? "wss://" : "ws://"
    var port = window.location.port ? ":" + window.location.port : ""
    var socket: WebSocket
    try {
      socket = new WebSocket(protocol + window.location.hostname + port + "/data/")
    } catch (_error) {
      console.warn("WebSocket connection failed")
      this.settlePreOpenFailure(null)
      return
    }
    this.socket = socket
    var opened = false

    socket.onopen = () => {
      if (this.socket !== socket || this.policyRejected) {
        return
      }
      opened = true
      this.preOpenFailures = 0
      WS_AVAILABLE = true
      console.log("WebSocket connection opened")
      this.pump()
    }

    socket.onmessage = (event: MessageEvent) => {
      if (this.socket !== socket || this.policyRejected) {
        return
      }
      var response: any
      try {
        response = JSON.parse(event.data)
      } catch (_error) {
        this.settleProtocolFailure()
        return
      }
      this.settleResponse(response)
    }

    socket.onerror = () => {
      if (this.socket !== socket || this.policyRejected) {
        return
      }
      if (!opened) {
        this.settlePreOpenFailure(socket)
        return
      }
      this.settleTransportFailure()
    }

    socket.onclose = (event: CloseEvent) => {
      if (event.code == 1008) {
        this.rejectPolicyClose(socket)
        return
      }
      if (this.socket !== socket || this.policyRejected) {
        return
      }
      if (!opened) {
        this.settlePreOpenFailure(socket)
        return
      }
      this.settleTransportFailure()
    }
  }

  pump(): void {
    if (this.policyRejected || this.transportFailureInProgress || this.active !== null || this.queue.length == 0 || this.reconnectTimeoutId !== null) {
      return
    }
    if (this.socket === null) {
      this.connect()
      return
    }
    if (this.socket.readyState !== WebSocket.OPEN) {
      return
    }

    var request = this.queue.shift()
    this.active = request
    request.sent = true
    request.timeoutId = setTimeout(() => {
      if (this.active !== request || request.settled) {
        return
      }
      this.settleTransportFailure()
    }, 30000)
    try {
      this.socket.send(JSON.stringify(request.data))
    } catch (_error) {
      this.settleTransportFailure()
    }
  }

  settleResponse(response: any): void {
    var responseIsObject = response && typeof response == "object" && !Array.isArray(response)
    if (this.active === null || !responseIsObject || response["requestId"] !== this.active.requestId) {
      this.settleProtocolFailure()
      return
    }

    var request = this.active
    if (request.settled) {
      return
    }
    request.settled = true
    clearTimeout(request.timeoutId)
    request.timeoutId = null
    this.active = null
    showElement("loading", false)
    console.log("WebSocket response received:", request.command)
    completeThreadfinRequestResponse(request, response)
    this.pump()
  }

  settleTransportFailure(): void {
    var socket = this.socket
    this.socket = null
    this.transportFailureInProgress = true
    this.settleActiveFailure("{{.sources.transportError}}", "transport")
    this.closeSocket(socket)
    this.transportFailureInProgress = false
    this.scheduleReconnect()
  }

  settleProtocolFailure(): void {
    var socket = this.socket
    this.socket = null
    this.transportFailureInProgress = true
    this.settleActiveFailure("{{.sources.responseInvalid}}", "transport")
    this.closeSocket(socket)
    this.transportFailureInProgress = false
    this.scheduleReconnect()
  }

  settlePreOpenFailure(socket: WebSocket): void {
    if (this.policyRejected || (socket !== null && this.socket !== socket)) {
      return
    }
    if (socket !== null) {
      this.socket = null
      this.closeSocket(socket)
    }
    this.preOpenFailures += 1
    if (this.preOpenFailures >= 2) {
      this.rejectTerminalFailure(socket, true)
      return
    }
    this.scheduleReconnect()
  }

  settleActiveFailure(message: string, failureKind: "busy" | "transport"): void {
    var request = this.active
    if (request === null || request.settled) {
      return
    }
    request.settled = true
    clearTimeout(request.timeoutId)
    request.timeoutId = null
    this.active = null
    showElement("loading", false)
    console.warn("WebSocket request failed:", request.command)
    completeThreadfinRequestFailure(request, message, failureKind)
  }

  closeSocket(socket: WebSocket): void {
    if (socket && socket.readyState !== WebSocket.CLOSED) {
      socket.close()
    }
  }

  scheduleReconnect(): void {
    if (this.policyRejected || this.reconnectTimeoutId !== null || this.active !== null || this.queue.length == 0) {
      return
    }
    this.reconnectTimeoutId = setTimeout(() => {
      this.reconnectTimeoutId = null
      this.pump()
    }, 250)
  }

  rejectPolicyClose(closedSocket: WebSocket): void {
    this.rejectTerminalFailure(closedSocket, false)
  }

  rejectTerminalFailure(closedSocket: WebSocket, showUnavailableAlert: boolean): void {
    if (this.policyRejected) {
      return
    }
    this.policyRejected = true
    clearTimeout(this.reconnectTimeoutId)
    this.reconnectTimeoutId = null
    this.preOpenFailures = 0
    var currentSocket = this.socket
    this.socket = null
    this.settleActiveFailure("{{.sources.transportError}}", "transport")
    var pending = this.queue
    this.queue = []
    pending.forEach(request => {
      if (!request.settled) {
        request.settled = true
        completeThreadfinRequestFailure(request, "{{.sources.transportError}}", "transport")
      }
    })
    if (currentSocket !== closedSocket) {
      this.closeSocket(currentSocket)
    }
    if (showUnavailableAlert) {
      alert("No websocket connection to Threadfin could be established. Check your network configuration.")
    }
    location.reload()
  }
}

var THREADFIN_CONNECTION = new ThreadfinConnection()

class Server {
  cmd: string

  constructor(cmd: string) {
    this.cmd = cmd
  }

  request(data: Object): any {
    if (this.cmd != "updateLog") {
      // showElement("loading", true)
      UNDO = new Object()
    }
    var requestData: any = Object.assign({}, data)
    requestData["cmd"] = this.cmd
    THREADFIN_CONNECTION.enqueue(this.cmd, requestData)
  }
}

function completeThreadfinRequestResponse(request: QueuedThreadfinRequest, response: any): void {
  var command = request.command
  var data = request.data
  if (typeof completeSourceRequest == "function") {
    completeSourceRequest(command, data, response)
  }
  if (typeof completeFilterRequest == "function") {
    completeFilterRequest(command, data, response)
  }
  if (command == "saveWizard" && typeof completeConfigurationWizardRequest == "function") {
    completeConfigurationWizardRequest(response)
  }
  if (typeof completeMappingRequest == "function") {
    completeMappingRequest(command, data, response)
  }

  applyThreadfinResponse(command, data, response)
}

function completeThreadfinRequestFailure(request: QueuedThreadfinRequest, message: string, failureKind: "busy" | "transport"): void {
  var command = request.command
  var data = request.data
  var response = { status: false, err: message, requestId: request.requestId }
  if (typeof completeSourceRequest == "function") {
    completeSourceRequest(command, data, response)
  }
  if (typeof completeFilterRequest == "function") {
    completeFilterRequest(command, data, response)
  }
  if (command == "saveWizard" && typeof completeConfigurationWizardRequest == "function") {
    completeConfigurationWizardRequest(response)
  }
  if (typeof completeMappingRequest == "function") {
    completeMappingRequest(command, data, response, failureKind)
  }
}

function applyThreadfinResponse(command: string, data: any, response: any): void {
  var responseIsObject = response && typeof response == "object" && !Array.isArray(response)
  if (!responseIsObject || response["status"] !== true) {
    if (command == "saveEpgMapping") {
      if (responseIsObject && response["xepg"] && response["xepg"]["epgMapping"] && SERVER["xepg"]) {
        SERVER["xepg"]["epgMapping"] = response["xepg"]["epgMapping"]
      }
      return
    }
    if (responseIsObject && response["status"] === false) {
      alert(response["err"] || "{{.sources.responseInvalid}}")
      if (response.hasOwnProperty("reload")) {
        location.reload()
      }
    } else {
      alert("{{.sources.responseInvalid}}")
    }
    return
  }

  if (response.hasOwnProperty("token")) {
    document.cookie = "Token=" + response["token"]
  }

  if (response.hasOwnProperty("probeInfo")) {
    var probeDetails = document.getElementById("probeDetails")
    if (probeDetails) {
      if (response["probeInfo"]["resolution"] !== undefined) {
        probeDetails.innerHTML = ""
        appendProbeDetail(probeDetails, "Resolution", response["probeInfo"]["resolution"], "")
        appendProbeDetail(probeDetails, "Frame Rate", response["probeInfo"]["frameRate"], " FPS")
        appendProbeDetail(probeDetails, "Audio", response["probeInfo"]["audioChannel"], "")
      }
    }
  }

  if (response.hasOwnProperty("logoURL")) {
    var div = (document.getElementById("channel-icon") as HTMLInputElement)
    div.value = response["logoURL"]
    div.className = "changed"
    return
  }

  switch (command) {
    case "updateLog":
      mergeUpdateLogResponse(response)
      refreshOverviewOperationalState(SERVER)
      refreshActivityOperationalState(SERVER)
      if (document.getElementById("content_log")) {
        showLogs(false)
      }
      return

    default:
      SERVER = new Object()
      SERVER = response
      break
  }

  if (response.hasOwnProperty("openMenu")) {
    openLegacyMenu(response["openMenu"])
    showElement("popup", false)
  }

  if (response.hasOwnProperty("openLink")) {
    window.location = response["openLink"]
  }

  if (response.hasOwnProperty("alert")) {
    alert(response["alert"])
  }

  if (response.hasOwnProperty("reload")) {
    if (command == "saveWizard" && typeof completeConfigurationWizard == "function") {
      completeConfigurationWizard()
      return
    }
    location.reload()
  }

  if (response.hasOwnProperty("wizard")) {
    createLayout()
    showConfigurationWizard(response["wizard"])
    return
  }

  createLayout()
}

function appendProbeDetail(container: HTMLElement, label: string, value: any, suffix: string): void {
  var row = document.createElement("p")
  row.appendChild(document.createTextNode(label + ": "))
  var detail = document.createElement("span")
  detail.className = "text-primary"
  detail.textContent = String(value) + suffix
  row.appendChild(detail)
  container.appendChild(row)
}

function mergeUpdateLogResponse(response: any): void {
  var server = overviewRecord(SERVER)
  var update = overviewRecord(response)
  mergeUpdateLogRecord(server, "clientInfo", update.clientInfo)
  mergeUpdateLogRecord(server, "log", update.log)
}

function mergeUpdateLogRecord(server: { [key: string]: any }, key: string, updateValue: any): void {
  var existing = overviewRecord(server[key])
  if (server[key] !== existing) {
    server[key] = existing
  }
  var update = overviewRecord(updateValue)
  Object.keys(update).forEach(updateKey => {
    existing[updateKey] = update[updateKey]
  })
}
