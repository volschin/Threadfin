type MappingSegment = "attention" | "active" | "inactive"
type MappingAttentionReason = "missing" | "invalid" | "hidden" | "inactive"
type MappingSaveState = "idle" | "pending" | "ambiguous"

interface MappingWorkspaceState {
  baseline: { [id: string]: any }
  draft: { [id: string]: any }
  xmltvMap: { [file: string]: any }
  dirtyIDs: Set<string>
  selected: Set<string>
  segment: MappingSegment
  selectionAnchor: string
  saveState: MappingSaveState
  feedback: string
}

interface MappingQuery {
  segment?: MappingSegment
  search?: string
  playlist?: string
  group?: string
  xmltv?: string
  activation?: "active" | "inactive" | ""
  reason?: MappingAttentionReason | ""
  sort?: "number" | "name" | "playlist" | "group" | "xmltv"
  descending?: boolean
}

interface MappingRow {
  id: string
  channel: any
  reasons: MappingAttentionReason[]
  originalIndex: number
}

interface MappingPatchOptions {
  numberChanged?: boolean
  sequentialStart?: string
}

interface MappingPatchResult {
  ok: boolean
  error?: string
}

var mappingDummyPrograms: string[] = [
  "PPV", "30_Minutes", "60_Minutes", "90_Minutes", "120_Minutes",
  "180_Minutes", "240_Minutes", "360_Minutes",
]

function mappingRecord(value: any): { [key: string]: any } {
  return value && typeof value == "object" && !Array.isArray(value) ? value : {}
}

function mappingString(value: any): string {
  return value === undefined || value === null ? "" : String(value)
}

function mappingDeepClone<T>(value: T): T {
  if (value === undefined || value === null) {
    return value
  }
  return JSON.parse(JSON.stringify(value)) as T
}

function mappingComparable(value: any): string {
  if (Array.isArray(value)) {
    return "[" + value.map(mappingComparable).join(",") + "]"
  }
  if (value && typeof value == "object") {
    return "{" + Object.keys(value).sort().map(function (key) {
      return JSON.stringify(key) + ":" + mappingComparable(value[key])
    }).join(",") + "}"
  }
  return JSON.stringify(value)
}

function mappingAuthoritativeMap(server: any): { [id: string]: any } {
  return mappingRecord(mappingRecord(mappingRecord(server).xepg).epgMapping)
}

function mappingXMLTVMap(server: any): { [file: string]: any } {
  return mappingRecord(mappingRecord(mappingRecord(server).xepg).xmltvMap)
}

function createMappingWorkspaceState(server: any): MappingWorkspaceState {
  var baseline = mappingDeepClone(mappingAuthoritativeMap(server))
  var xmltvMap = mappingDeepClone(mappingXMLTVMap(server))
  var draft = mappingDeepClone(baseline)
  var hasAttention = Object.keys(draft).some(function (id) {
    return mappingAttentionReasons(draft[id], xmltvMap).length > 0
  })
  return {
    baseline: baseline,
    draft: draft,
    xmltvMap: xmltvMap,
    dirtyIDs: new Set<string>(),
    selected: new Set<string>(),
    segment: hasAttention ? "attention" : "active",
    selectionAnchor: "",
    saveState: "idle",
    feedback: "",
  }
}

function mappingAttentionReasons(channelValue: any, xmltvValue: any): MappingAttentionReason[] {
  var channel = mappingRecord(channelValue)
  var xmltvMap = mappingRecord(xmltvValue)
  var file = mappingString(channel["x-xmltv-file"])
  var program = mappingString(channel["x-mapping"])
  var reasons: MappingAttentionReason[] = []
  var missing = !file || file == "-" || !program || program == "-"
  if (missing) {
    reasons.push("missing")
  } else if (file == "Threadfin Dummy") {
    if (mappingDummyPrograms.indexOf(program) == -1) {
      reasons.push("invalid")
    }
  } else {
    var guide = mappingRecord(xmltvMap[file])
    if (!Object.prototype.hasOwnProperty.call(xmltvMap, file) || !Object.prototype.hasOwnProperty.call(guide, program)) {
      reasons.push("invalid")
    }
  }
  if (channel["x-hide-channel"] === true) {
    reasons.push("hidden")
  }
  if (channel["x-active"] !== true) {
    reasons.push("inactive")
  }
  return reasons
}

function mappingSearchText(channel: any, reasons: MappingAttentionReason[]): string {
  var values: string[] = []
  function collect(value: any): void {
    if (value === undefined || value === null) {
      return
    }
    if (Array.isArray(value)) {
      value.forEach(collect)
    } else if (typeof value == "object") {
      Object.keys(value).forEach(function (key) { collect(value[key]) })
    } else {
      values.push(String(value))
    }
  }
  collect(channel)
  reasons.forEach(function (reason) { values.push(mappingAttentionReasonLabel(reason)) })
  return values.join(" ").toLocaleLowerCase()
}

function mappingAttentionReasonLabel(reason: MappingAttentionReason): string {
  switch (reason) {
    case "missing": return "Missing EPG assignment"
    case "invalid": return "Invalid EPG assignment"
    case "hidden": return "Hidden from outputs"
    case "inactive": return "Inactive"
  }
}

function mappingVisibleRows(state: MappingWorkspaceState, queryValue?: MappingQuery): MappingRow[] {
  var query = queryValue || {}
  var segment = query.segment || state.segment
  var search = mappingString(query.search).trim().toLocaleLowerCase()
  var rows: MappingRow[] = []
  Object.keys(state.draft).forEach(function (id, originalIndex) {
    var channel = mappingRecord(state.draft[id])
    var reasons = mappingAttentionReasons(channel, state.xmltvMap)
    if (segment == "attention" && reasons.length == 0) {
      return
    }
    if (segment == "active" && channel["x-active"] !== true) {
      return
    }
    if (segment == "inactive" && channel["x-active"] === true) {
      return
    }
    if (query.playlist && mappingString(channel["_file.m3u.id"]) != query.playlist) {
      return
    }
    if (query.group && mappingString(channel["x-group-title"] || channel["group-title"]) != query.group) {
      return
    }
    if (query.xmltv && mappingString(channel["x-xmltv-file"]) != query.xmltv) {
      return
    }
    if (query.activation == "active" && channel["x-active"] !== true) {
      return
    }
    if (query.activation == "inactive" && channel["x-active"] === true) {
      return
    }
    if (query.reason && reasons.indexOf(query.reason) == -1) {
      return
    }
    if (search && mappingSearchText(channel, reasons).indexOf(search) == -1) {
      return
    }
    rows.push({ id: id, channel: channel, reasons: reasons, originalIndex: originalIndex })
  })
  return mappingSortRows(rows, query.sort || "number", query.descending === true)
}

function mappingSortRows(rows: MappingRow[], sort: string, descending: boolean): MappingRow[] {
  var sorted = rows.slice()
  var direction = descending ? -1 : 1
  sorted.sort(function (left, right) {
    var comparison = 0
    if (sort == "number") {
      var leftNumber = Number.parseFloat(mappingString(left.channel["x-channelID"]))
      var rightNumber = Number.parseFloat(mappingString(right.channel["x-channelID"]))
      var leftValid = Number.isFinite(leftNumber)
      var rightValid = Number.isFinite(rightNumber)
      comparison = leftValid && rightValid ? leftNumber - rightNumber : leftValid ? -1 : rightValid ? 1 : mappingString(left.channel["x-channelID"]).localeCompare(mappingString(right.channel["x-channelID"]))
    } else {
      var key = sort == "name" ? "x-name" : sort == "playlist" ? "_file.m3u.id" : sort == "group" ? "x-group-title" : "x-xmltv-file"
      comparison = mappingString(left.channel[key]).localeCompare(mappingString(right.channel[key]), undefined, { numeric: true, sensitivity: "base" })
    }
    return comparison == 0 ? left.originalIndex - right.originalIndex : comparison * direction
  })
  return sorted
}

function mappingSetSelected(state: MappingWorkspaceState, id: string, selected: boolean): void {
  if (!Object.prototype.hasOwnProperty.call(state.draft, id)) {
    return
  }
  if (selected) {
    state.selected.add(id)
  } else {
    state.selected.delete(id)
  }
  state.selectionAnchor = id
}

function mappingSelectRange(state: MappingWorkspaceState, visibleIDs: string[], id: string, selected: boolean, shift: boolean): void {
  var target = visibleIDs.indexOf(id)
  var anchor = visibleIDs.indexOf(state.selectionAnchor)
  if (shift && target >= 0 && anchor >= 0) {
    var start = Math.min(target, anchor)
    var end = Math.max(target, anchor)
    for (var index = start; index <= end; index++) {
      if (selected) state.selected.add(visibleIDs[index])
      else state.selected.delete(visibleIDs[index])
    }
    state.selectionAnchor = id
    return
  }
  mappingSetSelected(state, id, selected)
}

function mappingSelectVisible(state: MappingWorkspaceState, visibleIDs: string[], selected: boolean): void {
  visibleIDs.forEach(function (id) {
    if (selected) state.selected.add(id)
    else state.selected.delete(id)
  })
}

function mappingSelectedVisibleCount(state: MappingWorkspaceState, visibleIDs: string[]): number {
  return visibleIDs.reduce(function (count, id) { return count + (state.selected.has(id) ? 1 : 0) }, 0)
}

function mappingSelectedVisibleIDs(state: MappingWorkspaceState, visibleRows: MappingRow[]): string[] {
  return visibleRows.filter(function (row) { return state.selected.has(row.id) }).map(function (row) { return row.id })
}

function mappingDraftIsLocked(state: MappingWorkspaceState): boolean {
  return state.saveState != "idle"
}

function mappingDirtyIDs(state: MappingWorkspaceState): string[] {
  var ids = new Set<string>(Object.keys(state.baseline).concat(Object.keys(state.draft)))
  var dirty: string[] = []
  ids.forEach(function (id) {
    if (mappingComparable(state.baseline[id]) != mappingComparable(state.draft[id])) {
      dirty.push(id)
    }
  })
  state.dirtyIDs = new Set<string>(dirty)
  return dirty
}

function mappingResolveChannelNumber(state: MappingWorkspaceState, value: any): string {
  var number = Number.parseFloat(mappingString(value))
  if (!Number.isFinite(number)) {
    return undefined
  }
  var used = Object.keys(state.draft).map(function (id) {
    return Number.parseFloat(mappingString(mappingRecord(state.draft[id])["x-channelID"]))
  }).filter(Number.isFinite)
  while (used.indexOf(number) != -1) {
    number = Math.floor(number) == number ? number + 1 : Math.round((number + 0.1) * 10) / 10
  }
  return String(number)
}

function mappingApplyChannelPatch(state: MappingWorkspaceState, ids: string[], patchValue: any, optionsValue?: MappingPatchOptions): MappingPatchResult {
  if (mappingDraftIsLocked(state)) {
    return { ok: false, error: "Wait for the current save result before editing the draft" }
  }
  var patch = mappingRecord(patchValue)
  var options = optionsValue || {}
  var sequential: number
  if (options.sequentialStart !== undefined) {
    sequential = Number.parseFloat(options.sequentialStart)
    if (!Number.isFinite(sequential)) {
      return { ok: false, error: "Invalid channel number" }
    }
  }
  var resolvedNumber: string
  if (ids.length == 1 && options.numberChanged && Object.prototype.hasOwnProperty.call(patch, "x-channelID")) {
    resolvedNumber = mappingResolveChannelNumber(state, patch["x-channelID"])
    if (resolvedNumber === undefined) {
      return { ok: false, error: "Invalid channel number" }
    }
  }
  ids.forEach(function (id, index) {
    if (!Object.prototype.hasOwnProperty.call(state.draft, id)) {
      return
    }
    var channel = state.draft[id]
    Object.keys(patch).forEach(function (key) {
      if (key == "x-channelID" && ids.length > 1) {
        return
      }
      channel[key] = mappingDeepClone(patch[key])
    })
    if (resolvedNumber !== undefined) {
      channel["x-channelID"] = resolvedNumber
    }
    if (options.sequentialStart !== undefined) {
      channel["x-channelID"] = String(sequential + index)
    }
    if (channel["x-xmltv-file"] == "-" || channel["x-mapping"] == "-") {
      channel["x-active"] = false
    }
  })
  mappingDirtyIDs(state)
  return { ok: true }
}

function mappingAssignDummy(state: MappingWorkspaceState, ids: string[], program: string): MappingPatchResult {
  if (mappingDummyPrograms.indexOf(program) == -1) {
    return { ok: false, error: "Invalid dummy guide" }
  }
  return mappingApplyChannelPatch(state, ids, {
    "x-xmltv-file": "Threadfin Dummy",
    "x-mapping": program,
    "x-active": true,
    "x-update-channel-icon": false,
  })
}

function mappingRevertDraft(state: MappingWorkspaceState): void {
  if (mappingDraftIsLocked(state)) {
    return
  }
  state.draft = mappingDeepClone(state.baseline)
  state.dirtyIDs.clear()
  state.saveState = "idle"
  state.feedback = ""
}

function mappingReconcileAuthoritative(state: MappingWorkspaceState, authoritativeValue: any, submittedValue?: any, successful?: boolean): { persisted: boolean } {
  var authoritative = mappingDeepClone(mappingRecord(authoritativeValue))
  var previousBaseline = mappingDeepClone(state.baseline)
  var localDraft = mappingDeepClone(state.draft)
  var authoritativeComparable = mappingComparable(authoritative)
  var submittedComparable = submittedValue === undefined ? "" : mappingComparable(mappingRecord(submittedValue))
  var persisted = successful === true || authoritativeComparable == submittedComparable
  state.baseline = mappingDeepClone(authoritative)
  if (persisted) {
    state.draft = mappingDeepClone(authoritative)
  } else {
    state.draft = mappingDeepClone(authoritative)
    var ids = new Set<string>(Object.keys(previousBaseline).concat(Object.keys(localDraft)))
    ids.forEach(function (id) {
      if (mappingComparable(previousBaseline[id]) == mappingComparable(localDraft[id])) {
        return
      }
      if (mappingComparable(authoritative[id]) == mappingComparable(localDraft[id])) {
        return
      }
      if (Object.prototype.hasOwnProperty.call(localDraft, id)) {
        state.draft[id] = mappingDeepClone(localDraft[id])
      } else {
        delete state.draft[id]
      }
    })
    persisted = mappingDirtyIDs(state).length == 0 && authoritativeComparable != mappingComparable(previousBaseline)
  }
  mappingDirtyIDs(state)
  state.selected.forEach(function (id) {
    if (!Object.prototype.hasOwnProperty.call(state.draft, id)) state.selected.delete(id)
  })
  return { persisted: persisted }
}
