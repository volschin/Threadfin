"use strict";
class SettingsCategory {
    constructor() {
        this.DocumentID = "content_settings";
    }
    createCategoryHeadline(value) {
        var element = document.createElement("H4");
        element.innerHTML = value;
        return element;
    }
    createHR() {
        var element = document.createElement("HR");
        return element;
    }
    createSettings(settingsKey) {
        var content = new PopupContent();
        var data = SERVER["settings"][settingsKey];
        switch (settingsKey) {
            // Texteingaben
            case "update":
                var title = "{{.settings.update.title}}" + ":";
                var input = content.createInput("text", "update", data.toString());
                input.setAttribute("placeholder", "{{.settings.update.placeholder}}");
                addRuntimeEventListener(input, "change", (event) => { event.currentTarget.classList.add("changed"); });
                break;
            case "backup.path":
                var title = "{{.settings.backupPath.title}}" + ":";
                var input = content.createInput("text", "backup.path", data);
                input.setAttribute("placeholder", "{{.settings.backupPath.placeholder}}");
                addRuntimeEventListener(input, "change", (event) => { event.currentTarget.classList.add("changed"); });
                break;
            case "temp.path":
                var title = "{{.settings.tempPath.title}}" + ":";
                var input = content.createInput("text", "temp.path", data);
                input.setAttribute("placeholder", "{{.settings.tmpPath.placeholder}}");
                addRuntimeEventListener(input, "change", (event) => { event.currentTarget.classList.add("changed"); });
                break;
            case "user.agent":
                var title = "{{.settings.userAgent.title}}" + ":";
                var input = content.createInput("text", "user.agent", data);
                input.setAttribute("placeholder", "{{.settings.userAgent.placeholder}}");
                addRuntimeEventListener(input, "change", (event) => { event.currentTarget.classList.add("changed"); });
                break;
            case "buffer.timeout":
                var title = "{{.settings.bufferTimeout.title}}" + ":";
                var input = content.createInput("text", "buffer.timeout", data);
                input.setAttribute("placeholder", "{{.settings.bufferTimeout.placeholder}}");
                addRuntimeEventListener(input, "change", (event) => { event.currentTarget.classList.add("changed"); });
                break;
            case "ffmpeg.path":
                var title = "{{.settings.ffmpegPath.title}}" + ":";
                var input = content.createInput("text", "ffmpeg.path", data);
                input.setAttribute("placeholder", "{{.settings.ffmpegPath.placeholder}}");
                addRuntimeEventListener(input, "change", (event) => { event.currentTarget.classList.add("changed"); });
                break;
            case "ffmpeg.options":
                var title = "{{.settings.ffmpegOptions.title}}" + ":";
                var input = content.createInput("text", "ffmpeg.options", data);
                input.setAttribute("placeholder", "{{.settings.ffmpegOptions.placeholder}}");
                addRuntimeEventListener(input, "change", (event) => { event.currentTarget.classList.add("changed"); });
                break;
            case "ffmpeg.forceHttp":
                var title = "{{.settings.ffmpegForceHttp.title}}" + ":";
                var input = content.createCheckbox(settingsKey);
                input.checked = data;
                addRuntimeEventListener(input, "change", (event) => { event.currentTarget.classList.add("changed"); });
                break;
            case "vlc.path":
                var title = "{{.settings.vlcPath.title}}" + ":";
                var input = content.createInput("text", "vlc.path", data);
                input.setAttribute("placeholder", "{{.settings.vlcPath.placeholder}}");
                addRuntimeEventListener(input, "change", (event) => { event.currentTarget.classList.add("changed"); });
                break;
            case "vlc.options":
                var title = "{{.settings.vlcOptions.title}}" + ":";
                var input = content.createInput("text", "vlc.options", data);
                input.setAttribute("placeholder", "{{.settings.vlcOptions.placeholder}}");
                addRuntimeEventListener(input, "change", (event) => { event.currentTarget.classList.add("changed"); });
                break;
            // Checkboxen
            case "authentication.web":
                var title = "{{.settings.authenticationWEB.title}}" + ":";
                var input = content.createCheckbox(settingsKey);
                input.checked = data;
                addRuntimeEventListener(input, "change", (event) => { event.currentTarget.classList.add("changed"); });
                break;
            case "authentication.pms":
                var title = "{{.settings.authenticationPMS.title}}" + ":";
                var input = content.createCheckbox(settingsKey);
                input.checked = data;
                addRuntimeEventListener(input, "change", (event) => { event.currentTarget.classList.add("changed"); });
                break;
            case "authentication.m3u":
                var title = "{{.settings.authenticationM3U.title}}" + ":";
                var input = content.createCheckbox(settingsKey);
                input.checked = data;
                addRuntimeEventListener(input, "change", (event) => { event.currentTarget.classList.add("changed"); });
                break;
            case "authentication.xml":
                var title = "{{.settings.authenticationXML.title}}" + ":";
                var input = content.createCheckbox(settingsKey);
                input.checked = data;
                addRuntimeEventListener(input, "change", (event) => { event.currentTarget.classList.add("changed"); });
                break;
            case "authentication.api":
                var title = "{{.settings.authenticationAPI.title}}" + ":";
                var input = content.createCheckbox(settingsKey);
                input.checked = data;
                addRuntimeEventListener(input, "change", (event) => { event.currentTarget.classList.add("changed"); });
                break;
            case "files.update":
                var title = "{{.settings.filesUpdate.title}}" + ":";
                var input = content.createCheckbox(settingsKey);
                input.checked = data;
                addRuntimeEventListener(input, "change", (event) => { event.currentTarget.classList.add("changed"); });
                break;
            case "cache.images":
                var title = "{{.settings.cacheImages.title}}" + ":";
                var input = content.createCheckbox(settingsKey);
                input.checked = data;
                addRuntimeEventListener(input, "change", (event) => { event.currentTarget.classList.add("changed"); });
                break;
            case "xepg.replace.missing.images":
                var title = "{{.settings.replaceEmptyImages.title}}" + ":";
                var input = content.createCheckbox(settingsKey);
                input.checked = data;
                addRuntimeEventListener(input, "change", (event) => { event.currentTarget.classList.add("changed"); });
                break;
            case "xepg.replace.channel.title":
                var title = "{{.settings.replaceChannelTitle.title}}" + ":";
                var input = content.createCheckbox(settingsKey);
                input.checked = data;
                addRuntimeEventListener(input, "change", (event) => { event.currentTarget.classList.add("changed"); });
                break;
            case "storeBufferInRAM":
                var title = "{{.settings.storeBufferInRAM.title}}" + ":";
                var input = content.createCheckbox(settingsKey);
                input.checked = data;
                addRuntimeEventListener(input, "change", (event) => { event.currentTarget.classList.add("changed"); });
                break;
            case "forceHttps":
                var title = "{{.settings.forceHttps.title}}" + ":";
                var input = content.createCheckbox(settingsKey);
                input.checked = data;
                addRuntimeEventListener(input, "change", (event) => { event.currentTarget.classList.add("changed"); });
                break;
            case "excludeStreamHttps":
                var title = "{{.settings.excludeStreamHttps.title}}" + ":";
                var input = content.createCheckbox(settingsKey);
                input.checked = data;
                addRuntimeEventListener(input, "change", (event) => { event.currentTarget.classList.add("changed"); });
                break;
            case "httpsPort":
                var title = "{{.settings.httpsPort.title}}" + ":";
                var input = content.createInput("text", "httpsPort", data.toString());
                input.setAttribute("placeholder", "{{.settings.httpsPort.placeholder}}");
                addRuntimeEventListener(input, "change", (event) => { event.currentTarget.classList.add("changed"); });
                break;
            case "httpsThreadfinDomain":
                var title = "{{.settings.httpsThreadfinDomain.title}}" + ":";
                var input = content.createInput("text", "httpsThreadfinDomain", data.toString());
                input.setAttribute("placeholder", "{{.settings.httpsThreadfinDomain.placeholder}}");
                addRuntimeEventListener(input, "change", (event) => { event.currentTarget.classList.add("changed"); });
                break;
            case "bindIpAddress":
                var title = "{{.settings.bindIpAddress.title}}" + ":";
                var input = content.createInput("text", "bindIpAddress", data.toString());
                input.setAttribute("placeholder", "{{.settings.bindIpAddress.placeholder}}");
                addRuntimeEventListener(input, "change", (event) => { event.currentTarget.classList.add("changed"); });
                break;
            case "httpThreadfinDomain":
                var title = "{{.settings.httpThreadfinDomain.title}}" + ":";
                var input = content.createInput("text", "httpThreadfinDomain", data.toString());
                input.setAttribute("placeholder", "{{.settings.httpThreadfinDomain.placeholder}}");
                addRuntimeEventListener(input, "change", (event) => { event.currentTarget.classList.add("changed"); });
                break;
            case "enableNonAscii":
                var title = "{{.settings.enableNonAscii.title}}" + ":";
                var input = content.createCheckbox(settingsKey);
                input.checked = data;
                addRuntimeEventListener(input, "change", (event) => { event.currentTarget.classList.add("changed"); });
                break;
            case "epgCategories":
                var title = "{{.settings.epgCategories.title}}" + ":";
                var input = content.createInput("text", "epgCategories", data.toString());
                input.setAttribute("placeholder", "{{.settings.epgCategories.placeholder}}");
                addRuntimeEventListener(input, "change", (event) => { event.currentTarget.classList.add("changed"); });
                break;
            case "epgCategoriesColors":
                var title = "{{.settings.epgCategoriesColors.title}}" + ":";
                var input = content.createInput("text", "epgCategoriesColors", data.toString());
                input.setAttribute("placeholder", "{{.settings.epgCategoriesColors.placeholder}}");
                addRuntimeEventListener(input, "change", (event) => { event.currentTarget.classList.add("changed"); });
                break;
            case "ThreadfinAutoUpdate":
                var title = "{{.settings.ThreadfinAutoUpdate.title}}" + ":";
                var input = content.createCheckbox(settingsKey);
                input.checked = data;
                addRuntimeEventListener(input, "change", (event) => { event.currentTarget.classList.add("changed"); });
                break;
            case "ssdp":
                var title = "{{.settings.ssdp.title}}" + ":";
                var input = content.createCheckbox(settingsKey);
                input.checked = data;
                addRuntimeEventListener(input, "change", (event) => { event.currentTarget.classList.add("changed"); });
                break;
            case "dummy":
                var title = "{{.settings.dummy.title}}" + ":";
                var input = content.createCheckbox(settingsKey);
                input.checked = data;
                addRuntimeEventListener(input, "change", (event) => { event.currentTarget.classList.add("changed"); });
                break;
            case "dummyChannel":
                var title = "{{.settings.dummyChannel.title}}" + ":";
                var text = ["PPV", "30 Minutes", "60 Minutes", "90 Minutes", "120 Minutes", "180 Minutes", "240 Minutes", "360 Minutes"];
                var values = ["PPV", "30_Minutes", "60_Minutes", "90_Minutes", "120_Minutes", "180_Minutes", "240_Minutes", "360_Minutes"];
                var select = content.createSelect(text, values, data, settingsKey);
                addRuntimeEventListener(select, "change", (event) => { event.currentTarget.classList.add("changed"); });
                break;
            case "ignoreFilters":
                var title = "{{.settings.ignoreFilters.title}}" + ":";
                var input = content.createCheckbox(settingsKey);
                input.checked = data;
                addRuntimeEventListener(input, "change", (event) => { event.currentTarget.classList.add("changed"); });
                break;
            case "api":
                var title = "{{.settings.api.title}}" + ":";
                var input = content.createCheckbox(settingsKey);
                input.checked = data;
                addRuntimeEventListener(input, "change", (event) => { event.currentTarget.classList.add("changed"); });
                break;
            // Select
            case "tuner":
                var title = "{{.settings.tuner.title}}" + ":";
                var text = new Array();
                var values = new Array();
                for (var i = 1; i <= 100; i++) {
                    text.push(i);
                    values.push(i);
                }
                var select = content.createSelect(text, values, data, settingsKey);
                addRuntimeEventListener(select, "change", (event) => { event.currentTarget.classList.add("changed"); });
                break;
            case "epgSource":
                var title = "{{.settings.epgSource.title}}" + ":";
                var text = ["PMS", "XEPG"];
                var values = ["PMS", "XEPG"];
                var select = content.createSelect(text, values, data, settingsKey);
                addRuntimeEventListener(select, "change", (event) => { event.currentTarget.classList.add("changed"); });
                break;
            case "backup.keep":
                var title = "{{.settings.backupKeep.title}}" + ":";
                var text = ["5", "10", "20", "30", "40", "50"];
                var values = ["5", "10", "20", "30", "40", "50"];
                var select = content.createSelect(text, values, data, settingsKey);
                addRuntimeEventListener(select, "change", (event) => { event.currentTarget.classList.add("changed"); });
                break;
            case "buffer.size.kb":
                var title = "{{.settings.bufferSize.title}}" + ":";
                var text = ["0.5 MB", "1 MB", "2 MB", "3 MB", "4 MB", "5 MB", "6 MB", "7 MB", "8 MB"];
                var values = ["512", "1024", "2048", "3072", "4096", "5120", "6144", "7168", "8192"];
                var select = content.createSelect(text, values, data, settingsKey);
                addRuntimeEventListener(select, "change", (event) => { event.currentTarget.classList.add("changed"); });
                break;
            case "buffer":
                var title = "{{.settings.streamBuffering.title}}" + ":";
                var text = ["{{.settings.streamBuffering.info_false}}", "FFmpeg: ({{.settings.streamBuffering.info_ffmpeg}})", "VLC: ({{.settings.streamBuffering.info_vlc}})"];
                var values = ["-", "ffmpeg", "vlc"];
                var select = content.createSelect(text, values, data, settingsKey);
                addRuntimeEventListener(select, "change", (event) => { event.currentTarget.classList.add("changed"); });
                break;
            case "udpxy":
                ;
                var title = "{{.settings.udpxy.title}}" + ":";
                var input = content.createInput("text", "udpxy", data);
                input.setAttribute("placeholder", "{{.settings.udpxy.placeholder}}");
                addRuntimeEventListener(input, "change", (event) => { event.currentTarget.classList.add("changed"); });
                break;
        }
        var control = input || select;
        if (!control)
            return document.createElement("div");
        return createUIFormField(title, control, "settings-field-" + settingsKey.replace(/[^A-Za-z0-9_-]+/g, "-"));
    }
    createDescription(settingsKey) {
        var text;
        switch (settingsKey) {
            case "authentication.web":
                text = "{{.settings.authenticationWEB.description}}";
                break;
            case "authentication.m3u":
                text = "{{.settings.authenticationM3U.description}}";
                break;
            case "authentication.pms":
                text = "{{.settings.authenticationPMS.description}}";
                break;
            case "authentication.xml":
                text = "{{.settings.authenticationXML.description}}";
                break;
            case "authentication.api":
                if (SERVER["settings"]["authentication.web"] == true) {
                    text = "{{.settings.authenticationAPI.description}}";
                }
                break;
            case "ThreadfinAutoUpdate":
                text = "{{.settings.ThreadfinAutoUpdate.description}}";
                break;
            case "backup.keep":
                text = "{{.settings.backupKeep.description}}";
                break;
            case "backup.path":
                text = "{{.settings.backupPath.description}}";
                break;
            case "temp.path":
                text = "{{.settings.tempPath.description}}";
                break;
            case "buffer":
                text = "{{.settings.streamBuffering.description}}";
                break;
            case "buffer.size.kb":
                text = "{{.settings.bufferSize.description}}";
                break;
            case "storeBufferInRAM":
                text = "{{.settings.storeBufferInRAM.description}}";
                break;
            case "forceHttps":
                text = "{{.settings.forceHttps.description}}";
                break;
            case "excludeStreamHttps":
                text = "{{.settings.excludeStreamHttps.description}}";
                break;
            case "httpsPort":
                text = "{{.settings.httpsPort.description}}";
                break;
            case "httpsThreadfinDomain":
                text = "{{.settings.httpsThreadfinDomain.description}}";
                break;
            case "bindIpAddress":
                text = "{{.settings.bindIpAddress.description}}";
                break;
            case "httpThreadfinDomain":
                text = "{{.settings.httpThreadfinDomain.description}}";
                break;
            case "enableNonAscii":
                text = "{{.settings.enableNonAscii.description}}";
                break;
            case "epgCategories":
                text = "{{.settings.epgCategories.description}}";
                break;
            case "epgCategoriesColors":
                text = "{{.settings.epgCategoriesColors.description}}";
                break;
            case "buffer.timeout":
                text = "{{.settings.bufferTimeout.description}}";
                break;
            case "user.agent":
                text = "{{.settings.userAgent.description}}";
                break;
            case "ffmpeg.path":
                text = "{{.settings.ffmpegPath.description}}";
                break;
            case "ffmpeg.options":
                text = "{{.settings.ffmpegOptions.description}}";
                break;
            case "ffmpeg.forceHttp":
                text = "{{.settings.ffmpegForceHttp.description}}";
                break;
            case "vlc.path":
                text = "{{.settings.vlcPath.description}}";
                break;
            case "vlc.options":
                text = "{{.settings.vlcOptions.description}}";
                break;
            case "epgSource":
                text = "{{.settings.epgSource.description}}";
                break;
            case "tuner":
                text = "{{.settings.tuner.description}}";
                break;
            case "update":
                text = "{{.settings.update.description}}";
                break;
            case "api":
                text = "{{.settings.api.description}}";
                break;
            case "ssdp":
                text = "{{.settings.ssdp.description}}";
                break;
            case "files.update":
                text = "{{.settings.filesUpdate.description}}";
                break;
            case "cache.images":
                text = "{{.settings.cacheImages.description}}";
                break;
            case "xepg.replace.missing.images":
                text = "{{.settings.replaceEmptyImages.description}}";
                break;
            case "xepg.replace.channel.title":
                text = "{{.settings.replaceChannelTitle.description}}";
                break;
            case "udpxy":
                text = "{{.settings.udpxy.description}}";
                break;
            default:
                text = "";
                break;
        }
        var help = document.createElement("div");
        help.className = "tf-field-help";
        help.innerHTML = text;
        return help;
    }
}
class SettingsCategoryItem extends SettingsCategory {
    constructor(headline, settingsKeys) {
        super();
        this.headline = headline;
        this.settingsKeys = settingsKeys;
    }
    createCategory() {
        var headline = this.createCategoryHeadline(this.headline);
        var settingsKeys = this.settingsKeys;
        var doc = document.getElementById(this.DocumentID);
        doc.appendChild(headline);
        // Tabelle für die Kategorie erstellen
        var table = document.createElement("div");
        table.className = "tf-form-fields";
        var keys = settingsKeys.split(",");
        keys.forEach(settingsKey => {
            switch (settingsKey) {
                case "authentication.pms":
                case "authentication.m3u":
                case "authentication.xml":
                case "authentication.api":
                    if (SERVER["settings"]["authentication.web"] == false) {
                        break;
                    }
                default:
                    var item = this.createSettings(settingsKey);
                    var description = this.createDescription(settingsKey);
                    table.appendChild(item);
                    appendUIFormHelp(item, description);
                    break;
            }
        });
        doc.appendChild(table);
        doc.appendChild(this.createHR());
    }
}
function showSettings() {
    console.log("SETTINGS");
    for (let i = 0; i < settingsCategory.length; i++) {
        settingsCategory[i].createCategory();
    }
}
function saveSettings() {
    if (settingsFeedback.state == "progress")
        return;
    var root = document.querySelector(".tf-settings");
    if (!root || !validateSettingsForm(root))
        return;
    var controls = Array.prototype.slice.call(root.querySelectorAll("input, select"));
    var changes = serializeSettingsChanges(controls);
    if (!Object.keys(changes).length) {
        settingsFeedback = { state: "idle", message: "No settings changed." };
        renderSettingsFeedback();
        return;
    }
    settingsFeedback = { state: "progress", message: "Saving settings…" };
    renderSettingsFeedback();
    var server = new Server("saveSettings");
    server.request({ settings: changes });
}
