type FilterType = "group-title" | "custom-filter"
type FilterFeedbackState = "progress" | "success" | "error"

interface FilterListItem {
  id: string
  type: FilterType
  name: string
  startingNumber: string
  rule: string
  summary: string
}

interface FilterStreamCounts {
  imported: number
  selected: number
  excluded: number
}

var filterPageFeedback: { [key: string]: { state: FilterFeedbackState, message: string } } = {}
var filterPopupInvoker: HTMLElement
var filterPopupFocusKey = ""
var filterPopupFocusListenerAttached = false
var filterFocusAfterDeletion = false
var filterDeletionFallback: HTMLElement

function filterRecord(value: any): { [key: string]: any } {
  return value && typeof value == "object" && !Array.isArray(value) ? value : {}
}

function filterString(value: any): string {
  return typeof value == "string" ? value : value === undefined || value === null ? "" : String(value)
}

function filterTerms(value: any): string[] {
  return filterString(value).split(",").map(function (term) { return term.trim() }).filter(function (term) { return term.length > 0 })
}

function filterQuotedTerms(value: any): string {
  return filterTerms(value).map(function (term) { return '"' + term + '"' }).join(" or ")
}

function filterRuleSummary(filter: any): string {
  var values = filterRecord(filter)
  var caseDescription = values.caseSensitive === true ? "case-sensitive" : "case-insensitive"
  if (values.type == "custom-filter") {
    return filterCustomRuleSummary(filterString(values.filter), caseDescription)
  }
  var summary = "Matches streams whose group is exactly \"" + filterString(values.filter) + "\" (" + caseDescription + ")."
  var include = filterQuotedTerms(values.include)
  if (include) {
    summary += " Includes stream names containing any of " + include + "."
  }
  var exclude = filterQuotedTerms(values.exclude)
  if (exclude) {
    summary += " Excludes stream names containing any of " + exclude + "."
  }
  return summary
}

function filterCanonicalCustomTerms(value: string): string[] {
  if (!value || value.indexOf(".") != -1) {
    return undefined
  }
  var terms = value.split(",").map(function (term) { return term.trim() })
  return terms.every(function (term) { return term.length > 0 }) ? terms : undefined
}

function filterQuotedCanonicalTerms(terms: string[]): string {
  return terms.map(function (term) { return '"' + term + '"' }).join(" or ")
}

function filterCustomRuleSummary(rule: string, caseDescription: string): string {
  // Canonical subset of the backend grammar: base, one optional include clause,
  // then one optional exclude clause, with a single separator space per clause.
  var parsed = /^([^\s{}!](?:[^{}!]*[^\s{}!])?)(?: \{([^{}!]+)\})?(?: !\{([^{}!]+)\})?$/.exec(rule)
  var includeTerms = parsed ? filterCanonicalCustomTerms(parsed[2]) : undefined
  var excludeTerms = parsed ? filterCanonicalCustomTerms(parsed[3]) : undefined
  if (!parsed || (parsed[2] !== undefined && !includeTerms) || (parsed[3] !== undefined && !excludeTerms)) {
    return "The custom rule will be evaluated as entered against complete stream data (" + caseDescription + "). Its include/exclude clauses are not summarized."
  }
  var summary = "Matches complete stream data containing \"" + parsed[1] + "\" (" + caseDescription + ")."
  var include = includeTerms ? filterQuotedCanonicalTerms(includeTerms) : ""
  if (include) {
    summary += " Includes complete stream data containing any of " + include + "."
  }
  var exclude = excludeTerms ? filterQuotedCanonicalTerms(excludeTerms) : ""
  if (exclude) {
    summary += " Excludes complete stream data containing any of " + exclude + "."
  }
  return summary
}

function filterStreamCounts(server: any): FilterStreamCounts {
  var data = filterRecord(filterRecord(server).data)
  var preview = filterRecord(data.StreamPreviewUI)
  var selected = Array.isArray(preview.activeStreams) ? preview.activeStreams.length : 0
  var excluded = Array.isArray(preview.inactiveStreams) ? preview.inactiveStreams.length : 0
  return { imported: selected + excluded, selected: selected, excluded: excluded }
}

function selectFilterList(server: any): FilterListItem[] {
  var settings = filterRecord(filterRecord(server).settings)
  var saved = filterRecord(settings.filter)
  var filters: FilterListItem[] = []
  Object.keys(saved).forEach(function (id) {
    if (id == "-1") {
      return
    }
    var filter = filterRecord(saved[id])
    if (filter.type != "group-title" && filter.type != "custom-filter") {
      return
    }
    filters.push({
      id: id,
      type: filter.type as FilterType,
      name: filterString(filter.name) || "{{.filter.unnamed}}",
      startingNumber: filterString(filter.startingNumber),
      rule: filterString(filter.filter),
      summary: filterRuleSummary(filter),
    })
  })
  filters.sort(function (left, right) {
    var typeOrder = left.type == right.type ? 0 : left.type == "group-title" ? -1 : 1
    if (typeOrder != 0) {
      return typeOrder
    }
    var nameOrder = left.name.localeCompare(right.name)
    return nameOrder == 0 ? left.id.localeCompare(right.id) : nameOrder
  })
  return filters
}

function renderFilterManagementPage(host: HTMLElement): void {
  host.innerHTML = ""
  var page = document.createElement("div")
  page.className = "tf-filters"

  var header = document.createElement("header")
  header.className = "tf-filters-header"
  var titles = document.createElement("div")
  var title = document.createElement("h1")
  title.id = "filter-heading"
  title.tabIndex = -1
  title.textContent = "Filter"
  var purpose = document.createElement("p")
  purpose.textContent = "Filters select which imported streams become channels available for Mapping."
  titles.appendChild(title)
  titles.appendChild(purpose)
  header.appendChild(titles)
  var addGroup = createFilterAction("{{.filter.addGroup}}", "add-group", function (button) { openFilterPopup("group-title", undefined, button) })
  header.appendChild(addGroup)
  page.appendChild(header)

  var counts = filterStreamCounts(SERVER)
  var metrics = document.createElement("dl")
  metrics.className = "tf-filter-counts"
  ;[
    ["{{.filter.counts.imported}}", counts.imported],
    ["{{.filter.counts.selected}}", counts.selected],
    ["{{.filter.counts.excluded}}", counts.excluded],
  ].forEach(function (count) {
    var group = document.createElement("div")
    var term = document.createElement("dt")
    term.textContent = count[0] as string
    var value = document.createElement("dd")
    value.textContent = String(count[1])
    group.appendChild(term)
    group.appendChild(value)
    metrics.appendChild(group)
  })
  page.appendChild(metrics)

  var feedback = document.createElement("p")
  feedback.className = "tf-filter-page-status"
  feedback.setAttribute("role", "status")
  feedback.setAttribute("aria-live", "polite")
  var pageFeedback = filterPageFeedback.filter
  if (pageFeedback) {
    feedback.textContent = pageFeedback.message
    feedback.setAttribute("data-state", pageFeedback.state)
  } else {
    feedback.hidden = true
  }
  page.appendChild(feedback)

  var filters = selectFilterList(SERVER)
  if (filters.length == 0) {
    var empty = document.createElement("section")
    empty.className = "tf-filter-empty"
    var emptyTitle = document.createElement("h2")
    emptyTitle.textContent = "{{.filter.empty.title}}"
    var emptyCopy = document.createElement("p")
    emptyCopy.textContent = "{{.filter.empty.description}}"
    empty.appendChild(emptyTitle)
    empty.appendChild(emptyCopy)
    empty.appendChild(createFilterAction("{{.filter.addGroup}}", "empty-add-group", function (button) { openFilterPopup("group-title", undefined, button) }))
    page.appendChild(empty)
  } else {
    var list = document.createElement("div")
    list.className = "tf-filter-list"
    list.setAttribute("role", "list")
    filters.forEach(function (filter) { list.appendChild(renderFilterRow(filter)) })
    page.appendChild(list)
  }

  var advanced = document.createElement("section")
  advanced.className = "tf-filter-advanced"
  var advancedTitle = document.createElement("h2")
  advancedTitle.textContent = "{{.filter.advanced.title}}"
  var advancedCopy = document.createElement("p")
  advancedCopy.textContent = "{{.filter.advanced.description}}"
  advanced.appendChild(advancedTitle)
  advanced.appendChild(advancedCopy)
  advanced.appendChild(createFilterAction("{{.filter.advanced.chooseType}}", "choose-type", function (button) { openFilterPopup("filter", undefined, button) }))
  advanced.appendChild(createFilterAction("{{.filter.advanced.addCustom}}", "add-custom", function (button) { openFilterPopup("custom-filter", undefined, button) }))
  page.appendChild(advanced)
  host.appendChild(page)
  if (filterFocusAfterDeletion) {
    filterDeletionFallback = addGroup || title
  }
}

function createFilterAction(label: string, focusKey: string, listener: (button: HTMLButtonElement) => void): HTMLButtonElement {
  var button = document.createElement("button")
  button.type = "button"
  button.className = "tf-filter-action"
  button.textContent = label
  button.setAttribute("data-filter-focus-key", focusKey)
  button.addEventListener("click", function () { listener(button) })
  return button
}

function renderFilterRow(filter: FilterListItem): HTMLElement {
  var row = document.createElement("article")
  row.className = "tf-filter-row"
  row.setAttribute("role", "listitem")
  row.setAttribute("data-filter-id", filter.id)
  row.setAttribute("data-filter-type", filter.type)
  var identity = document.createElement("div")
  var name = document.createElement("h2")
  name.textContent = filter.name
  var meta = document.createElement("p")
  meta.textContent = (filter.type == "group-title" ? "{{.filter.group}}" : "{{.filter.custom}}") + " · {{.filter.startingnumber.title}}: " + filter.startingNumber
  var rule = document.createElement("code")
  rule.textContent = filter.rule
  var summary = document.createElement("p")
  summary.className = "tf-filter-summary"
  summary.textContent = filter.summary
  identity.appendChild(name)
  identity.appendChild(meta)
  identity.appendChild(rule)
  identity.appendChild(summary)
  row.appendChild(identity)
  var actions = document.createElement("div")
  actions.className = "tf-filter-actions"
  actions.appendChild(createFilterAction("{{.sources.actions.edit}}", filter.id + ":edit", function (button) { openFilterPopup(filter.type, { id: filter.id }, button) }))
  var remove = createFilterAction("{{.sources.actions.delete}}", filter.id + ":delete", function (button) {
    openFilterPopup(filter.type, { id: filter.id }, button)
    savePopupData("filter", filter.id, true, 0)
  })
  remove.className += " tf-filter-delete-action"
  actions.appendChild(remove)
  row.appendChild(actions)
  return row
}

function openFilterPopup(dataType: string, element: any, invoker: HTMLElement): void {
  filterPopupInvoker = invoker
  filterPopupFocusKey = invoker ? filterString(invoker.getAttribute("data-filter-focus-key")) : ""
  var modal = document.getElementById("popup")
  if (modal && !filterPopupFocusListenerAttached) {
    filterPopupFocusListenerAttached = true
    modal.addEventListener("hidden.bs.modal", function () {
      var target: HTMLElement
      if (filterFocusAfterDeletion) {
        target = filterDeletionFallback || filterPopupReplacement("add-group") || document.getElementById("filter-heading") as HTMLElement
        filterFocusAfterDeletion = false
        filterDeletionFallback = undefined
      } else {
        target = filterPopupInvoker && document.contains(filterPopupInvoker) ? filterPopupInvoker : filterPopupReplacement(filterPopupFocusKey)
      }
      if (target) {
        target.focus()
      }
      filterPopupInvoker = undefined
      filterPopupFocusKey = ""
    })
  }
  openPopUp(dataType, element)
}

function filterPopupReplacement(focusKey: string): HTMLElement {
  if (!focusKey) {
    return undefined
  }
  var candidates = document.querySelectorAll("[data-filter-focus-key]")
  for (var index = 0; index < candidates.length; index++) {
    var candidate = candidates[index] as HTMLElement
    if (candidate.getAttribute("data-filter-focus-key") == focusKey) {
      return candidate
    }
  }
  return undefined
}

function enhanceFilterPopup(dataType: string): void {
  var popup = document.getElementById("popup-custom")
  if (!popup) {
    return
  }
  popup.classList.remove("tf-filter-popup")
  if (dataType != "group-title" && dataType != "custom-filter") {
    return
  }
  popup.classList.add("tf-filter-popup")
  var fields = popup.querySelectorAll("input, select")
  Array.prototype.forEach.call(fields, function (field: HTMLElement) {
    var row = field.closest("tr")
    var title = row ? row.querySelector("td:first-child") : undefined
    if (title && title.textContent) {
      field.setAttribute("aria-label", title.textContent.replace(/:\s*$/, ""))
    }
  })
  ;["name", "filter"].forEach(function (fieldName) {
    var field = popup.querySelector('[name="' + fieldName + '"]') as HTMLInputElement
    if (!field || !field.parentElement) {
      return
    }
    field.setAttribute("aria-required", "true")
    var error = document.createElement("p")
    error.id = "filter-" + fieldName + "-error"
    error.className = "tf-filter-field-error"
    error.hidden = true
    field.setAttribute("aria-errormessage", error.id)
    field.parentElement.appendChild(error)
  })
  var example = document.createElement("p")
  example.className = "tf-filter-field-help"
  example.textContent = "{{.filter.example}}"
  popup.appendChild(example)
  var summary = document.createElement("p")
  summary.id = "filter-rule-summary"
  summary.className = "tf-filter-rule-summary"
  summary.setAttribute("role", "status")
  summary.setAttribute("aria-live", "polite")
  popup.appendChild(summary)
  var status = document.createElement("p")
  status.id = "filter-form-status"
  status.className = "tf-filter-form-status"
  status.setAttribute("role", "status")
  status.setAttribute("aria-live", "polite")
  status.hidden = true
  popup.appendChild(status)
  var update = function () { filterUpdatePopupSummary() }
  Array.prototype.forEach.call(fields, function (field: HTMLElement) {
    field.addEventListener("input", update)
    field.addEventListener("change", update)
  })
  filterUpdatePopupSummary()
}

function filterPopupValues(): { [key: string]: any } {
  var popup = document.getElementById("popup-custom")
  var values: { [key: string]: any } = {}
  if (!popup) {
    return values
  }
  var fields = popup.querySelectorAll("input, select")
  Array.prototype.forEach.call(fields, function (field: HTMLInputElement) {
    if (!field.name) {
      return
    }
    values[field.name] = field.type == "checkbox" ? field.checked : field.value
  })
  return values
}

function filterUpdatePopupSummary(): void {
  var summary = document.getElementById("filter-rule-summary")
  if (summary) {
    summary.textContent = filterRuleSummary(filterPopupValues())
  }
}

function validateFilterPopup(dataType: string): boolean {
  if (dataType != "group-title" && dataType != "custom-filter") {
    return true
  }
  var popup = document.getElementById("popup-custom")
  if (!popup) {
    return false
  }
  filterClearPopupErrors()
  var name = popup.querySelector('[name="name"]') as HTMLInputElement
  if (!name || !name.value.trim()) {
    filterSetPopupError(name, "filter-name-error", "{{.filter.validation.nameRequired}}")
    return false
  }
  var rule = popup.querySelector('[name="filter"]') as HTMLInputElement
  if (!rule || !rule.value.trim()) {
    filterSetPopupError(rule, "filter-filter-error", "{{.filter.validation.ruleRequired}}")
    return false
  }
  return true
}

function filterClearPopupErrors(): void {
  var popup = document.getElementById("popup-custom")
  if (!popup) {
    return
  }
  Array.prototype.forEach.call(popup.querySelectorAll('[aria-invalid="true"]'), function (field: HTMLElement) { field.removeAttribute("aria-invalid") })
  Array.prototype.forEach.call(popup.querySelectorAll(".tf-filter-field-error"), function (error: HTMLElement) { error.textContent = ""; error.hidden = true })
}

function filterSetPopupError(field: HTMLInputElement, errorID: string, message: string): void {
  if (field) {
    field.setAttribute("aria-invalid", "true")
  }
  var error = document.getElementById(errorID)
  if (error) {
    error.textContent = message
    error.hidden = false
  }
  if (field) {
    field.focus()
  }
}

function filterSetFormStatus(message: string, state: FilterFeedbackState): void {
  var status = document.getElementById("filter-form-status")
  if (!status) {
    return
  }
  status.textContent = message
  status.setAttribute("data-state", state)
  status.hidden = false
}

function beginFilterRequest(dataType: string, id: string, remove: Boolean): void {
  if (dataType != "filter") {
    return
  }
  filterSetFormStatus(remove == true ? "{{.filter.feedback.deleting}}" : "{{.filter.feedback.saving}}", "progress")
}

function completeFilterRequest(command: string, data: any, response: any): void {
  if (command != "saveFilter") {
    return
  }
  var result = filterRecord(response)
  if (result.status !== true) {
    var error = result.status === false ? filterString(result.err) || "{{.filter.feedback.error}}" : "{{.sources.responseInvalid}}"
    filterSetFormStatus(error, "error")
    filterPageFeedback.filter = { state: "error", message: error }
    return
  }
  var request = filterRecord(filterRecord(data).filter)
  var ids = Object.keys(request)
  var deleted = ids.length == 1 && filterRecord(request[ids[0]]).delete === true
  if (deleted) {
    filterFocusAfterDeletion = true
  }
  filterPageFeedback.filter = { state: "success", message: deleted ? "{{.filter.feedback.deleted}}" : "{{.filter.feedback.saved}}" }
}
