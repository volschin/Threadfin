"use strict";
class WizardCategory {
    constructor() {
        this.DocumentID = "content";
    }
    createCategoryHeadline(value) {
        var element = document.createElement("H2");
        element.textContent = value;
        return element;
    }
}
class WizardItem extends WizardCategory {
    constructor(key, headline) {
        super();
        this.headline = headline;
        this.key = key;
    }
    createWizard() {
        var headline = this.createCategoryHeadline(this.headline);
        headline.id = "wizard-field-label";
        var key = this.key;
        var content = new PopupContent();
        var description;
        var wizardField;
        var doc = document.getElementById(this.DocumentID);
        doc.innerHTML = "";
        doc.appendChild(headline);
        switch (key) {
            case "tuner":
                var text = new Array();
                var values = new Array();
                for (var i = 1; i <= 100; i++) {
                    text.push(i);
                    values.push(i);
                }
                var select = content.createSelect(text, values, "1", key);
                select.setAttribute("class", "wizard");
                select.id = key;
                wizardField = select;
                doc.appendChild(select);
                description = "{{.wizard.tuner.description}}";
                break;
            case "epgSource":
                var text = ["PMS", "XEPG"];
                var values = ["PMS", "XEPG"];
                var select = content.createSelect(text, values, "XEPG", key);
                select.setAttribute("class", "wizard");
                select.id = key;
                wizardField = select;
                doc.appendChild(select);
                description = "{{.wizard.epgSource.description}}";
                break;
            case "m3u":
                var input = content.createInput("text", key, "");
                input.setAttribute("placeholder", "{{.wizard.m3u.placeholder}}");
                input.setAttribute("class", "wizard");
                input.id = key;
                wizardField = input;
                doc.appendChild(input);
                description = "{{.wizard.m3u.description}}";
                break;
            case "xmltv":
                var input = content.createInput("text", key, "");
                input.setAttribute("placeholder", "{{.wizard.xmltv.placeholder}}");
                input.setAttribute("class", "wizard");
                input.id = key;
                wizardField = input;
                doc.appendChild(input);
                description = "{{.wizard.xmltv.description}}";
                break;
            default:
                break;
        }
        if (wizardField) {
            wizardField.setAttribute("aria-labelledby", headline.id);
            wizardField.setAttribute("aria-describedby", "wizard-description wizard-field-error");
        }
        var help = document.createElement("div");
        help.id = "wizard-description";
        help.className = "tf-wizard-description";
        help.innerHTML = description;
        doc.appendChild(help);
        var fieldError = document.createElement("p");
        fieldError.id = "wizard-field-error";
        fieldError.className = "tf-wizard-field-error";
        fieldError.hidden = true;
        doc.appendChild(fieldError);
    }
}
function readyForConfiguration(wizard) {
    var server = new Server("getServerConfig");
    server.request(new Object());
    showElement("loading", false);
    showConfigurationWizard(wizard);
}
function showConfigurationWizard(wizard) {
    configurationWizard[wizard].createWizard();
    var visibleSteps = wizard == 1 ? configurationWizard.length : configurationWizardVisibleSteps();
    var progress = document.querySelectorAll("#wizard-progress li");
    Array.prototype.forEach.call(progress, function (item, index) {
        item.hidden = index >= visibleSteps;
        if (index == wizard) {
            item.setAttribute("aria-current", "step");
        }
        else {
            item.removeAttribute("aria-current");
        }
    });
    var step = document.getElementById("wizard-step-status");
    if (step) {
        step.textContent = "{{.wizard.progress}}".replace("{current}", String(wizard + 1)).replace("{total}", String(visibleSteps));
    }
    var next = document.getElementById("next");
    if (next) {
        next.disabled = false;
        next.value = wizard == visibleSteps - 1 ? "{{.wizard.finish}}" : "{{.button.next}}";
    }
    var requestStatus = document.getElementById("wizard-request-status");
    if (requestStatus) {
        requestStatus.textContent = "";
    }
}
function configurationWizardVisibleSteps() {
    var server = typeof SERVER == "object" && SERVER ? SERVER : {};
    var settings = server.settings && typeof server.settings == "object" ? server.settings : {};
    var clientInfo = server.clientInfo && typeof server.clientInfo == "object" ? server.clientInfo : {};
    var epgSource = String(settings.epgSource || clientInfo.epgSource || "XEPG").toUpperCase();
    return epgSource == "PMS" ? 3 : configurationWizard.length;
}
function saveWizard() {
    var cmd = "saveWizard";
    var div = document.getElementById("content");
    var config = div.getElementsByClassName("wizard");
    var wizard = new Object();
    var error = document.getElementById("wizard-field-error");
    if (error) {
        error.textContent = "";
        error.hidden = true;
    }
    var invalid = div.querySelectorAll('[aria-invalid="true"]');
    Array.prototype.forEach.call(invalid, function (field) {
        field.removeAttribute("aria-invalid");
    });
    for (var i = 0; i < config.length; i++) {
        var name;
        var value;
        switch (config[i].tagName) {
            case "SELECT":
                name = config[i].name;
                value = config[i].value;
                // Wenn der Wert eine Zahl ist, wird dieser als Zahl gespeichert
                if (isNaN(value)) {
                    wizard[name] = value;
                }
                else {
                    wizard[name] = parseInt(value);
                }
                break;
            case "INPUT":
                switch (config[i].type) {
                    case "text":
                        name = config[i].name;
                        value = config[i].value;
                        if (value.length == 0) {
                            showWizardFieldError(config[i], name.toUpperCase() + ": " + "{{.alert.missingInput}}");
                            return;
                        }
                        if ((name == "m3u" || name == "xmltv") && !sourceLocationAccepted(value)) {
                            showWizardFieldError(config[i], "{{.sources.forms.locationInvalid}}");
                            return;
                        }
                        wizard[name] = value;
                        break;
                }
                break;
            default:
                // code...
                break;
        }
    }
    var data = new Object();
    data["wizard"] = wizard;
    var requestStatus = document.getElementById("wizard-request-status");
    if (requestStatus) {
        requestStatus.textContent = "{{.wizard.saving}}";
    }
    var next = document.getElementById("next");
    if (next) {
        next.disabled = true;
    }
    var server = new Server(cmd);
    server.request(data);
}
function showWizardFieldError(field, message) {
    var error = document.getElementById("wizard-field-error");
    if (error) {
        error.textContent = message;
        error.hidden = false;
    }
    field.setAttribute("aria-invalid", "true");
    field.focus();
}
function completeConfigurationWizardRequest(response) {
    var status = document.getElementById("wizard-request-status");
    var next = document.getElementById("next");
    if (next) {
        next.disabled = false;
    }
    if (!status) {
        return;
    }
    status.textContent = response && response.status === true ? "{{.wizard.saved}}" : response && response.status === false ? sourceString(response.err) || "{{.sources.responseInvalid}}" : "{{.sources.responseInvalid}}";
}
function completeConfigurationWizard() {
    window.location.replace("/web/#overview");
    window.location.reload();
}
// Wizard
var configurationWizard = new Array();
configurationWizard.push(new WizardItem("tuner", "{{.wizard.tuner.title}}"));
configurationWizard.push(new WizardItem("epgSource", "{{.wizard.epgSource.title}}"));
configurationWizard.push(new WizardItem("m3u", "{{.wizard.m3u.title}}"));
configurationWizard.push(new WizardItem("xmltv", "{{.wizard.xmltv.title}}"));
