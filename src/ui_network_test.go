package src

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

func TestGeneratedWebSocketQueue(t *testing.T) {
	temp := t.TempDir()
	scriptPath := filepath.Join(temp, "websocket-queue.js")
	if err := os.WriteFile(scriptPath, []byte(webSocketQueueNodeScript), 0o600); err != nil {
		t.Fatal(err)
	}

	output, err := exec.Command("node", scriptPath, filepath.Join("..", "html", "js", "network_ts.js")).CombinedOutput()
	if err != nil {
		t.Fatalf("execute generated WebSocket queue contract: %v\n%s", err, output)
	}

	var got struct {
		Happy struct {
			URL                  string   `json:"url"`
			SocketCount          int      `json:"socketCount"`
			InitialSendCount     int      `json:"initialSendCount"`
			QueuedBeforeReply    int      `json:"queuedBeforeReply"`
			CallerInputUnchanged bool     `json:"callerInputUnchanged"`
			Commands             []string `json:"commands"`
			RequestIDs           []string `json:"requestIds"`
			SendsAfterReplies    []int    `json:"sendsAfterReplies"`
			CompletionCounts     []int    `json:"completionCounts"`
		} `json:"happy"`
		Mismatch struct {
			FirstFailedOnce    bool     `json:"firstFailedOnce"`
			Reconnected        bool     `json:"reconnected"`
			ReplayCommands     []string `json:"replayCommands"`
			SecondStillPending bool     `json:"secondStillPending"`
		} `json:"mismatch"`
		Transport struct {
			FirstFailedOnce   bool     `json:"firstFailedOnce"`
			Reconnected       bool     `json:"reconnected"`
			ReconnectCommands []string `json:"reconnectCommands"`
			ThirdStillPending bool     `json:"thirdStillPending"`
		} `json:"transport"`
		Timeout struct {
			Delay             int      `json:"delay"`
			FirstFailedOnce   bool     `json:"firstFailedOnce"`
			ReconnectCommands []string `json:"reconnectCommands"`
		} `json:"timeout"`
		Policy struct {
			FailureCounts []int `json:"failureCounts"`
			Reloads       int   `json:"reloads"`
			SocketCount   int   `json:"socketCount"`
			QueueLength   int   `json:"queueLength"`
			ActiveCleared bool  `json:"activeCleared"`
		} `json:"policy"`
	}
	if err := json.Unmarshal(output, &got); err != nil {
		t.Fatalf("decode generated WebSocket queue contract: %v\n%s", err, output)
	}

	if got.Happy.URL != "ws://127.0.0.1:34400/data/" {
		t.Errorf("WebSocket URL = %q, want credential-free /data/ URL", got.Happy.URL)
	}
	if got.Happy.SocketCount != 1 || got.Happy.InitialSendCount != 1 || got.Happy.QueuedBeforeReply != 2 {
		t.Errorf("initial queue state = sockets %d sends %d queued %d, want 1 / 1 / 2", got.Happy.SocketCount, got.Happy.InitialSendCount, got.Happy.QueuedBeforeReply)
	}
	if !got.Happy.CallerInputUnchanged {
		t.Error("Server.request mutated the caller's input while adding cmd/requestId")
	}
	if !reflect.DeepEqual(got.Happy.Commands, []string{"first", "second", "third"}) {
		t.Errorf("FIFO commands = %v, want first second third", got.Happy.Commands)
	}
	if len(got.Happy.RequestIDs) != 3 || got.Happy.RequestIDs[0] == "" || got.Happy.RequestIDs[0] == got.Happy.RequestIDs[1] || got.Happy.RequestIDs[1] == got.Happy.RequestIDs[2] || got.Happy.RequestIDs[0] == got.Happy.RequestIDs[2] {
		t.Errorf("request IDs = %v, want three unique non-empty IDs", got.Happy.RequestIDs)
	}
	if !reflect.DeepEqual(got.Happy.SendsAfterReplies, []int{1, 2, 3}) {
		t.Errorf("send counts after replies = %v, want one-at-a-time release [1 2 3]", got.Happy.SendsAfterReplies)
	}
	if !reflect.DeepEqual(got.Happy.CompletionCounts, []int{1, 1, 1}) {
		t.Errorf("successful request completion counts = %v, want exactly once each", got.Happy.CompletionCounts)
	}
	if !got.Mismatch.FirstFailedOnce || !got.Mismatch.Reconnected || !reflect.DeepEqual(got.Mismatch.ReplayCommands, []string{"second"}) || !got.Mismatch.SecondStillPending {
		t.Errorf("mismatched-ID handling = %+v, want active failed once and only unsent second command on a fresh socket", got.Mismatch)
	}
	if !got.Transport.FirstFailedOnce || !got.Transport.Reconnected || !reflect.DeepEqual(got.Transport.ReconnectCommands, []string{"second"}) || !got.Transport.ThirdStillPending {
		t.Errorf("error-plus-close handling = %+v, want one settlement, no replay, and retained unsent work", got.Transport)
	}
	if got.Timeout.Delay != 30000 || !got.Timeout.FirstFailedOnce || !reflect.DeepEqual(got.Timeout.ReconnectCommands, []string{"second"}) {
		t.Errorf("timeout handling = %+v, want 30000 ms, one settlement, and only unsent work reconnected", got.Timeout)
	}
	if !reflect.DeepEqual(got.Policy.FailureCounts, []int{1, 1, 1}) || got.Policy.Reloads != 1 || got.Policy.SocketCount != 1 || got.Policy.QueueLength != 0 || !got.Policy.ActiveCleared {
		t.Errorf("policy-close handling = %+v, want queue rejected once, no reconnect, and reload", got.Policy)
	}
}

const webSocketQueueNodeScript = `
const fs = require("fs");
const vm = require("vm");

function makeHarness() {
  const sockets = [];
  const completions = [];
  const timers = [];
  let reloads = 0;

  class FakeWebSocket {
    constructor(url) {
      this.url = url;
      this.readyState = FakeWebSocket.CONNECTING;
      this.OPEN = FakeWebSocket.OPEN;
      this.sent = [];
      this.closeCount = 0;
      sockets.push(this);
    }
    send(value) {
      if (this.readyState !== FakeWebSocket.OPEN) throw new Error("send while socket is not open");
      this.sent.push(value);
    }
    open() {
      this.readyState = FakeWebSocket.OPEN;
      if (this.onopen) this.onopen.call(this, {});
    }
    close(code = 1000) {
      this.closeCount += 1;
      if (this.readyState === FakeWebSocket.CLOSED) return;
      this.readyState = FakeWebSocket.CLOSED;
      if (this.onclose) this.onclose.call(this, {code});
    }
    emitClose(code) {
      this.readyState = FakeWebSocket.CLOSED;
      if (this.onclose) this.onclose.call(this, {code});
    }
    emitError() {
      if (this.onerror) this.onerror.call(this, {});
    }
    emitMessage(value) {
      if (this.onmessage) this.onmessage.call(this, {data: typeof value === "string" ? value : JSON.stringify(value)});
    }
  }
  FakeWebSocket.CONNECTING = 0;
  FakeWebSocket.OPEN = 1;
  FakeWebSocket.CLOSING = 2;
  FakeWebSocket.CLOSED = 3;

  const context = {
    console: {log() {}, warn() {}},
    document: {cookie: "Token=credential-marker", getElementById() { return null; }},
    window: {location: {protocol: "http:", hostname: "127.0.0.1", port: "34400"}},
    location: {reload() { reloads += 1; }},
    WebSocket: FakeWebSocket,
    SERVER: {},
    UNDO: {},
    SERVER_CONNECTION: false,
    WS_AVAILABLE: false,
    showElement() {},
    createLayout() {},
    alert() {},
    completeSourceRequest(command, data, response) {
      completions.push({command, requestId: data.requestId, status: response && response.status, err: response && response.err});
    },
    setTimeout(callback, delay) {
      const timer = {callback, delay, cleared: false};
      timers.push(timer);
      return timer;
    },
    clearTimeout(timer) { if (timer) timer.cleared = true; },
  };
  vm.createContext(context);
  vm.runInContext(fs.readFileSync(process.argv[2], "utf8")
    .replaceAll("{{.sources.requestBusy}}", "busy")
    .replaceAll("{{.sources.transportError}}", "transport")
    .replaceAll("{{.sources.responseInvalid}}", "invalid"), context);
  return {context, sockets, completions, timers, reloads: () => reloads};
}

function request(harness, command, data) {
  harness.context.command = command;
  harness.context.requestData = data;
  vm.runInContext("new Server(command).request(requestData)", harness.context);
}

function sent(socket) {
  return socket ? socket.sent.map(value => JSON.parse(value)) : [];
}

function completionCount(harness, command) {
  return harness.completions.filter(item => item.command === command).length;
}

const happyHarness = makeHarness();
const firstInput = {value: "one"};
request(happyHarness, "first", firstInput);
request(happyHarness, "second", {value: "two"});
request(happyHarness, "third", {value: "three"});
const happySocket = happyHarness.sockets[0];
if (happySocket) happySocket.open();
const happyInitial = sent(happySocket);
const queuedBeforeReply = happyHarness.context.THREADFIN_CONNECTION ? happyHarness.context.THREADFIN_CONNECTION.queue.length : -1;
const sendsAfterReplies = [happyInitial.length];
if (happyInitial[0]) {
  happySocket.emitMessage({status: true, requestId: happyInitial[0].requestId});
  sendsAfterReplies.push(sent(happySocket).length);
}
const happyAfterFirst = sent(happySocket);
if (happyAfterFirst[1]) {
  happySocket.emitMessage({status: true, requestId: happyAfterFirst[1].requestId});
  sendsAfterReplies.push(sent(happySocket).length);
}
const happyAfterSecond = sent(happySocket);
if (happyAfterSecond[2]) {
  happySocket.emitMessage({status: true, requestId: happyAfterSecond[2].requestId});
}
const happySent = sent(happySocket);

const mismatchHarness = makeHarness();
request(mismatchHarness, "first", {});
request(mismatchHarness, "second", {});
const mismatchFirstSocket = mismatchHarness.sockets[0];
if (mismatchFirstSocket) mismatchFirstSocket.open();
const mismatchFirstSent = sent(mismatchFirstSocket)[0];
if (mismatchFirstSent) mismatchFirstSocket.emitMessage({status: true, requestId: "wrong-request-id"});
const mismatchSecondSocket = mismatchHarness.sockets[1];
if (mismatchSecondSocket) mismatchSecondSocket.open();
const mismatchReplay = sent(mismatchSecondSocket);

const transportHarness = makeHarness();
request(transportHarness, "first", {});
request(transportHarness, "second", {});
request(transportHarness, "third", {});
const transportFirstSocket = transportHarness.sockets[0];
if (transportFirstSocket) transportFirstSocket.open();
transportFirstSocket.emitError();
transportFirstSocket.emitClose(1006);
const transportSecondSocket = transportHarness.sockets[1];
if (transportSecondSocket) transportSecondSocket.open();
const transportReconnect = sent(transportSecondSocket);

const timeoutHarness = makeHarness();
request(timeoutHarness, "first", {});
request(timeoutHarness, "second", {});
const timeoutFirstSocket = timeoutHarness.sockets[0];
if (timeoutFirstSocket) timeoutFirstSocket.open();
const activeTimer = timeoutHarness.timers.find(timer => !timer.cleared);
if (activeTimer) activeTimer.callback();
const timeoutSecondSocket = timeoutHarness.sockets[1];
if (timeoutSecondSocket) timeoutSecondSocket.open();
const timeoutReconnect = sent(timeoutSecondSocket);

const policyHarness = makeHarness();
request(policyHarness, "first", {});
request(policyHarness, "second", {});
request(policyHarness, "third", {});
const policySocket = policyHarness.sockets[0];
if (policySocket) {
  policySocket.open();
  policySocket.emitClose(1008);
}

process.stdout.write(JSON.stringify({
  happy: {
    url: happySocket ? happySocket.url : "",
    socketCount: happyHarness.sockets.length,
    initialSendCount: happyInitial.length,
    queuedBeforeReply,
    callerInputUnchanged: JSON.stringify(firstInput) === '{"value":"one"}',
    commands: happySent.map(message => message.cmd),
    requestIds: happySent.map(message => message.requestId || ""),
    sendsAfterReplies,
    completionCounts: ["first", "second", "third"].map(command => completionCount(happyHarness, command)),
  },
  mismatch: {
    firstFailedOnce: completionCount(mismatchHarness, "first") === 1 && mismatchHarness.completions.find(item => item.command === "first").status === false,
    reconnected: mismatchHarness.sockets.length === 2,
    replayCommands: mismatchReplay.map(message => message.cmd),
    secondStillPending: completionCount(mismatchHarness, "second") === 0,
  },
  transport: {
    firstFailedOnce: completionCount(transportHarness, "first") === 1,
    reconnected: transportHarness.sockets.length === 2,
    reconnectCommands: transportReconnect.map(message => message.cmd),
    thirdStillPending: completionCount(transportHarness, "third") === 0 && transportReconnect.length === 1,
  },
  timeout: {
    delay: activeTimer ? activeTimer.delay : -1,
    firstFailedOnce: completionCount(timeoutHarness, "first") === 1,
    reconnectCommands: timeoutReconnect.map(message => message.cmd),
  },
  policy: {
    failureCounts: ["first", "second", "third"].map(command => completionCount(policyHarness, command)),
    reloads: policyHarness.reloads(),
    socketCount: policyHarness.sockets.length,
    queueLength: policyHarness.context.THREADFIN_CONNECTION ? policyHarness.context.THREADFIN_CONNECTION.queue.length : -1,
    activeCleared: policyHarness.context.THREADFIN_CONNECTION ? policyHarness.context.THREADFIN_CONNECTION.active === null : false,
  },
}));
`
