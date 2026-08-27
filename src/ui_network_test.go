package src

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

type webSocketPreOpenTerminalResult struct {
	Attempts                 int   `json:"attempts"`
	AttemptsBeforeRetry      int   `json:"attemptsBeforeRetry"`
	PendingTimersBeforeRetry int   `json:"pendingTimersBeforeRetry"`
	RetryDelay               int   `json:"retryDelay"`
	CompletionCounts         []int `json:"completionCounts"`
	Alerts                   int   `json:"alerts"`
	Reloads                  int   `json:"reloads"`
	QueueLength              int   `json:"queueLength"`
	PendingTimers            int   `json:"pendingTimers"`
	PostFailureCount         int   `json:"postFailureCount"`
	AttemptsAfterPost        int   `json:"attemptsAfterPost"`
	SocketCleared            bool  `json:"socketCleared"`
	ActiveCleared            bool  `json:"activeCleared"`
	CurrentSocketClosed      bool  `json:"currentSocketClosed"`
}

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
			NoSynchronousRetry bool     `json:"noSynchronousRetry"`
			RetryDelay         int      `json:"retryDelay"`
			PendingRetryTimers int      `json:"pendingRetryTimers"`
			Reconnected        bool     `json:"reconnected"`
			ReplayCommands     []string `json:"replayCommands"`
			SecondStillPending bool     `json:"secondStillPending"`
		} `json:"mismatch"`
		ConstructorFailure struct {
			EscapedErrors       int      `json:"escapedErrors"`
			AttemptsBeforeRetry int      `json:"attemptsBeforeRetry"`
			AlertsBeforeRetry   int      `json:"alertsBeforeRetry"`
			PendingRetryTimers  int      `json:"pendingRetryTimers"`
			RetryDelay          int      `json:"retryDelay"`
			QueueBeforeRetry    []string `json:"queueBeforeRetry"`
			SocketsBeforeRetry  int      `json:"socketsBeforeRetry"`
			SentAfterRetry      []string `json:"sentAfterRetry"`
			CompletionCounts    []int    `json:"completionCounts"`
		} `json:"constructorFailure"`
		SendThrow struct {
			EscapedErrors      int      `json:"escapedErrors"`
			FailureCounts      []int    `json:"failureCounts"`
			NoSynchronousRetry bool     `json:"noSynchronousRetry"`
			RetryDelay         int      `json:"retryDelay"`
			ReconnectCommands  []string `json:"reconnectCommands"`
			SocketCount        int      `json:"socketCount"`
		} `json:"sendThrow"`
		PreOpen struct {
			CloseCommands     []string `json:"closeCommands"`
			CloseCompleted    []int    `json:"closeCompleted"`
			ErrorCommands     []string `json:"errorCommands"`
			ErrorCompleted    []int    `json:"errorCompleted"`
			NoEarlyFailures   bool     `json:"noEarlyFailures"`
			DelayedRetries    bool     `json:"delayedRetries"`
			SingleRetryTimers bool     `json:"singleRetryTimers"`
		} `json:"preOpen"`
		Transport struct {
			FirstFailedOnce    bool     `json:"firstFailedOnce"`
			NoSynchronousRetry bool     `json:"noSynchronousRetry"`
			RetryDelay         int      `json:"retryDelay"`
			Reconnected        bool     `json:"reconnected"`
			ReconnectCommands  []string `json:"reconnectCommands"`
			CompletionCounts   []int    `json:"completionCounts"`
		} `json:"transport"`
		Timeout struct {
			Delay                int      `json:"delay"`
			FirstFailedOnce      bool     `json:"firstFailedOnce"`
			ReconnectCommands    []string `json:"reconnectCommands"`
			LateMessageIgnored   bool     `json:"lateMessageIgnored"`
			LateErrorIgnored     bool     `json:"lateErrorIgnored"`
			LateCloseIgnored     bool     `json:"lateCloseIgnored"`
			CompletionCounts     []int    `json:"completionCounts"`
			ReplacementUnchanged bool     `json:"replacementUnchanged"`
		} `json:"timeout"`
		StalePolicy struct {
			ActiveBeforePolicy string   `json:"activeBeforePolicy"`
			QueueBeforePolicy  []string `json:"queueBeforePolicy"`
			FailureCounts      []int    `json:"failureCounts"`
			FailureStatuses    []bool   `json:"failureStatuses"`
			Reloads            int      `json:"reloads"`
			SocketCount        int      `json:"socketCount"`
			QueueLength        int      `json:"queueLength"`
			ActiveCleared      bool     `json:"activeCleared"`
			SocketCleared      bool     `json:"socketCleared"`
			ReplacementClosed  bool     `json:"replacementClosed"`
			NoReconnect        bool     `json:"noReconnect"`
		} `json:"stalePolicy"`
		CallbackFIFO struct {
			Commands         []string `json:"commands"`
			CompletionCounts []int    `json:"completionCounts"`
		} `json:"callbackFifo"`
		Policy struct {
			FailureCounts    []int `json:"failureCounts"`
			PostFailureCount int   `json:"postFailureCount"`
			Reloads          int   `json:"reloads"`
			SocketCount      int   `json:"socketCount"`
			QueueLength      int   `json:"queueLength"`
			ActiveCleared    bool  `json:"activeCleared"`
		} `json:"policy"`
		PreOpenTerminal struct {
			Constructor webSocketPreOpenTerminalResult `json:"constructor"`
			AsyncError  webSocketPreOpenTerminalResult `json:"asyncError"`
			AsyncClose  webSocketPreOpenTerminalResult `json:"asyncClose"`
		} `json:"preOpenTerminal"`
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
	if !got.Mismatch.FirstFailedOnce || !got.Mismatch.NoSynchronousRetry || got.Mismatch.RetryDelay != 250 || got.Mismatch.PendingRetryTimers != 1 || !got.Mismatch.Reconnected || !reflect.DeepEqual(got.Mismatch.ReplayCommands, []string{"second"}) || !got.Mismatch.SecondStillPending {
		t.Errorf("mismatched-ID handling = %+v, want active failed once and only unsent second command after one delayed reconnect", got.Mismatch)
	}
	if got.ConstructorFailure.EscapedErrors != 0 || got.ConstructorFailure.AttemptsBeforeRetry != 1 || got.ConstructorFailure.AlertsBeforeRetry != 0 || got.ConstructorFailure.PendingRetryTimers != 1 || got.ConstructorFailure.RetryDelay != 250 || !reflect.DeepEqual(got.ConstructorFailure.QueueBeforeRetry, []string{"first", "second"}) || got.ConstructorFailure.SocketsBeforeRetry != 0 || !reflect.DeepEqual(got.ConstructorFailure.SentAfterRetry, []string{"first", "second"}) || !reflect.DeepEqual(got.ConstructorFailure.CompletionCounts, []int{1, 1}) {
		t.Errorf("constructor failure transition = %+v, want contained error, deferred retry, and complete unsent FIFO", got.ConstructorFailure)
	}
	if got.SendThrow.EscapedErrors != 0 || !reflect.DeepEqual(got.SendThrow.FailureCounts, []int{1, 1}) || !got.SendThrow.NoSynchronousRetry || got.SendThrow.RetryDelay != 250 || !reflect.DeepEqual(got.SendThrow.ReconnectCommands, []string{"second"}) || got.SendThrow.SocketCount != 2 {
		t.Errorf("send throw handling = %+v, want active failed once, tail retained, and delayed reconnect without exception/replay", got.SendThrow)
	}
	if !reflect.DeepEqual(got.PreOpen.CloseCommands, []string{"first", "second"}) || !reflect.DeepEqual(got.PreOpen.CloseCompleted, []int{1, 1}) || !reflect.DeepEqual(got.PreOpen.ErrorCommands, []string{"first", "second"}) || !reflect.DeepEqual(got.PreOpen.ErrorCompleted, []int{1, 1}) || !got.PreOpen.NoEarlyFailures || !got.PreOpen.DelayedRetries || !got.PreOpen.SingleRetryTimers {
		t.Errorf("pre-open failure handling = %+v, want one delayed retry and complete unsent FIFO after close and error", got.PreOpen)
	}
	if !got.Transport.FirstFailedOnce || !got.Transport.NoSynchronousRetry || got.Transport.RetryDelay != 250 || !got.Transport.Reconnected || !reflect.DeepEqual(got.Transport.ReconnectCommands, []string{"second", "third"}) || !reflect.DeepEqual(got.Transport.CompletionCounts, []int{1, 1, 1}) {
		t.Errorf("error-plus-close handling = %+v, want one settlement, no replay, and complete FIFO tail after a delayed reconnect", got.Transport)
	}
	if got.Timeout.Delay != 30000 || !got.Timeout.FirstFailedOnce || !reflect.DeepEqual(got.Timeout.ReconnectCommands, []string{"second", "third"}) || !got.Timeout.LateMessageIgnored || !got.Timeout.LateErrorIgnored || !got.Timeout.LateCloseIgnored || !got.Timeout.ReplacementUnchanged || !reflect.DeepEqual(got.Timeout.CompletionCounts, []int{1, 1, 1}) {
		t.Errorf("timeout/late-event handling = %+v, want 30s, isolated stale events, and complete FIFO tail", got.Timeout)
	}
	if got.StalePolicy.ActiveBeforePolicy != "second" || !reflect.DeepEqual(got.StalePolicy.QueueBeforePolicy, []string{"third", "fourth"}) || !reflect.DeepEqual(got.StalePolicy.FailureCounts, []int{1, 1, 1, 1}) || !reflect.DeepEqual(got.StalePolicy.FailureStatuses, []bool{true, true, true, true}) || got.StalePolicy.Reloads != 1 || got.StalePolicy.SocketCount != 2 || got.StalePolicy.QueueLength != 0 || !got.StalePolicy.ActiveCleared || !got.StalePolicy.SocketCleared || !got.StalePolicy.ReplacementClosed || !got.StalePolicy.NoReconnect {
		t.Errorf("stale policy-close handling = %+v, want delayed stale 1008 to fail replacement active and full tail once, empty state, close replacement, reload once, and never reconnect", got.StalePolicy)
	}
	if !reflect.DeepEqual(got.CallbackFIFO.Commands, []string{"first", "tail", "callback"}) || !reflect.DeepEqual(got.CallbackFIFO.CompletionCounts, []int{1, 1, 1}) {
		t.Errorf("callback-enqueued FIFO = %+v, want existing tail before callback work", got.CallbackFIFO)
	}
	if !reflect.DeepEqual(got.Policy.FailureCounts, []int{1, 1, 1}) || got.Policy.PostFailureCount != 1 || got.Policy.Reloads != 1 || got.Policy.SocketCount != 1 || got.Policy.QueueLength != 0 || !got.Policy.ActiveCleared {
		t.Errorf("policy-close handling = %+v, want queue and post-policy work rejected once without reconnect", got.Policy)
	}
	assertPreOpenTerminal := func(name string, result webSocketPreOpenTerminalResult, wantSocketClosed bool) {
		t.Helper()
		if result.Attempts != 2 || result.AttemptsBeforeRetry != 1 || result.PendingTimersBeforeRetry != 1 || result.RetryDelay != 250 || !reflect.DeepEqual(result.CompletionCounts, []int{1, 1}) || result.Alerts != 1 || result.Reloads != 1 || result.QueueLength != 0 || result.PendingTimers != 0 || result.PostFailureCount != 1 || result.AttemptsAfterPost != 2 || !result.SocketCleared || !result.ActiveCleared || (wantSocketClosed && !result.CurrentSocketClosed) {
			t.Errorf("%s terminal pre-open handling = %+v, want exactly two attempts, one delayed retry, one settlement/alert/reload, cleared state, and no reconnect after terminal rejection", name, result)
		}
	}
	assertPreOpenTerminal("constructor", got.PreOpenTerminal.Constructor, false)
	assertPreOpenTerminal("async error", got.PreOpenTerminal.AsyncError, true)
	assertPreOpenTerminal("async close", got.PreOpenTerminal.AsyncClose, false)
}

const webSocketQueueNodeScript = `
const fs = require("fs");
const vm = require("vm");

function makeHarness(options = {}) {
  const sockets = [];
  const completions = [];
  const timers = [];
  const alerts = [];
  let reloads = 0;
  let constructorAttempts = 0;
  let constructorFailures = options.constructorFailures || 0;
  let sendFailures = options.sendFailures || 0;
  let callbackEnqueued = false;
  let harness;

  class FakeWebSocket {
    constructor(url) {
      constructorAttempts += 1;
      if (constructorFailures > 0) {
        constructorFailures -= 1;
        throw new Error("fixture constructor failure");
      }
      this.url = url;
      this.readyState = FakeWebSocket.CONNECTING;
      this.OPEN = FakeWebSocket.OPEN;
      this.sent = [];
      this.closeCount = 0;
      sockets.push(this);
    }
    send(value) {
      if (this.readyState !== FakeWebSocket.OPEN) throw new Error("send while socket is not open");
      if (sendFailures > 0) {
        sendFailures -= 1;
        throw new Error("fixture send failure");
      }
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
    alert(message) { alerts.push(message); },
    completeSourceRequest(command, data, response) {
      completions.push({command, requestId: data.requestId, status: response && response.status, err: response && response.err});
      if (!callbackEnqueued && options.enqueueAfterSuccess === command && response && response.status === true) {
        callbackEnqueued = true;
        request(harness, options.callbackCommand, {});
      }
    },
    setTimeout(callback, delay) {
      const timer = {callback, delay, cleared: false, fired: false};
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
  harness = {
    context,
    sockets,
    completions,
    timers,
    alerts,
    escapedErrors: [],
    constructorAttempts: () => constructorAttempts,
    reloads: () => reloads,
  };
  return harness;
}

function request(harness, command, data) {
  harness.context.command = command;
  harness.context.requestData = data;
  try {
    vm.runInContext("new Server(command).request(requestData)", harness.context);
  } catch (error) {
    harness.escapedErrors.push(String(error));
  }
}

function openSocket(harness, socket) {
  if (!socket) return;
  try {
    socket.open();
  } catch (error) {
    harness.escapedErrors.push(String(error));
  }
}

function respond(socket, response) {
  const messages = sent(socket);
  if (!socket || messages.length === 0) return;
  response.requestId = messages[messages.length - 1].requestId;
  socket.emitMessage(response);
}

function fireTimer(timer) {
  if (!timer || timer.cleared || timer.fired) return;
  timer.fired = true;
  timer.callback();
}

function pendingTimers(harness) {
  return harness.timers.filter(timer => !timer.cleared && !timer.fired);
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
openSocket(happyHarness, happySocket);
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
const mismatchRetryTimers = pendingTimers(mismatchHarness).filter(timer => timer.delay === 250);
const mismatchNoSynchronousRetry = mismatchHarness.sockets.length === 1;
fireTimer(mismatchRetryTimers[0]);
const mismatchSecondSocket = mismatchHarness.sockets[1];
openSocket(mismatchHarness, mismatchSecondSocket);
const mismatchReplay = sent(mismatchSecondSocket);
const mismatchSecondStillPending = completionCount(mismatchHarness, "second") === 0;
respond(mismatchSecondSocket, {status: true});

const constructorHarness = makeHarness({constructorFailures: 1});
request(constructorHarness, "first", {});
request(constructorHarness, "second", {});
const constructorQueueBeforeRetry = constructorHarness.context.THREADFIN_CONNECTION.queue.map(item => item.command);
const constructorAttemptsBeforeRetry = constructorHarness.constructorAttempts();
const constructorAlertsBeforeRetry = constructorHarness.alerts.length;
const constructorSocketsBeforeRetry = constructorHarness.sockets.length;
const constructorRetryTimers = pendingTimers(constructorHarness).filter(timer => timer.delay === 250);
fireTimer(constructorRetryTimers[0]);
const constructorSocket = constructorHarness.sockets[0];
openSocket(constructorHarness, constructorSocket);
respond(constructorSocket, {status: true});
respond(constructorSocket, {status: true});
const constructorSentAfterRetry = sent(constructorSocket).map(message => message.cmd);

const sendThrowHarness = makeHarness({sendFailures: 1});
request(sendThrowHarness, "first", {});
request(sendThrowHarness, "second", {});
openSocket(sendThrowHarness, sendThrowHarness.sockets[0]);
const sendThrowRetryTimers = pendingTimers(sendThrowHarness).filter(timer => timer.delay === 250);
const sendThrowNoSynchronousRetry = sendThrowHarness.sockets.length === 1;
fireTimer(sendThrowRetryTimers[0]);
const sendThrowSecondSocket = sendThrowHarness.sockets[1];
openSocket(sendThrowHarness, sendThrowSecondSocket);
const sendThrowReconnect = sent(sendThrowSecondSocket);
respond(sendThrowSecondSocket, {status: true});

function runPreOpenFailure(eventName) {
  const harness = makeHarness();
  request(harness, "first", {});
  request(harness, "second", {});
  const failedSocket = harness.sockets[0];
  if (eventName === "close") failedSocket.emitClose(1006);
  else failedSocket.emitError();
  const noEarlyFailures = completionCount(harness, "first") === 0 && completionCount(harness, "second") === 0;
  const retryTimers = pendingTimers(harness).filter(timer => timer.delay === 250);
  const delayedRetry = harness.sockets.length === 1;
  fireTimer(retryTimers[0]);
  const replacement = harness.sockets[1];
  openSocket(harness, replacement);
  respond(replacement, {status: true});
  respond(replacement, {status: true});
  return {
    commands: sent(replacement).map(message => message.cmd),
    completed: ["first", "second"].map(command => completionCount(harness, command)),
    noEarlyFailures,
    delayedRetry,
    singleRetryTimer: retryTimers.length === 1,
  };
}
const preOpenClose = runPreOpenFailure("close");
const preOpenError = runPreOpenFailure("error");

const transportHarness = makeHarness();
request(transportHarness, "first", {});
request(transportHarness, "second", {});
request(transportHarness, "third", {});
const transportFirstSocket = transportHarness.sockets[0];
openSocket(transportHarness, transportFirstSocket);
transportFirstSocket.emitError();
transportFirstSocket.emitClose(1006);
const transportRetryTimers = pendingTimers(transportHarness).filter(timer => timer.delay === 250);
const transportNoSynchronousRetry = transportHarness.sockets.length === 1;
fireTimer(transportRetryTimers[0]);
const transportSecondSocket = transportHarness.sockets[1];
openSocket(transportHarness, transportSecondSocket);
respond(transportSecondSocket, {status: true});
respond(transportSecondSocket, {status: true});
const transportReconnect = sent(transportSecondSocket);

const timeoutHarness = makeHarness();
request(timeoutHarness, "first", {});
request(timeoutHarness, "second", {});
request(timeoutHarness, "third", {});
const timeoutFirstSocket = timeoutHarness.sockets[0];
openSocket(timeoutHarness, timeoutFirstSocket);
const timedOutRequest = sent(timeoutFirstSocket)[0];
const activeTimer = pendingTimers(timeoutHarness).find(timer => timer.delay === 30000);
fireTimer(activeTimer);
fireTimer(pendingTimers(timeoutHarness).find(timer => timer.delay === 250));
const timeoutSecondSocket = timeoutHarness.sockets[1];
openSocket(timeoutHarness, timeoutSecondSocket);
const timeoutActiveRequest = timeoutHarness.context.THREADFIN_CONNECTION.active;
timeoutFirstSocket.emitMessage({status: true, requestId: timedOutRequest.requestId});
const lateMessageIgnored = timeoutHarness.context.THREADFIN_CONNECTION.active === timeoutActiveRequest && completionCount(timeoutHarness, "second") === 0;
timeoutFirstSocket.emitError();
const lateErrorIgnored = timeoutHarness.context.THREADFIN_CONNECTION.active === timeoutActiveRequest && completionCount(timeoutHarness, "second") === 0;
timeoutFirstSocket.emitClose(1006);
const lateCloseIgnored = timeoutHarness.context.THREADFIN_CONNECTION.active === timeoutActiveRequest && completionCount(timeoutHarness, "second") === 0;
const replacementUnchanged = timeoutHarness.context.THREADFIN_CONNECTION.socket === timeoutSecondSocket && timeoutHarness.sockets.length === 2;
respond(timeoutSecondSocket, {status: true});
respond(timeoutSecondSocket, {status: true});
const timeoutReconnect = sent(timeoutSecondSocket);

const stalePolicyHarness = makeHarness();
request(stalePolicyHarness, "first", {});
request(stalePolicyHarness, "second", {});
request(stalePolicyHarness, "third", {});
request(stalePolicyHarness, "fourth", {});
const stalePolicyOldSocket = stalePolicyHarness.sockets[0];
openSocket(stalePolicyHarness, stalePolicyOldSocket);
const stalePolicyTimer = pendingTimers(stalePolicyHarness).find(timer => timer.delay === 30000);
fireTimer(stalePolicyTimer);
fireTimer(pendingTimers(stalePolicyHarness).find(timer => timer.delay === 250));
const stalePolicyReplacement = stalePolicyHarness.sockets[1];
openSocket(stalePolicyHarness, stalePolicyReplacement);
const stalePolicyConnection = stalePolicyHarness.context.THREADFIN_CONNECTION;
const stalePolicyActiveBefore = stalePolicyConnection.active ? stalePolicyConnection.active.command : "";
const stalePolicyQueueBefore = stalePolicyConnection.queue.map(item => item.command);
const stalePolicySocketCountBefore = stalePolicyHarness.sockets.length;
stalePolicyOldSocket.emitClose(1008);
for (const timer of pendingTimers(stalePolicyHarness)) fireTimer(timer);
const stalePolicyFailureStatuses = ["first", "second", "third", "fourth"].map(command => {
  const completion = stalePolicyHarness.completions.find(item => item.command === command);
  return completion ? completion.status === false : false;
});

const callbackHarness = makeHarness({enqueueAfterSuccess: "first", callbackCommand: "callback"});
request(callbackHarness, "first", {});
request(callbackHarness, "tail", {});
const callbackSocket = callbackHarness.sockets[0];
openSocket(callbackHarness, callbackSocket);
respond(callbackSocket, {status: true});
respond(callbackSocket, {status: true});
respond(callbackSocket, {status: true});

function terminalPreOpenResult(harness, attemptsBeforeRetry, retryTimers, currentSocket) {
  const connection = harness.context.THREADFIN_CONNECTION;
  request(harness, "afterTerminal", {});
  return {
    attempts: harness.constructorAttempts(),
    attemptsBeforeRetry,
    pendingTimersBeforeRetry: retryTimers.length,
    retryDelay: retryTimers[0] ? retryTimers[0].delay : -1,
    completionCounts: ["first", "second"].map(command => completionCount(harness, command)),
    alerts: harness.alerts.length,
    reloads: harness.reloads(),
    queueLength: connection.queue.length,
    pendingTimers: pendingTimers(harness).length,
    postFailureCount: completionCount(harness, "afterTerminal"),
    attemptsAfterPost: harness.constructorAttempts(),
    socketCleared: connection.socket === null,
    activeCleared: connection.active === null,
    currentSocketClosed: !currentSocket || (currentSocket.readyState === 3 && currentSocket.closeCount === 1),
  };
}

const terminalConstructorHarness = makeHarness({constructorFailures: 2});
request(terminalConstructorHarness, "first", {});
request(terminalConstructorHarness, "second", {});
const terminalConstructorAttemptsBeforeRetry = terminalConstructorHarness.constructorAttempts();
const terminalConstructorRetryTimers = pendingTimers(terminalConstructorHarness).filter(timer => timer.delay === 250);
fireTimer(terminalConstructorRetryTimers[0]);
const terminalConstructor = terminalPreOpenResult(terminalConstructorHarness, terminalConstructorAttemptsBeforeRetry, terminalConstructorRetryTimers, null);

function runTerminalAsyncFailure(eventName) {
  const harness = makeHarness();
  request(harness, "first", {});
  request(harness, "second", {});
  const firstSocket = harness.sockets[0];
  if (eventName === "error") {
    firstSocket.emitError();
    firstSocket.emitClose(1006);
  } else {
    firstSocket.emitClose(1006);
  }
  const attemptsBeforeRetry = harness.constructorAttempts();
  const retryTimers = pendingTimers(harness).filter(timer => timer.delay === 250);
  fireTimer(retryTimers[0]);
  const secondSocket = harness.sockets[1];
  if (eventName === "error") {
    secondSocket.emitError();
    secondSocket.emitClose(1006);
  } else {
    secondSocket.emitClose(1006);
  }
  return terminalPreOpenResult(harness, attemptsBeforeRetry, retryTimers, secondSocket);
}

const terminalAsyncError = runTerminalAsyncFailure("error");
const terminalAsyncClose = runTerminalAsyncFailure("close");

const policyHarness = makeHarness();
request(policyHarness, "first", {});
request(policyHarness, "second", {});
request(policyHarness, "third", {});
const policySocket = policyHarness.sockets[0];
if (policySocket) {
  policySocket.open();
  policySocket.emitClose(1008);
}
request(policyHarness, "afterPolicy", {});

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
    noSynchronousRetry: mismatchNoSynchronousRetry,
    retryDelay: mismatchRetryTimers[0] ? mismatchRetryTimers[0].delay : -1,
    pendingRetryTimers: mismatchRetryTimers.length,
    reconnected: mismatchHarness.sockets.length === 2,
    replayCommands: mismatchReplay.map(message => message.cmd),
    secondStillPending: mismatchSecondStillPending,
  },
  constructorFailure: {
    escapedErrors: constructorHarness.escapedErrors.length,
    attemptsBeforeRetry: constructorAttemptsBeforeRetry,
    alertsBeforeRetry: constructorAlertsBeforeRetry,
    pendingRetryTimers: constructorRetryTimers.length,
    retryDelay: constructorRetryTimers[0] ? constructorRetryTimers[0].delay : -1,
    queueBeforeRetry: constructorQueueBeforeRetry,
    socketsBeforeRetry: constructorSocketsBeforeRetry,
    sentAfterRetry: constructorSentAfterRetry,
    completionCounts: ["first", "second"].map(command => completionCount(constructorHarness, command)),
  },
  sendThrow: {
    escapedErrors: sendThrowHarness.escapedErrors.length,
    failureCounts: ["first", "second"].map(command => completionCount(sendThrowHarness, command)),
    noSynchronousRetry: sendThrowNoSynchronousRetry,
    retryDelay: sendThrowRetryTimers[0] ? sendThrowRetryTimers[0].delay : -1,
    reconnectCommands: sendThrowReconnect.map(message => message.cmd),
    socketCount: sendThrowHarness.sockets.length,
  },
  preOpen: {
    closeCommands: preOpenClose.commands,
    closeCompleted: preOpenClose.completed,
    errorCommands: preOpenError.commands,
    errorCompleted: preOpenError.completed,
    noEarlyFailures: preOpenClose.noEarlyFailures && preOpenError.noEarlyFailures,
    delayedRetries: preOpenClose.delayedRetry && preOpenError.delayedRetry,
    singleRetryTimers: preOpenClose.singleRetryTimer && preOpenError.singleRetryTimer,
  },
  transport: {
    firstFailedOnce: completionCount(transportHarness, "first") === 1,
    noSynchronousRetry: transportNoSynchronousRetry,
    retryDelay: transportRetryTimers[0] ? transportRetryTimers[0].delay : -1,
    reconnected: transportHarness.sockets.length === 2,
    reconnectCommands: transportReconnect.map(message => message.cmd),
    completionCounts: ["first", "second", "third"].map(command => completionCount(transportHarness, command)),
  },
  timeout: {
    delay: activeTimer ? activeTimer.delay : -1,
    firstFailedOnce: completionCount(timeoutHarness, "first") === 1,
    reconnectCommands: timeoutReconnect.map(message => message.cmd),
    lateMessageIgnored,
    lateErrorIgnored,
    lateCloseIgnored,
    completionCounts: ["first", "second", "third"].map(command => completionCount(timeoutHarness, command)),
    replacementUnchanged,
  },
  stalePolicy: {
    activeBeforePolicy: stalePolicyActiveBefore,
    queueBeforePolicy: stalePolicyQueueBefore,
    failureCounts: ["first", "second", "third", "fourth"].map(command => completionCount(stalePolicyHarness, command)),
    failureStatuses: stalePolicyFailureStatuses,
    reloads: stalePolicyHarness.reloads(),
    socketCount: stalePolicyHarness.sockets.length,
    queueLength: stalePolicyConnection.queue.length,
    activeCleared: stalePolicyConnection.active === null,
    socketCleared: stalePolicyConnection.socket === null,
    replacementClosed: stalePolicyReplacement.readyState === 3 && stalePolicyReplacement.closeCount === 1,
    noReconnect: stalePolicyHarness.sockets.length === stalePolicySocketCountBefore,
  },
  callbackFifo: {
    commands: sent(callbackSocket).map(message => message.cmd),
    completionCounts: ["first", "tail", "callback"].map(command => completionCount(callbackHarness, command)),
  },
  policy: {
    failureCounts: ["first", "second", "third"].map(command => completionCount(policyHarness, command)),
    postFailureCount: completionCount(policyHarness, "afterPolicy"),
    reloads: policyHarness.reloads(),
    socketCount: policyHarness.sockets.length,
    queueLength: policyHarness.context.THREADFIN_CONNECTION ? policyHarness.context.THREADFIN_CONNECTION.queue.length : -1,
    activeCleared: policyHarness.context.THREADFIN_CONNECTION ? policyHarness.context.THREADFIN_CONNECTION.active === null : false,
  },
  preOpenTerminal: {
    constructor: terminalConstructor,
    asyncError: terminalAsyncError,
    asyncClose: terminalAsyncClose,
  },
}));
`
