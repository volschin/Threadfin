var mappingWorkspaceState: MappingWorkspaceState
var mappingCurrentQuery: MappingQuery = { sort: "number" }
var mappingPendingNavigation: () => void
var mappingAfterSave: () => void
var mappingProbeFeedback = ""
var mappingGuardDialog: HTMLElement
var mappingGuardReturnFocus: HTMLElement
var mappingBeforeUnloadBound = false

interface MappingFocusSnapshot {
  key: string
  selectionStart?: number
  selectionEnd?: number
}

function initializeMappingWorkspace(server: any, force: boolean = false): MappingWorkspaceState {
  var authoritative = mappingAuthoritativeMap(server)
  if (force || !mappingWorkspaceState || (mappingDirtyIDs(mappingWorkspaceState).length == 0 && mappingComparable(mappingWorkspaceState.baseline) != mappingComparable(authoritative))) {
    mappingWorkspaceState = createMappingWorkspaceState(server)
    mappingCurrentQuery.segment = mappingWorkspaceState.segment
  } else {
    mappingWorkspaceState.xmltvMap = mappingDeepClone(mappingXMLTVMap(server))
  }
  bindMappingBeforeUnload()
  return mappingWorkspaceState
}

function mappingHasDirtyDraft(): boolean {
  return !!mappingWorkspaceState && mappingDirtyIDs(mappingWorkspaceState).length > 0
}

function saveMappingDraft(afterSave?: () => void): void {
  if (!mappingWorkspaceState || !mappingHasDirtyDraft() || mappingWorkspaceState.saveState == "pending" || mappingWorkspaceState.saveState == "ambiguous") {
    return
  }
  mappingWorkspaceState.saveState = "pending"
  mappingWorkspaceState.feedback = "Saving mapping and rebuilding outputs"
  mappingAfterSave = afterSave
  updateMappingGuardButtons()
  renderMappingWorkspaceIfOpen()
  var server = new Server("saveEpgMapping")
  server.request({ epgMapping: mappingDeepClone(mappingWorkspaceState.draft) })
}

function mappingResponseAuthoritative(response: any): { present: boolean, mapping: { [id: string]: any }, xmltvMap: { [file: string]: any } } {
  var root = mappingRecord(response)
  var xepg = mappingRecord(root.xepg)
  var present = Object.prototype.hasOwnProperty.call(xepg, "epgMapping")
  return { present: present, mapping: mappingRecord(xepg.epgMapping), xmltvMap: mappingRecord(xepg.xmltvMap) }
}

function completeMappingRequest(command: string, data: any, response: any, failureKind?: "busy" | "transport"): void {
  if (command == "probeChannel") {
    completeMappingProbe(response)
    return
  }
  if (command == "getServerConfig" && mappingWorkspaceState && mappingWorkspaceState.saveState == "ambiguous") {
    var refetched = mappingResponseAuthoritative(response)
    if (mappingRecord(response).status === true && refetched.present) {
      mappingWorkspaceState.xmltvMap = mappingDeepClone(refetched.xmltvMap)
      mappingReconcileAuthoritative(mappingWorkspaceState, refetched.mapping, mappingRecord(data).epgMapping, false)
      mappingWorkspaceState.saveState = "idle"
      mappingWorkspaceState.feedback = mappingHasDirtyDraft() ? "Authoritative mapping reloaded; review the retained draft before saving again" : "Authoritative mapping reloaded"
    } else {
      mappingWorkspaceState.feedback = "Save result remains unknown; authoritative mapping could not be reloaded"
    }
    renderMappingWorkspaceIfOpen()
    return
  }
  if (command != "saveEpgMapping" || !mappingWorkspaceState) {
    return
  }
  if (failureKind == "busy") {
    mappingWorkspaceState.saveState = "idle"
    mappingWorkspaceState.feedback = "Another request is active; the mapping draft was retained"
    renderMappingWorkspaceIfOpen()
    return
  }
  if (failureKind == "transport") {
    mappingWorkspaceState.saveState = "ambiguous"
    mappingWorkspaceState.feedback = "Connection lost after save; checking the authoritative mapping before allowing another save"
    renderMappingWorkspaceIfOpen()
    new Server("getServerConfig").request(new Object())
    return
  }

  var root = mappingRecord(response)
  var authoritative = mappingResponseAuthoritative(root)
  if (!authoritative.present) {
    mappingWorkspaceState.saveState = "idle"
    mappingWorkspaceState.feedback = "The server response did not include the authoritative mapping; the draft was retained"
    renderMappingWorkspaceIfOpen()
    return
  }
  mappingWorkspaceState.xmltvMap = mappingDeepClone(authoritative.xmltvMap)
  var reconciliation = mappingReconcileAuthoritative(mappingWorkspaceState, authoritative.mapping, mappingRecord(data).epgMapping, root.status === true)
  mappingWorkspaceState.saveState = "idle"
  if (root.status === true && root.mappingSaveResult == "outputsRebuilt") {
    mappingWorkspaceState.feedback = "Mapping saved; outputs rebuilt"
  } else if (root.status === true && root.mappingSaveResult == "outputRebuildRequested") {
    mappingWorkspaceState.feedback = "Mapping saved; output rebuild requested"
  } else if (root.status === true) {
    mappingWorkspaceState.feedback = "Mapping saved; output rebuild state was not confirmed"
  } else if (reconciliation.persisted) {
    mappingWorkspaceState.feedback = "Mapping saved, but output rebuilding failed" + (mappingString(root.err) ? ": " + mappingString(root.err) : "")
  } else {
    mappingWorkspaceState.feedback = mappingString(root.err) || "Mapping was not saved; the draft was retained"
  }
  var continuation = reconciliation.persisted ? mappingAfterSave : undefined
  mappingAfterSave = undefined
  renderMappingWorkspaceIfOpen()
  if (continuation) {
    closeMappingGuard()
    continuation()
  }
}

function requestMappingProbe(url: string): void {
  mappingProbeFeedback = "Probing channel…"
  renderMappingProbeFeedback()
  new Server("probeChannel").request({ probeUrl: url })
}

function completeMappingProbe(response: any): void {
  var root = mappingRecord(response)
  if (root.status !== true) {
    mappingProbeFeedback = mappingString(root.err) || "Probe failed"
  } else {
    var info = mappingRecord(root.probeInfo)
    var values: string[] = []
    if (info.resolution) values.push("Resolution: " + mappingString(info.resolution))
    if (info.frameRate) values.push("Frame rate: " + mappingString(info.frameRate) + " FPS")
    if (info.audioChannel) values.push("Audio: " + mappingString(info.audioChannel))
    mappingProbeFeedback = values.length ? values.join(" · ") : "Details unavailable"
  }
  renderMappingProbeFeedback()
}

function renderMappingProbeFeedback(): void {
  var status = document.getElementById("mapping-probe-status")
  if (status) {
    status.textContent = mappingProbeFeedback
  }
}

function renderMappingWorkspaceIfOpen(): void {
  if (typeof document == "undefined") return
  var host = document.getElementById("content")
  if (host && (typeof currentDestination == "undefined" || currentDestination == "mapping")) {
    var focus = captureMappingWorkspaceFocus(host)
    renderMappingPage(host)
    restoreMappingWorkspaceFocus(focus)
  }
  updateMappingGuardButtons()
  updateMappingMutationControls()
}

function captureMappingWorkspaceFocus(host: HTMLElement): MappingFocusSnapshot {
  var active = document.activeElement as HTMLInputElement
  if (!active || !host.contains(active)) return undefined
  var key = active.getAttribute("data-mapping-focus")
  if (!key) return undefined
  var snapshot: MappingFocusSnapshot = { key: key }
  if (typeof active.selectionStart == "number") {
    snapshot.selectionStart = active.selectionStart
    snapshot.selectionEnd = active.selectionEnd
  }
  return snapshot
}

function restoreMappingWorkspaceFocus(snapshot: MappingFocusSnapshot): void {
  if (!snapshot) return
  var controls = document.querySelectorAll("[data-mapping-focus]")
  var target: HTMLInputElement
  Array.prototype.some.call(controls, function (control: HTMLInputElement) {
    if (control.getAttribute("data-mapping-focus") == snapshot.key) {
      target = control
      return true
    }
    return false
  })
  if (!target) return
  target.focus()
  if (snapshot.selectionStart !== undefined && typeof target.setSelectionRange == "function") {
    try {
      target.setSelectionRange(snapshot.selectionStart, snapshot.selectionEnd)
    } catch (_error) {
      // Controls such as selects do not expose a text selection.
    }
  }
}

function mappingRequestNavigation(continuation: () => void, invoker?: HTMLElement): boolean {
  if (!mappingHasDirtyDraft()) {
    return true
  }
  var candidate = arguments.length > 1 ? invoker : document.activeElement as HTMLElement
  mappingGuardReturnFocus = candidate
  mappingPendingNavigation = continuation
  openMappingGuard()
  return false
}

function openMappingGuard(): void {
  if (mappingGuardDialog && mappingGuardDialog.parentElement) {
    updateMappingGuardButtons()
    return
  }
  var guard = document.createElement("div")
  guard.className = "tf-mapping-guard"
  guard.setAttribute("role", "dialog")
  guard.setAttribute("aria-modal", "true")
  guard.setAttribute("aria-labelledby", "mapping-guard-title")
  var panel = document.createElement("div")
  panel.className = "tf-mapping-guard-panel"
  var title = document.createElement("h2")
  title.id = "mapping-guard-title"
  title.textContent = "Unsaved mapping draft"
  var copy = document.createElement("p")
  copy.textContent = "Save or discard staged Mapping changes before leaving."
  var actions = document.createElement("div")
  actions.className = "tf-mapping-guard-actions"
  var save = mappingButton("Save mapping", function () { saveMappingDraft(mappingPendingNavigation) })
  save.setAttribute("data-mapping-guard", "save")
  var discard = mappingButton("Discard draft", function () {
    if (mappingWorkspaceState && mappingWorkspaceState.saveState == "idle") {
      mappingRevertDraft(mappingWorkspaceState)
      var continuation = mappingPendingNavigation
      closeMappingGuard()
      renderMappingWorkspaceIfOpen()
      if (continuation) continuation()
    }
  })
  discard.setAttribute("data-mapping-guard", "discard")
  var stay = mappingButton("Stay", stayOnMapping)
  stay.setAttribute("data-mapping-guard", "stay")
  actions.appendChild(save); actions.appendChild(discard); actions.appendChild(stay)
  panel.appendChild(title); panel.appendChild(copy); panel.appendChild(actions); guard.appendChild(panel)
  mappingOverlayHost().appendChild(guard)
  mappingGuardDialog = guard
  bindMappingDialogKeyboard(guard, stayOnMapping)
  updateMappingGuardButtons()
  stay.focus()
}

function updateMappingGuardButtons(): void {
  if (!mappingGuardDialog || !mappingWorkspaceState) return
  var disabled = mappingWorkspaceState.saveState == "pending" || mappingWorkspaceState.saveState == "ambiguous"
  var controls = mappingGuardDialog.querySelectorAll("button")
  Array.prototype.forEach.call(controls, function (button: HTMLButtonElement) {
    button.disabled = button.getAttribute("data-mapping-guard") == "stay" ? false : disabled
  })
}

function stayOnMapping(): void {
  var returnFocus = mappingGuardReturnFocus
  mappingAfterSave = undefined
  closeMappingGuard()
  restoreMappingGuardFocus(returnFocus)
}

function closeMappingGuard(): void {
  if (mappingGuardDialog && mappingGuardDialog.parentElement) mappingGuardDialog.parentElement.removeChild(mappingGuardDialog)
  mappingGuardDialog = undefined
  mappingPendingNavigation = undefined
  mappingGuardReturnFocus = undefined
}

function mappingGuardFocusIsAvailable(target: HTMLElement): boolean {
  return !!target && typeof target.focus == "function" && document.contains(target) &&
    !(target as HTMLButtonElement).disabled && target.getAttribute("aria-disabled") != "true" && !target.hidden
}

function restoreMappingGuardFocus(returnFocus: HTMLElement): void {
  if (mappingGuardFocusIsAvailable(returnFocus)) {
    returnFocus.focus()
    return
  }
  var fallback = document.getElementById("mapping-heading") as HTMLElement
  if (!mappingGuardFocusIsAvailable(fallback)) {
    fallback = document.querySelector(".tf-mapping") as HTMLElement
  }
  if (mappingGuardFocusIsAvailable(fallback)) {
    if (fallback.getAttribute("tabindex") == null) fallback.setAttribute("tabindex", "-1")
    fallback.focus()
  }
}

function mappingDialogFocusables(dialog: HTMLElement): HTMLElement[] {
  var controls = dialog.querySelectorAll("button, input, select, summary, [tabindex]")
  return Array.prototype.filter.call(controls, function (control: HTMLElement) {
    if ((control as HTMLButtonElement).disabled || control.getAttribute("tabindex") == "-1") return false
    var current = control
    while (current && current != dialog) {
      if (current.hidden) return false
      if (current.parentElement && current.parentElement.tagName == "DETAILS" && !(current.parentElement as HTMLDetailsElement).open && current.tagName != "SUMMARY") return false
      current = current.parentElement
    }
    return true
  })
}

function bindMappingDialogKeyboard(dialog: HTMLElement, onEscape: () => void): void {
  dialog.addEventListener("keydown", function (event: KeyboardEvent) {
    if (event.key == "Escape") {
      event.preventDefault()
      event.stopPropagation()
      onEscape()
      return
    }
    if (event.key != "Tab") return
    var focusables = mappingDialogFocusables(dialog)
    if (!focusables.length) {
      event.preventDefault()
      return
    }
    var first = focusables[0]
    var last = focusables[focusables.length - 1]
    if (event.shiftKey && document.activeElement == first) {
      event.preventDefault()
      last.focus()
    } else if (!event.shiftKey && document.activeElement == last) {
      event.preventDefault()
      first.focus()
    }
  })
}

function bindMappingBeforeUnload(): void {
  if (mappingBeforeUnloadBound || typeof window == "undefined") return
  mappingBeforeUnloadBound = true
  window.addEventListener("beforeunload", function (event: BeforeUnloadEvent) {
    if (mappingHasDirtyDraft()) {
      event.preventDefault()
      event.returnValue = ""
    }
  })
}

function mappingOverlayHost(): HTMLElement {
  return document.querySelector(".tf-app") as HTMLElement || document.body
}

function renderMappingPage(host: HTMLElement): void {
  host.innerHTML = ""
  var page = document.createElement("div")
  page.className = "tf-mapping"
  var pmsMode = mappingString(mappingRecord(mappingRecord(SERVER).settings).epgSource).toUpperCase() == "PMS"
  if (!pmsMode) {
    initializeMappingWorkspace(SERVER)
  }
  page.appendChild(renderMappingHeader(!pmsMode))
  if (pmsMode) {
    var note = document.createElement("section")
    note.className = "tf-mapping-empty tf-mapping-mode-note"
    note.setAttribute("role", "note")
    var title = document.createElement("h2")
    title.textContent = "Mapping requires XEPG"
    var copy = document.createElement("p")
    copy.textContent = "PMS mode leaves guide data management to the connected client. Switch the EPG Source to XEPG to map channels and generate M3U/XMLTV guide outputs in Threadfin."
    var settings = mappingButton("Open EPG Source settings", function () { openDestination("settings", true, settings) })
    note.appendChild(title)
    note.appendChild(copy)
    note.appendChild(settings)
    page.appendChild(note)
    host.appendChild(page)
    return
  }
  page.appendChild(renderMappingFilters())
  var rows = mappingVisibleRows(mappingWorkspaceState, mappingCurrentQuery)
  page.appendChild(renderMappingCounts(rows))
  if (mappingWorkspaceState.selected.size > 0) page.appendChild(renderMappingBulkBar(rows))
  page.appendChild(renderMappingTable(rows))
  page.appendChild(renderMappingBackupOptions())
  page.appendChild(renderMappingSaveBar())
  host.appendChild(page)
}

function renderMappingHeader(showViews: boolean = true): HTMLElement {
  var header = document.createElement("header")
  header.className = "tf-mapping-header"
  var group = document.createElement("div")
  var title = document.createElement("h1")
  title.id = "mapping-heading"
  title.setAttribute("tabindex", "-1")
  title.textContent = "Mapping"
  var purpose = document.createElement("p")
  purpose.textContent = "Assign guide data, channel metadata, numbering, and backup streams before saving the complete lineup."
  group.appendChild(title); group.appendChild(purpose); header.appendChild(group)
  if (!showViews) {
    return header
  }
  var segments = document.createElement("div")
  segments.className = "tf-mapping-segments"
  segments.setAttribute("role", "group")
  segments.setAttribute("aria-label", "Mapping view")
  ;[["attention", "Needs attention"], ["active", "Active"], ["inactive", "Inactive"]].forEach(function (entry) {
    var segment = entry[0] as MappingSegment
    var button = mappingButton(entry[1], function () {
      mappingWorkspaceState.segment = segment
      mappingCurrentQuery.segment = segment
      renderMappingWorkspaceIfOpen()
    })
    button.setAttribute("data-mapping-focus", "segment-" + segment)
    button.setAttribute("aria-pressed", String((mappingCurrentQuery.segment || mappingWorkspaceState.segment) == segment))
    segments.appendChild(button)
  })
  header.appendChild(segments)
  return header
}

function renderMappingFilters(): HTMLElement {
  var region = document.createElement("section")
  region.className = "tf-mapping-filters"
  region.setAttribute("aria-label", "Mapping search and filters")
  region.appendChild(mappingFilterInput("Search mapping", mappingString(mappingCurrentQuery.search), function (value) { mappingCurrentQuery.search = value }))
  region.appendChild(mappingFilterSelect("Playlist", "playlist", mappingDistinctValues("_file.m3u.id"), mappingString(mappingCurrentQuery.playlist)))
  region.appendChild(mappingFilterSelect("Group", "group", mappingDistinctGroups(), mappingString(mappingCurrentQuery.group)))
  region.appendChild(mappingFilterSelect("XMLTV source", "xmltv", mappingDistinctValues("x-xmltv-file"), mappingString(mappingCurrentQuery.xmltv)))
  region.appendChild(mappingFilterSelect("Activation", "activation", [{ value: "active", label: "Active" }, { value: "inactive", label: "Inactive" }], mappingString(mappingCurrentQuery.activation)))
  region.appendChild(mappingFilterSelect("Attention reason", "reason", [
    { value: "missing", label: "Missing EPG assignment" }, { value: "invalid", label: "Invalid EPG assignment" },
    { value: "hidden", label: "Hidden from outputs" }, { value: "inactive", label: "Inactive" },
  ], mappingString(mappingCurrentQuery.reason)))
  return region
}

function mappingFilterInput(label: string, value: string, update: (value: string) => void): HTMLElement {
  var wrapper = document.createElement("label")
  wrapper.textContent = label
  var input = document.createElement("input")
  input.type = "search"; input.value = value; input.setAttribute("aria-label", label)
  input.setAttribute("data-mapping-focus", "filter-search")
  input.addEventListener("input", function () { update(input.value); renderMappingWorkspaceIfOpen() })
  wrapper.appendChild(input)
  return wrapper
}

function mappingFilterSelect(label: string, key: string, options: { value: string, label: string }[], selected: string): HTMLElement {
  var wrapper = document.createElement("label")
  wrapper.textContent = label
  var select = document.createElement("select")
  select.setAttribute("aria-label", label)
  select.setAttribute("data-mapping-focus", "filter-" + key)
  var all = document.createElement("option"); all.value = ""; all.textContent = "All"; select.appendChild(all)
  options.forEach(function (entry) { var option = document.createElement("option"); option.value = entry.value; option.textContent = entry.label; select.appendChild(option) })
  select.value = selected
  select.addEventListener("change", function () { mappingCurrentQuery[key] = select.value; renderMappingWorkspaceIfOpen() })
  wrapper.appendChild(select)
  return wrapper
}

function mappingDistinctValues(key: string): { value: string, label: string }[] {
  var values = new Set<string>()
  Object.keys(mappingWorkspaceState.draft).forEach(function (id) { var value = mappingString(mappingWorkspaceState.draft[id][key]); if (value) values.add(value) })
  return Array.from(values).sort().map(function (value) { return { value: value, label: mappingProviderLabel(key, value) } })
}

function mappingDistinctGroups(): { value: string, label: string }[] {
  var values = new Set<string>()
  Object.keys(mappingWorkspaceState.draft).forEach(function (id) { var channel = mappingWorkspaceState.draft[id]; var value = mappingString(channel["x-group-title"] || channel["group-title"]); if (value) values.add(value) })
  return Array.from(values).sort().map(function (value) { return { value: value, label: value } })
}

function mappingProviderLabel(key: string, value: string): string {
  if (value == "-" || value == "Threadfin Dummy") return value
  var files = mappingRecord(mappingRecord(mappingRecord(SERVER).settings).files)
  var type = key == "_file.m3u.id" ? "m3u" : "xmltv"
  var fileID = type == "xmltv" && value.lastIndexOf(".") > 0 ? value.substring(0, value.lastIndexOf(".")) : value
  var provider = mappingRecord(mappingRecord(files[type])[fileID])
  return mappingString(provider.name) || value
}

function renderMappingCounts(rows: MappingRow[]): HTMLElement {
  var counts = document.createElement("p")
  counts.className = "tf-mapping-counts"
  counts.setAttribute("role", "status")
  counts.setAttribute("aria-live", "polite")
  counts.textContent = rows.length + " results · " + mappingWorkspaceState.selected.size + " selected (" + mappingSelectedVisibleCount(mappingWorkspaceState, rows.map(function (row) { return row.id })) + " in results)"
  return counts
}

function renderMappingBulkBar(rows: MappingRow[]): HTMLElement {
  var bar = document.createElement("div")
  bar.className = "tf-mapping-bulk-bar"
  var summary = document.createElement("strong")
  var visibleSelected = mappingSelectedVisibleIDs(mappingWorkspaceState, rows)
  summary.textContent = mappingWorkspaceState.selected.size + " selected (" + visibleSelected.length + " in results)"
  bar.appendChild(summary)
  var edit = mappingButton("Edit selected", function () { openMappingEditor(visibleSelected, bar) })
  edit.setAttribute("data-mapping-focus", "bulk-edit")
  edit.setAttribute("data-mapping-mutation", "true")
  edit.setAttribute("data-mapping-idle-disabled", String(visibleSelected.length == 0))
  edit.disabled = visibleSelected.length == 0 || mappingDraftIsLocked(mappingWorkspaceState)
  bar.appendChild(edit)
  var dummy = document.createElement("select")
  dummy.setAttribute("aria-label", "Dummy guide duration")
  dummy.setAttribute("data-mapping-focus", "bulk-dummy")
  mappingDummyPrograms.forEach(function (value) { var option = document.createElement("option"); option.value = value; option.textContent = value.replace("_", " "); dummy.appendChild(option) })
  bar.appendChild(dummy)
  var applyDummy = mappingButton("Apply dummy guide", function () { mappingAssignDummy(mappingWorkspaceState, visibleSelected, dummy.value); renderMappingWorkspaceIfOpen() })
  applyDummy.setAttribute("data-mapping-focus", "bulk-apply-dummy")
  applyDummy.setAttribute("data-mapping-mutation", "true")
  applyDummy.setAttribute("data-mapping-idle-disabled", String(visibleSelected.length == 0))
  applyDummy.disabled = visibleSelected.length == 0 || mappingDraftIsLocked(mappingWorkspaceState)
  bar.appendChild(applyDummy)
  var clear = mappingButton("Clear selection", function () { mappingWorkspaceState.selected.clear(); renderMappingWorkspaceIfOpen() })
  clear.setAttribute("data-mapping-focus", "bulk-clear")
  bar.appendChild(clear)
  return bar
}

function renderMappingTable(rows: MappingRow[]): HTMLElement {
  if (rows.length == 0) {
    var empty = document.createElement("section")
    empty.className = "tf-mapping-empty"
    var title = document.createElement("h2"); title.textContent = "No channels in this view"
    var copy = document.createElement("p"); copy.textContent = "Change the search, filters, or Mapping view."
    empty.appendChild(title); empty.appendChild(copy); return empty
  }
  var scroll = document.createElement("div")
  scroll.className = "tf-mapping-table-scroll"
  var table = document.createElement("table")
  table.className = "tf-mapping-table"
  var caption = document.createElement("caption"); caption.textContent = "Mapping channels"; table.appendChild(caption)
  var head = document.createElement("thead"), header = document.createElement("tr")
  var selection = document.createElement("th")
  var all = document.createElement("input"); all.type = "checkbox"; all.setAttribute("aria-label", "Select all visible channels")
  all.setAttribute("data-mapping-focus", "select-all-visible")
  all.checked = rows.length > 0 && mappingSelectedVisibleCount(mappingWorkspaceState, rows.map(function (row) { return row.id })) == rows.length
  all.addEventListener("change", function () { mappingSelectVisible(mappingWorkspaceState, rows.map(function (row) { return row.id }), all.checked); renderMappingWorkspaceIfOpen() })
  selection.appendChild(all); header.appendChild(selection)
  ;[["number", "Channel"], ["name", "Name"], ["playlist", "Playlist"], ["group", "Group"], ["xmltv", "XMLTV / guide"], ["", "State"], ["", "Actions"]].forEach(function (entry) {
    var th = document.createElement("th"); th.scope = "col"
    if (entry[0]) {
      var sort = entry[0]
      var button = mappingButton(entry[1], function () {
        if (mappingCurrentQuery.sort == sort) mappingCurrentQuery.descending = !mappingCurrentQuery.descending
        else { mappingCurrentQuery.sort = sort as any; mappingCurrentQuery.descending = false }
        renderMappingWorkspaceIfOpen()
      })
      button.setAttribute("data-mapping-focus", "sort-" + sort)
      if ((mappingCurrentQuery.sort || "number") == sort) th.setAttribute("aria-sort", mappingCurrentQuery.descending ? "descending" : "ascending")
      th.appendChild(button)
    }
    else th.textContent = entry[1]
    header.appendChild(th)
  })
  head.appendChild(header); table.appendChild(head)
  var body = document.createElement("tbody")
  rows.forEach(function (row) { body.appendChild(renderMappingRow(row, rows)) })
  table.appendChild(body); scroll.appendChild(table); return scroll
}

function renderMappingRow(row: MappingRow, visibleRows: MappingRow[]): HTMLElement {
  var tr = document.createElement("tr")
  tr.setAttribute("data-mapping-id", row.id)
  var selected = document.createElement("td"), checkbox = document.createElement("input")
  checkbox.type = "checkbox"; checkbox.checked = mappingWorkspaceState.selected.has(row.id); checkbox.setAttribute("aria-label", "Select " + (mappingString(row.channel["x-name"]) || row.id))
  checkbox.setAttribute("data-mapping-focus", "row-" + row.id)
  checkbox.addEventListener("click", function (event: MouseEvent) { mappingSelectRange(mappingWorkspaceState, visibleRows.map(function (item) { return item.id }), row.id, checkbox.checked, event.shiftKey); renderMappingWorkspaceIfOpen() })
  selected.appendChild(checkbox); tr.appendChild(selected)
  mappingAppendCell(tr, mappingString(row.channel["x-channelID"]))
  mappingAppendCell(tr, mappingString(row.channel["x-name"] || row.channel["tvg-name"] || row.id))
  mappingAppendCell(tr, mappingProviderLabel("_file.m3u.id", mappingString(row.channel["_file.m3u.id"])))
  mappingAppendCell(tr, mappingString(row.channel["x-group-title"] || row.channel["group-title"]))
  mappingAppendCell(tr, mappingProviderLabel("x-xmltv-file", mappingString(row.channel["x-xmltv-file"])) + " · " + mappingString(row.channel["x-mapping"]))
  var state = document.createElement("td")
  state.className = "tf-mapping-state"
  state.textContent = row.reasons.length ? row.reasons.map(mappingAttentionReasonLabel).join(" · ") : "Active"
  tr.appendChild(state)
  var action = document.createElement("td")
  var edit = mappingButton("Edit", function () { openMappingEditor([row.id], edit) })
  edit.setAttribute("data-mapping-edit", row.id); edit.setAttribute("data-mapping-focus", "edit-" + row.id); edit.setAttribute("data-mapping-mutation", "true"); edit.setAttribute("data-mapping-idle-disabled", "false"); edit.disabled = mappingDraftIsLocked(mappingWorkspaceState); action.appendChild(edit); tr.appendChild(action)
  return tr
}

function mappingAppendCell(row: HTMLElement, value: string): void {
  var cell = document.createElement("td"); cell.textContent = value; row.appendChild(cell)
}

function renderMappingBackupOptions(): HTMLElement {
  var datalist = document.createElement("datalist")
  datalist.id = "mapping-backup-options"
  var none = document.createElement("option"); none.value = "-"; datalist.appendChild(none)
  var names = new Set<string>()
  Object.keys(mappingWorkspaceState.draft).forEach(function (id) { var name = mappingString(mappingWorkspaceState.draft[id]["tvg-name"]); if (name) names.add(name) })
  Array.from(names).sort().forEach(function (name) { var option = document.createElement("option"); option.value = name; datalist.appendChild(option) })
  return datalist
}

function renderMappingSaveBar(): HTMLElement {
  var bar = document.createElement("div")
  bar.className = "tf-mapping-save-bar"
  var dirty = mappingDirtyIDs(mappingWorkspaceState)
  bar.hidden = dirty.length == 0 && !mappingWorkspaceState.feedback
  var status = document.createElement("p")
  status.setAttribute("role", "status"); status.setAttribute("aria-live", "polite")
  status.textContent = mappingWorkspaceState.feedback || dirty.length + " staged channels"
  bar.appendChild(status)
  var save = mappingButton("Save mapping", function () { saveMappingDraft() })
  save.id = "mapping-save"; save.setAttribute("data-mapping-focus", "save"); save.disabled = dirty.length == 0 || mappingWorkspaceState.saveState != "idle"
  var discard = mappingButton("Discard draft", function () { if (mappingWorkspaceState.saveState == "idle") { mappingRevertDraft(mappingWorkspaceState); renderMappingWorkspaceIfOpen() } })
  discard.setAttribute("data-mapping-focus", "discard")
  discard.setAttribute("data-mapping-mutation", "true")
  discard.setAttribute("data-mapping-idle-disabled", String(dirty.length == 0))
  discard.disabled = dirty.length == 0 || mappingWorkspaceState.saveState != "idle"
  bar.appendChild(save); bar.appendChild(discard)
  return bar
}

function mappingButton(label: string, listener: () => void): HTMLButtonElement {
  var button = document.createElement("button")
  button.type = "button"; button.textContent = label; button.addEventListener("click", listener)
  return button
}

function updateMappingMutationControls(): void {
  if (!mappingWorkspaceState || typeof document.querySelectorAll != "function") return
  var controls = document.querySelectorAll("[data-mapping-mutation]")
  var locked = mappingDraftIsLocked(mappingWorkspaceState)
  Array.prototype.forEach.call(controls, function (control: HTMLButtonElement) { control.disabled = locked || control.getAttribute("data-mapping-idle-disabled") == "true" })
}

function openMappingEditor(ids: string[], invoker?: HTMLElement): void {
  if (!ids.length || mappingDraftIsLocked(mappingWorkspaceState)) return
  var existing = document.getElementById("mapping-editor")
  if (existing && existing.parentElement) existing.parentElement.removeChild(existing)
  var first = mappingRecord(mappingWorkspaceState.draft[ids[0]])
  var overlay = document.createElement("div")
  overlay.id = "mapping-editor"; overlay.className = "tf-mapping-editor"; overlay.setAttribute("role", "dialog"); overlay.setAttribute("aria-modal", "true"); overlay.setAttribute("aria-labelledby", "mapping-editor-title")
  var panel = document.createElement("div"); panel.className = "tf-mapping-editor-panel"
  var title = document.createElement("h2"); title.id = "mapping-editor-title"; title.textContent = ids.length == 1 ? "Edit " + (mappingString(first["x-name"]) || ids[0]) : "Edit " + ids.length + " channels"; panel.appendChild(title)
  var fieldSearch = document.createElement("input"); fieldSearch.type = "search"; fieldSearch.setAttribute("aria-label", "Search editor fields"); fieldSearch.placeholder = "Search fields"; panel.appendChild(fieldSearch)
  var form = document.createElement("form"); form.className = "tf-mapping-editor-form"
  form.addEventListener("submit", function (event) { event.preventDefault(); applyMappingEditor(ids, form, invoker) })
  var basic = document.createElement("div"); basic.className = "tf-mapping-editor-fields"
  basic.appendChild(mappingEditorCheckbox("Active", "x-active", first["x-active"] === true))
  basic.appendChild(mappingEditorInput(ids.length == 1 ? "Channel number" : "Starting channel number", ids.length == 1 ? "x-channelID" : "x-channels-start", ids.length == 1 ? mappingString(first["x-channelID"]) : ""))
  basic.appendChild(mappingEditorInput("Name", "x-name", mappingString(first["x-name"])))
  basic.appendChild(mappingEditorSelect("XMLTV source", "x-xmltv-file", ["-"].concat(Object.keys(mappingWorkspaceState.xmltvMap)), mappingString(first["x-xmltv-file"])))
  var mappingInput = mappingEditorInput("XMLTV channel", "x-mapping", mappingString(first["x-mapping"])); var mappingControl = mappingInput.querySelector("input") as HTMLInputElement; mappingControl.setAttribute("list", "mapping-program-options"); basic.appendChild(mappingInput)
  form.appendChild(basic)
  var advanced = document.createElement("details"); advanced.className = "tf-mapping-editor-advanced"
  var summary = document.createElement("summary"); summary.textContent = "Advanced fields"; advanced.appendChild(summary)
  var ppvField = mappingEditorInput("PPV extra", "x-ppv-extra", mappingString(first["x-ppv-extra"]))
  ppvField.hidden = mappingString(first["x-mapping"]) != "PPV"
  mappingControl.addEventListener("input", function () { ppvField.hidden = mappingControl.value != "PPV" })
  ;[
    mappingEditorInput("Description", "x-description", mappingString(first["x-description"])), mappingEditorCheckbox("Update channel name", "x-update-channel-name", first["x-update-channel-name"] === true),
    mappingEditorInput("Logo", "tvg-logo", mappingString(first["tvg-logo"])), mappingEditorCheckbox("Update channel logo", "x-update-channel-icon", first["x-update-channel-icon"] === true),
    mappingEditorInput("Category", "x-category", mappingString(first["x-category"])), mappingEditorInput("Group", "x-group-title", mappingString(first["x-group-title"])),
    ppvField, mappingEditorInput("Backup channel 1", "x-backup-channel-1", mappingString(first["x-backup-channel-1"]), "mapping-backup-options"),
    mappingEditorInput("Backup channel 2", "x-backup-channel-2", mappingString(first["x-backup-channel-2"]), "mapping-backup-options"), mappingEditorInput("Backup channel 3", "x-backup-channel-3", mappingString(first["x-backup-channel-3"]), "mapping-backup-options"),
    mappingEditorCheckbox("Hidden from outputs", "x-hide-channel", first["x-hide-channel"] === true),
  ].forEach(function (field) { advanced.appendChild(field) })
  form.appendChild(advanced)
  var programOptions = document.createElement("datalist"); programOptions.id = "mapping-program-options"
  var programs = new Set<string>(["-"])
  Object.keys(mappingWorkspaceState.xmltvMap).forEach(function (file) { Object.keys(mappingRecord(mappingWorkspaceState.xmltvMap[file])).forEach(function (program) { programs.add(program) }) })
  mappingDummyPrograms.forEach(function (program) { programs.add(program) })
  programs.forEach(function (program) { var option = document.createElement("option"); option.value = program; programOptions.appendChild(option) }); form.appendChild(programOptions)
  var probe = mappingButton("Probe channel", function () { requestMappingProbe(mappingString(first.url)) }); probe.type = "button"; form.appendChild(probe)
  var probeStatus = document.createElement("p"); probeStatus.id = "mapping-probe-status"; probeStatus.setAttribute("role", "status"); probeStatus.setAttribute("aria-live", "polite"); probeStatus.textContent = mappingProbeFeedback; form.appendChild(probeStatus)
  var formStatus = document.createElement("p"); formStatus.id = "mapping-editor-status"; formStatus.setAttribute("role", "alert"); form.appendChild(formStatus)
  var actions = document.createElement("div"); actions.className = "tf-mapping-editor-actions"
  actions.appendChild(mappingButton("Cancel", function () { closeMappingEditor(invoker) }))
  var apply = mappingButton("Apply to draft", function () { applyMappingEditor(ids, form, invoker) }); apply.id = "mapping-apply"; actions.appendChild(apply); form.appendChild(actions)
  apply.setAttribute("data-mapping-mutation", "true")
  apply.setAttribute("data-mapping-idle-disabled", "false")
  apply.disabled = mappingDraftIsLocked(mappingWorkspaceState)
  fieldSearch.addEventListener("input", function () { filterMappingEditorFields(form, advanced, fieldSearch.value) })
  panel.appendChild(form); overlay.appendChild(panel); mappingOverlayHost().appendChild(overlay)
  bindMappingDialogKeyboard(overlay, function () { closeMappingEditor(invoker) })
  var firstField = form.querySelector("input, select") as HTMLElement; if (firstField) firstField.focus()
}

function mappingEditorInput(label: string, name: string, value: string, list?: string): HTMLElement {
  var wrapper = document.createElement("label"); wrapper.className = "tf-mapping-editor-field"; wrapper.setAttribute("data-label", label.toLocaleLowerCase()); wrapper.textContent = label
  var input = document.createElement("input"); input.type = "text"; input.name = name; input.value = value; if (list) input.setAttribute("list", list)
  input.addEventListener("input", function () { input.setAttribute("data-changed", "true") }); wrapper.appendChild(input); return wrapper
}

function mappingEditorCheckbox(label: string, name: string, checked: boolean): HTMLElement {
  var wrapper = document.createElement("label"); wrapper.className = "tf-mapping-editor-field tf-mapping-editor-checkbox"; wrapper.setAttribute("data-label", label.toLocaleLowerCase())
  var input = document.createElement("input"); input.type = "checkbox"; input.name = name; input.checked = checked; input.addEventListener("change", function () { input.setAttribute("data-changed", "true") })
  wrapper.appendChild(input); wrapper.appendChild(document.createTextNode(label)); return wrapper
}

function mappingEditorSelect(label: string, name: string, values: string[], selected: string): HTMLElement {
  var wrapper = document.createElement("label"); wrapper.className = "tf-mapping-editor-field"; wrapper.setAttribute("data-label", label.toLocaleLowerCase()); wrapper.textContent = label
  var select = document.createElement("select"); select.name = name
  values.forEach(function (value) { var option = document.createElement("option"); option.value = value; option.textContent = mappingProviderLabel("x-xmltv-file", value); select.appendChild(option) }); select.value = selected
  select.addEventListener("change", function () { select.setAttribute("data-changed", "true") }); wrapper.appendChild(select); return wrapper
}

function filterMappingEditorFields(form: HTMLElement, advanced: HTMLDetailsElement, value: string): void {
  var search = value.trim().toLocaleLowerCase(); var fields = form.querySelectorAll(".tf-mapping-editor-field"); var advancedMatch = false
  Array.prototype.forEach.call(fields, function (field: HTMLElement) { var match = !search || mappingString(field.getAttribute("data-label")).indexOf(search) != -1; field.hidden = !match; if (match && advanced.contains(field)) advancedMatch = true })
  if (search && advancedMatch) advanced.open = true
}

function applyMappingEditor(ids: string[], form: HTMLElement, invoker?: HTMLElement): void {
  var patch: { [key: string]: any } = {}; var fields = form.querySelectorAll("input[name], select[name]")
  Array.prototype.forEach.call(fields, function (field: HTMLInputElement | HTMLSelectElement) {
    if (ids.length > 1 && field.getAttribute("data-changed") != "true") return
    if (field.name == "x-channels-start") return
    patch[field.name] = field instanceof HTMLInputElement && field.type == "checkbox" ? field.checked : field.value
  })
  var number = form.querySelector('[name="x-channelID"]') as HTMLInputElement
  var start = form.querySelector('[name="x-channels-start"]') as HTMLInputElement
  var result = mappingApplyChannelPatch(mappingWorkspaceState, ids, patch, { numberChanged: !!number && number.getAttribute("data-changed") == "true", sequentialStart: start && start.getAttribute("data-changed") == "true" ? start.value : undefined })
  if (!result.ok) { var status = document.getElementById("mapping-editor-status"); if (status) status.textContent = result.error; if (number) number.focus(); else if (start) start.focus(); return }
  closeMappingEditor(); renderMappingWorkspaceIfOpen()
  var save = document.getElementById("mapping-save") as HTMLElement; if (save) save.focus()
}

function closeMappingEditor(invoker?: HTMLElement): void {
  var editor = document.getElementById("mapping-editor"); if (editor && editor.parentElement) editor.parentElement.removeChild(editor)
  if (invoker && document.contains(invoker)) invoker.focus()
}
