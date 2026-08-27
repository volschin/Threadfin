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
    if (this.policyRejected || this.socket !== null || (this.active === null && this.queue.length == 0)) {
      return
    }
    var protocol = window.location.protocol == "https:" ? "wss://" : "ws://"
    var port = window.location.port ? ":" + window.location.port : ""
    var socket = new WebSocket(protocol + window.location.hostname + port + "/data/")
    this.socket = socket

    socket.onopen = () => {
      if (this.socket !== socket || this.policyRejected) {
        return
      }
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
      var unavailable = WS_AVAILABLE == false
      this.settleTransportFailure()
      if (unavailable) {
        alert("No websocket connection to Threadfin could be established. Check your network configuration.")
      }
    }

    socket.onclose = (event: CloseEvent) => {
      if (event.code == 1008) {
        this.rejectPolicyClose(socket)
        return
      }
      if (this.socket !== socket || this.policyRejected) {
        return
      }
      this.settleTransportFailure()
    }
  }

  pump(): void {
    if (this.policyRejected || this.active !== null || this.queue.length == 0) {
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
    this.settleActiveFailure("{{.sources.transportError}}", "transport")
    this.closeSocket(socket)
    this.pump()
  }

  settleProtocolFailure(): void {
    var socket = this.socket
    this.socket = null
    this.settleActiveFailure("{{.sources.responseInvalid}}", "transport")
    this.closeSocket(socket)
    this.pump()
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

  rejectPolicyClose(closedSocket: WebSocket): void {
    if (this.policyRejected) {
      return
    }
    this.policyRejected = true
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
    if (document.getElementById("probeDetails")) {
      if (response["probeInfo"]["resolution"] !== undefined) {
        document.getElementById("probeDetails").innerHTML = "<p>Resolution: <span class='text-primary'>" + response["probeInfo"]["resolution"] + "</span></p><p>Frame Rate: <span class='text-primary'>" + response["probeInfo"]["frameRate"] + " FPS</span></p><p>Audio: <span class='text-primary'>" + response["probeInfo"]["audioChannel"] + "</span></p>"
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
