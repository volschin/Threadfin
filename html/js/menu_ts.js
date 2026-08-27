"use strict";
class MainMenuItem {
    constructor(menuKey, value) {
        this.menuKey = menuKey;
        this.value = value;
    }
}
class ShowContent {
    constructor(menuID) {
        this.menuID = menuID;
    }
    show() {
        COLUMN_TO_SORT = -1;
        var doc = document.getElementById("content");
        if (!doc) {
            return;
        }
        doc.innerHTML = "";
        showPreview(false);
        var menuKey = menuItems[this.menuID].menuKey;
        switch (menuKey) {
            case "playlist":
            case "xmltv":
                renderSourceManagementPage(menuKey, doc);
                break;
            case "filter":
                renderFilterManagementPage(doc);
                break;
            case "mapping":
                renderMappingPage(doc);
                break;
            case "settings":
                renderSettingsPage(doc);
                break;
            case "users":
                renderUsersPage(doc);
                break;
            case "log":
                renderLogPage(doc);
                break;
            case "logout":
                document.cookie = "Token= ; expires = Thu, 01 Jan 1970 00:00:00 GMT";
                location.reload();
                return;
            default:
                console.warn("Ignoring unknown legacy menu destination", menuKey);
                return;
        }
        showElement("loading", false);
    }
}
function PageReady() {
    var server = new Server("getServerConfig");
    server.request(new Object());
    setInterval(function () {
        updateLog();
    }, 10000);
    return;
}
function createLayout() {
    // Client Info
    var obj = SERVER["clientInfo"];
    var keys = getObjKeys(obj);
    console.log("KEYS: ", keys);
    for (var i = 0; i < keys.length; i++) {
        if (document.getElementById(keys[i])) {
            document.getElementById(keys[i]).value = obj[keys[i]];
        }
    }
    if (document.getElementById("playlist-connection-information")) {
        let activeClass = "text-primary";
        if (SERVER["clientInfo"]["activePlaylist"] / SERVER["clientInfo"]["totalPlaylist"] >= 0.6 && SERVER["clientInfo"]["activePlaylist"] / SERVER["clientInfo"]["totalPlaylist"] < 0.8) {
            activeClass = "text-warning";
        }
        else if (SERVER["clientInfo"]["activePlaylist"] / SERVER["clientInfo"]["totalPlaylist"] >= 0.8) {
            activeClass = "text-danger";
        }
        document.getElementById("playlist-connection-information").innerHTML = "Playlist Connections: <span class='" + activeClass + "'>" + SERVER["clientInfo"]["activePlaylist"] + " / " + SERVER["clientInfo"]["totalPlaylist"] + "</span>";
    }
    if (document.getElementById("client-connection-information")) {
        let activeClass = "text-primary";
        if (SERVER["clientInfo"]["activeClients"] / SERVER["clientInfo"]["totalClients"] >= 0.6 && SERVER["clientInfo"]["activeClients"] / SERVER["clientInfo"]["totalClients"] < 0.8) {
            activeClass = "text-warning";
        }
        else if (SERVER["clientInfo"]["activeClients"] / SERVER["clientInfo"]["totalClients"] >= 0.8) {
            activeClass = "text-danger";
        }
        document.getElementById("client-connection-information").innerHTML = "Client Connections: <span class='" + activeClass + "'>" + SERVER["clientInfo"]["activeClients"] + " / " + SERVER["clientInfo"]["totalClients"] + "</span>";
    }
    if (!document.getElementById("main-menu")) {
        return;
    }
    renderNavigation();
    restoreInitialDestinationFromHistory();
    return;
}
class PopupWindow {
    constructor() {
        this.DocumentID = "popup-custom";
        this.InteractionID = "interaction";
        this.doc = document.getElementById(this.DocumentID);
    }
    createTitle(title) {
        var td = document.createElement("TD");
        td.className = "left";
        td.innerHTML = title + ":";
        return td;
    }
    createContent(element) {
        var td = document.createElement("TD");
        td.appendChild(element);
        return td;
    }
    createInteraction() {
        var div = document.createElement("div");
        div.setAttribute("id", "popup-interaction");
        div.className = "interaction";
        this.doc.appendChild(div);
    }
}
class PopupContent extends PopupWindow {
    constructor() {
        super(...arguments);
        this.table = document.createElement("TABLE");
        this.rowIndex = 0;
    }
    createHeadline(headline) {
        this.doc.innerHTML = "";
        var element = document.createElement("H3");
        element.innerHTML = headline.toUpperCase();
        this.doc.appendChild(element);
        // Tabelle erstellen
        this.table = document.createElement("TABLE");
        this.doc.appendChild(this.table);
    }
    appendRow(title, element) {
        var tr = document.createElement("TR");
        var titleCell = null;
        var rowIndex = this.rowIndex++;
        // Bezeichnung
        if (title.length != 0) {
            titleCell = this.createTitle(title);
            titleCell.id = "popup-field-label-" + rowIndex;
            tr.appendChild(titleCell);
        }
        // Content
        tr.appendChild(this.createContent(element));
        if (titleCell) {
            var controls = [];
            var tagName = element && element.tagName ? String(element.tagName).toUpperCase() : "";
            if (tagName == "INPUT" || tagName == "SELECT" || tagName == "TEXTAREA") {
                controls.push(element);
            }
            if (element && typeof element.querySelectorAll == "function") {
                var descendants = element.querySelectorAll("input, select, textarea");
                for (var controlIndex = 0; controlIndex < descendants.length; controlIndex++) {
                    if (controls.indexOf(descendants[controlIndex]) == -1) {
                        controls.push(descendants[controlIndex]);
                    }
                }
            }
            controls.forEach((control, controlIndex) => {
                if (String(control.type || "").toLowerCase() == "hidden") {
                    return;
                }
                if (!control.id) {
                    control.id = "popup-field-" + rowIndex + "-" + controlIndex;
                }
                if (!control.getAttribute("aria-label") && !control.getAttribute("aria-labelledby")) {
                    control.setAttribute("aria-labelledby", titleCell.id);
                }
            });
        }
        this.table.appendChild(tr);
    }
    createInput(type, name, value) {
        var input = document.createElement("INPUT");
        if (value == undefined) {
            value = "";
        }
        input.setAttribute("type", type);
        input.setAttribute("name", name);
        input.setAttribute("value", value);
        return input;
    }
    createCheckbox(name) {
        var input = document.createElement("INPUT");
        input.setAttribute("type", "checkbox");
        input.setAttribute("name", name);
        return input;
    }
    createSelect(text, values, set, dbKey) {
        var select = document.createElement("SELECT");
        select.setAttribute("name", dbKey);
        for (let i = 0; i < text.length; i++) {
            var option = document.createElement("OPTION");
            option.setAttribute("value", values[i]);
            option.innerText = text[i];
            select.appendChild(option);
        }
        if (set != "") {
            select.value = set;
        }
        if (set == undefined) {
            select.value = values[0];
        }
        return select;
    }
    selectOption(select, value) {
        //select.selectedOptions = value
        var s = select;
        s.options[s.selectedIndex].value = value;
        return select;
    }
    description(value) {
        var tr = document.createElement("TR");
        var td = document.createElement("TD");
        var span = document.createElement("PRE");
        span.innerHTML = value;
        tr.appendChild(td);
        tr.appendChild(this.createContent(span));
        this.table.appendChild(tr);
    }
    // Interaktion
    addInteraction(element) {
        var interaction = document.getElementById("popup-interaction");
        interaction.appendChild(element);
    }
}
function openPopUp(dataType, element) {
    var data = new Object();
    var id;
    switch (element) {
        case undefined:
            switch (dataType) {
                case "group-title":
                    if (id == undefined) {
                        id = -1;
                    }
                    data = getLocalData("filter", id);
                    data["type"] = "group-title";
                    break;
                case "custom-filter":
                    if (id == undefined) {
                        id = -1;
                    }
                    data = getLocalData("filter", id);
                    data["type"] = "custom-filter";
                    break;
                default:
                    data["id.provider"] = "-";
                    data["type"] = dataType;
                    id = "-";
                    break;
            }
            break;
        default:
            id = element.id;
            data = getLocalData(dataType, id);
            break;
    }
    var content = new PopupContent();
    switch (dataType) {
        case "playlist":
            content.createHeadline("{{.playlist.playlistType.title}}");
            // Type
            var text = ["M3U", "HDHomeRun"];
            var values = ["javascript: openPopUp('m3u')", "javascript: openPopUp('hdhr')"];
            var select = content.createSelect(text, values, "", "type");
            select.setAttribute("id", "type");
            select.setAttribute("onchange", 'javascript: changeButtonAction(this, "next", "onclick")'); // changeButtonAction
            content.appendRow("{{.playlist.type.title}}", select);
            // Interaktion
            content.createInteraction();
            // Abbrechen
            var input = content.createInput("button", "cancel", "{{.button.cancel}}");
            input.setAttribute("onclick", 'javascript: showElement("popup", false);');
            content.addInteraction(input);
            // Weiter
            var input = content.createInput("button", "next", "{{.button.next}}");
            input.setAttribute("onclick", 'javascript: openPopUp("m3u")');
            input.setAttribute("id", 'next');
            content.addInteraction(input);
            break;
        case "m3u":
            content.createHeadline(dataType);
            // Name
            var dbKey = "name";
            var input = content.createInput("text", dbKey, data[dbKey]);
            input.setAttribute("placeholder", "{{.playlist.name.placeholder}}");
            content.appendRow("{{.playlist.name.title}}", input);
            // Beschreibung
            var dbKey = "description";
            var input = content.createInput("text", dbKey, data[dbKey]);
            input.setAttribute("placeholder", "{{.playlist.description.placeholder}}");
            content.appendRow("{{.playlist.description.title}}", input);
            // URL
            var dbKey = "file.source";
            var input = content.createInput("text", dbKey, data[dbKey]);
            input.setAttribute("placeholder", "{{.playlist.fileM3U.placeholder}}");
            content.appendRow("{{.playlist.fileM3U.title}}", input);
            var text = ["-", "FFmpeg", "VLC"];
            var values = ["-", "ffmpeg", "vlc"];
            var selected = SERVER["settings"]["buffer"];
            if (data["buffer"] != undefined) {
                selected = data["buffer"];
            }
            var select = content.createSelect(text, values, selected, "buffer");
            select.setAttribute("id", "buffer");
            content.appendRow("{{.playlist.buffer.title}}", select);
            // Tuner
            var text = new Array();
            var values = new Array();
            for (var i = 1; i <= 100; i++) {
                text.push(i.toString());
                values.push(i.toString());
            }
            var dbKey = "tuner";
            var select = content.createSelect(text, values, data[dbKey], dbKey);
            select.setAttribute("onfocus", "javascript: return;");
            content.appendRow("{{.playlist.tuner.title}}", select);
            content.description("{{.playlist.tuner.description}}");
            var dbKey = "http_proxy.ip";
            var input = content.createInput("text", dbKey, data[dbKey]);
            input.setAttribute("placeholder", "{{.playlist.http_proxy_ip.placeholder}}");
            content.appendRow("{{.playlist.http_proxy_ip.title}}", input);
            content.description("{{.playlist.http_proxy_ip.description}}");
            var dbKey = "http_proxy.port";
            var input = content.createInput("text", dbKey, data[dbKey]);
            input.setAttribute("placeholder", "{{.playlist.http_proxy_port.placeholder}}");
            content.appendRow("{{.playlist.http_proxy_port.title}}", input);
            content.description("{{.playlist.http_proxy_port.description}}");
            var dbKey = "http_headers.origin";
            var input = content.createInput("text", dbKey, data[dbKey]);
            input.setAttribute("placeholder", "{{.playlist.http_user_origin.placeholder}}");
            content.appendRow("{{.playlist.http_user_origin.title}}", input);
            content.description("{{.playlist.http_user_origin.description}}");
            var dbKey = "http_headers.referer";
            var input = content.createInput("text", dbKey, data[dbKey]);
            input.setAttribute("placeholder", "{{.playlist.http_user_referer.placeholder}}");
            content.appendRow("{{.playlist.http_user_referer.title}}", input);
            content.description("{{.playlist.http_user_referer.description}}");
            // Interaktion
            content.createInteraction();
            // Löschen
            if (data["id.provider"] != "-") {
                var input = content.createInput("button", "delete", "{{.button.delete}}");
                input.className = "delete";
                input.setAttribute('onclick', 'javascript: savePopupData("m3u", "' + id + '", true, 0)');
                content.addInteraction(input);
            }
            else {
                var input = content.createInput("button", "back", "{{.button.back}}");
                input.setAttribute("onclick", 'javascript: openPopUp("playlist")');
                content.addInteraction(input);
            }
            // Abbrechen
            var input = content.createInput("button", "cancel", "{{.button.cancel}}");
            input.setAttribute("onclick", 'javascript: showElement("popup", false);');
            content.addInteraction(input);
            // Aktualisieren
            if (data["id.provider"] != "-") {
                var input = content.createInput("button", "update", "{{.button.update}}");
                input.setAttribute('onclick', 'javascript: savePopupData("m3u", "' + id + '", false, 1)');
                content.addInteraction(input);
            }
            // Speichern
            var input = content.createInput("button", "save", "{{.button.save}}");
            input.setAttribute('onclick', 'javascript: savePopupData("m3u", "' + id + '", false, 0)');
            content.addInteraction(input);
            break;
        case "hdhr":
            content.createHeadline(dataType);
            // Name
            var dbKey = "name";
            var input = content.createInput("text", dbKey, data[dbKey]);
            input.setAttribute("placeholder", "{{.playlist.name.placeholder}}");
            content.appendRow("{{.playlist.name.title}}", input);
            // Beschreibung
            var dbKey = "description";
            var input = content.createInput("text", dbKey, data[dbKey]);
            input.setAttribute("placeholder", "{{.playlist.description.placeholder}}");
            content.appendRow("{{.playlist.description.placeholder}}", input);
            // URL
            var dbKey = "file.source";
            var input = content.createInput("text", dbKey, data[dbKey]);
            input.setAttribute("placeholder", "{{.playlist.fileHDHR.placeholder}}");
            content.appendRow("{{.playlist.fileHDHR.title}}", input);
            var text = ["-", "FFmpeg", "VLC"];
            var values = ["-", "ffmpeg", "vlc"];
            var selected = SERVER["settings"]["buffer"];
            if (data["buffer"] != undefined) {
                selected = data["buffer"];
            }
            var select = content.createSelect(text, values, selected, "buffer");
            select.setAttribute("id", "buffer");
            content.appendRow("{{.playlist.buffer.title}}", select);
            // Tuner
            var text = new Array();
            var values = new Array();
            for (var i = 1; i <= 100; i++) {
                text.push(i.toString());
                values.push(i.toString());
            }
            var dbKey = "tuner";
            var select = content.createSelect(text, values, data[dbKey], dbKey);
            select.setAttribute("onfocus", "javascript: return;");
            content.appendRow("{{.playlist.tuner.title}}", select);
            content.description("{{.playlist.tuner.description}}");
            var dbKey = "http_proxy.ip";
            var input = content.createInput("text", dbKey, data[dbKey]);
            input.setAttribute("placeholder", "{{.playlist.http_proxy_ip.placeholder}}");
            content.appendRow("{{.playlist.http_proxy_ip.title}}", input);
            content.description("{{.playlist.http_proxy_ip.description}}");
            var dbKey = "http_proxy.port";
            var input = content.createInput("text", dbKey, data[dbKey]);
            input.setAttribute("placeholder", "{{.playlist.http_proxy_port.placeholder}}");
            content.appendRow("{{.playlist.http_proxy_port.title}}", input);
            content.description("{{.playlist.http_proxy_port.description}}");
            // Interaktion
            content.createInteraction();
            // Löschen
            if (data["id.provider"] != "-") {
                var input = content.createInput("button", "delete", "{{.button.delete}}");
                input.setAttribute('onclick', 'javascript: savePopupData("hdhr", "' + id + '", true, 0)');
                input.className = "delete";
                content.addInteraction(input);
            }
            else {
                var input = content.createInput("button", "back", "{{.button.back}}");
                input.setAttribute("onclick", 'javascript: openPopUp("playlist")');
                content.addInteraction(input);
            }
            // Abbrechen
            var input = content.createInput("button", "cancel", "{{.button.cancel}}");
            input.setAttribute("onclick", 'javascript: showElement("popup", false);');
            content.addInteraction(input);
            // Aktualisieren
            if (data["id.provider"] != "-") {
                var input = content.createInput("button", "update", "{{.button.update}}");
                input.setAttribute('onclick', 'javascript: savePopupData("hdhr", "' + id + '", false, 1)');
                content.addInteraction(input);
            }
            // Speichern
            var input = content.createInput("button", "save", "{{.button.save}}");
            input.setAttribute('onclick', 'javascript: savePopupData("hdhr", "' + id + '", false, 0)');
            content.addInteraction(input);
            break;
        case "filter":
            content.createHeadline(dataType);
            // Type
            var dbKey = "type";
            var text = ["M3U: " + "{{.filter.type.groupTitle}}", "Threadfin: " + "{{.filter.type.customFilter}}"];
            var values = ["javascript: openPopUp('group-title')", "javascript: openPopUp('custom-filter')"];
            var select = content.createSelect(text, values, "javascript: openPopUp('group-title')", dbKey);
            select.setAttribute("id", id);
            select.setAttribute("onchange", 'javascript: changeButtonAction(this, "next", "onclick");'); // changeButtonAction
            content.appendRow("{{.filter.type.title}}", select);
            // Interaktion
            content.createInteraction();
            // Abbrechen
            var input = content.createInput("button", "cancel", "{{.button.cancel}}");
            input.setAttribute("onclick", 'javascript: showElement("popup", false);');
            content.addInteraction(input);
            // Weiter
            var input = content.createInput("button", "next", "{{.button.next}}");
            input.setAttribute("onclick", 'javascript: openPopUp("group-title")');
            input.setAttribute("id", 'next');
            content.addInteraction(input);
            break;
        case "custom-filter":
        case "group-title":
            switch (dataType) {
                case "custom-filter":
                    content.createHeadline("{{.filter.custom}}");
                    break;
                case "group-title":
                    content.createHeadline("{{.filter.group}}");
                    break;
            }
            // Name      
            var dbKey = "name";
            var input = content.createInput("text", dbKey, data[dbKey]);
            input.setAttribute("placeholder", "{{.filter.name.placeholder}}");
            content.appendRow("{{.filter.name.title}}", input);
            // Beschreibung
            var dbKey = "description";
            var input = content.createInput("text", dbKey, data[dbKey]);
            input.setAttribute("placeholder", "{{.filter.description.placeholder}}");
            content.appendRow("{{.filter.description.title}}", input);
            // Typ
            var dbKey = "type";
            var input = content.createInput("hidden", dbKey, data[dbKey]);
            content.appendRow("", input);
            var filterType = data[dbKey];
            switch (filterType) {
                case "custom-filter":
                    // Groß- Kleinschreibung beachten
                    var dbKey = "caseSensitive";
                    var input = content.createCheckbox(dbKey);
                    input.checked = data[dbKey];
                    content.appendRow("{{.filter.caseSensitive.title}}", input);
                    // Filterregel (Benutzerdefiniert)
                    var dbKey = "filter";
                    var input = content.createInput("text", dbKey, data[dbKey]);
                    input.setAttribute("placeholder", "{{.filter.filterRule.placeholder}}");
                    content.appendRow("{{.filter.filterRule.title}}", input);
                    break;
                case "group-title":
                    //alert(dbKey + " " + filterType)
                    // Filter basierend auf den Gruppen in der M3U
                    var dbKey = "filter";
                    var groupsM3U = getLocalData("m3uGroups", "");
                    var text = groupsM3U["text"];
                    var values = groupsM3U["value"];
                    var select = content.createSelect(text, values, data[dbKey], dbKey);
                    select.setAttribute("onchange", "javascript: this.className = 'changed'");
                    content.appendRow("{{.filter.filterGroup.title}}", select);
                    content.description("{{.filter.filterGroup.description}}");
                    var dbKey = "liveEvent";
                    var input = content.createCheckbox(dbKey);
                    input.checked = data[dbKey];
                    content.appendRow("{{.filter.liveEvent.title}}", input);
                    // Groß- Kleinschreibung beachten
                    var dbKey = "caseSensitive";
                    var input = content.createCheckbox(dbKey);
                    input.checked = data[dbKey];
                    content.appendRow("{{.filter.caseSensitive.title}}", input);
                    var dbKey = "include";
                    var input = content.createInput("text", dbKey, data[dbKey]);
                    input.setAttribute("placeholder", "{{.filter.include.placeholder}}");
                    content.appendRow("{{.filter.include.title}}", input);
                    content.description("{{.filter.include.description}}");
                    var dbKey = "exclude";
                    var input = content.createInput("text", dbKey, data[dbKey]);
                    input.setAttribute("placeholder", "{{.filter.exclude.placeholder}}");
                    content.appendRow("{{.filter.exclude.title}}", input);
                    content.description("{{.filter.exclude.description}}");
                    break;
                default:
                    break;
            }
            // Name      
            var dbKey = "startingNumber";
            if (data[dbKey] !== undefined) {
                var input = content.createInput("text", dbKey, data[dbKey]);
            }
            else {
                var input = content.createInput("text", dbKey, "1000");
            }
            input.setAttribute("placeholder", "{{.filter.startingnumber.placeholder}}");
            content.appendRow("{{.filter.startingnumber.title}}", input);
            content.description("{{.filter.startingnumber.description}}");
            var dbKey = "x-category";
            var text = ["-"];
            var values = [""];
            var epgCategories = SERVER["settings"]["epgCategories"];
            var categories = epgCategories.split("|");
            for (i = 0; i <= categories.length; i++) {
                var cat = categories[i];
                if (cat) {
                    var cat_split = cat.split(":");
                    text.push(cat_split[0]);
                    values.push(cat_split[1]);
                }
            }
            var select = content.createSelect(text, values, data[dbKey], dbKey);
            select.setAttribute("onchange", "javascript: this.className = 'changed'");
            content.appendRow("{{.filter.category.title}}", select);
            // Interaktion
            content.createInteraction();
            // Löschen
            var input = content.createInput("button", "delete", "{{.button.delete}}");
            input.setAttribute('onclick', 'javascript: savePopupData("filter", "' + id + '", true, 0)');
            input.className = "delete";
            content.addInteraction(input);
            // Abbrechen
            var input = content.createInput("button", "cancel", "{{.button.cancel}}");
            input.setAttribute("onclick", 'javascript: showElement("popup", false);');
            content.addInteraction(input);
            // Speichern
            var input = content.createInput("button", "save", "{{.button.save}}");
            input.setAttribute('onclick', 'javascript: savePopupData("filter", "' + id + '", false, 0)');
            content.addInteraction(input);
            break;
        case "xmltv":
            content.createHeadline(dataType);
            // Name
            var dbKey = "name";
            var input = content.createInput("text", dbKey, data[dbKey]);
            input.setAttribute("placeholder", "{{.xmltv.name.placeholder}}");
            content.appendRow("{{.xmltv.name.title}}", input);
            // Beschreibung
            var dbKey = "description";
            var input = content.createInput("text", dbKey, data[dbKey]);
            input.setAttribute("placeholder", "{{.xmltv.description.placeholder}}");
            content.appendRow("{{.xmltv.description.title}}", input);
            // URL
            var dbKey = "file.source";
            var input = content.createInput("text", dbKey, data[dbKey]);
            input.setAttribute("placeholder", "{{.xmltv.fileXMLTV.placeholder}}");
            content.appendRow("{{.xmltv.fileXMLTV.title}}", input);
            var dbKey = "http_proxy.ip";
            var input = content.createInput("text", dbKey, data[dbKey]);
            input.setAttribute("placeholder", "{{.xmltv.http_proxy_ip.placeholder}}");
            content.appendRow("{{.xmltv.http_proxy_ip.title}}", input);
            content.description("{{.xmltv.http_proxy_ip.description}}");
            var dbKey = "http_proxy.port";
            var input = content.createInput("text", dbKey, data[dbKey]);
            input.setAttribute("placeholder", "{{.xmltv.http_proxy_port.placeholder}}");
            content.appendRow("{{.xmltv.http_proxy_port.title}}", input);
            content.description("{{.xmltv.http_proxy_port.description}}");
            // Interaktion
            content.createInteraction();
            // Löschen
            if (data["id.provider"] != "-") {
                var input = content.createInput("button", "delete", "{{.button.delete}}");
                input.setAttribute('onclick', 'javascript: savePopupData("xmltv", "' + id + '", true, 0)');
                input.className = "delete";
                content.addInteraction(input);
            }
            // Abbrechen
            var input = content.createInput("button", "cancel", "{{.button.cancel}}");
            input.setAttribute("onclick", 'javascript: showElement("popup", false);');
            content.addInteraction(input);
            // Aktualisieren
            if (data["id.provider"] != "-") {
                var input = content.createInput("button", "update", "{{.button.update}}");
                input.setAttribute('onclick', 'javascript: savePopupData("xmltv", "' + id + '", false, 1)');
                content.addInteraction(input);
            }
            // Speichern
            var input = content.createInput("button", "save", "{{.button.save}}");
            input.setAttribute('onclick', 'javascript: savePopupData("xmltv", "' + id + '", false, 0)');
            content.addInteraction(input);
            break;
        case "users":
            content.createHeadline("{{.mainMenu.item.users}}");
            // Benutzername 
            var dbKey = "username";
            var input = content.createInput("text", dbKey, data[dbKey]);
            input.setAttribute("placeholder", "{{.users.username.placeholder}}");
            content.appendRow("{{.users.username.title}}", input);
            // Neues Passwort 
            var dbKey = "password";
            var input = content.createInput("password", dbKey, "");
            input.setAttribute("placeholder", "{{.users.password.placeholder}}");
            content.appendRow("{{.users.password.title}}", input);
            // Bestätigung 
            var dbKey = "confirm";
            var input = content.createInput("password", dbKey, "");
            input.setAttribute("placeholder", "{{.users.confirm.placeholder}}");
            content.appendRow("{{.users.confirm.title}}", input);
            // Berechtigung WEB
            var dbKey = "authentication.web";
            var input = content.createCheckbox(dbKey);
            input.checked = data[dbKey];
            if (data["defaultUser"] == true) {
                input.setAttribute("onclick", "javascript: return false");
            }
            content.appendRow("{{.users.web.title}}", input);
            // Berechtigung PMS
            var dbKey = "authentication.pms";
            var input = content.createCheckbox(dbKey);
            input.checked = data[dbKey];
            content.appendRow("{{.users.pms.title}}", input);
            // Berechtigung M3U
            var dbKey = "authentication.m3u";
            var input = content.createCheckbox(dbKey);
            input.checked = data[dbKey];
            content.appendRow("{{.users.m3u.title}}", input);
            // Berechtigung XML
            var dbKey = "authentication.xml";
            var input = content.createCheckbox(dbKey);
            input.checked = data[dbKey];
            content.appendRow("{{.users.xml.title}}", input);
            // Berechtigung API
            var dbKey = "authentication.api";
            var input = content.createCheckbox(dbKey);
            input.checked = data[dbKey];
            content.appendRow("{{.users.api.title}}", input);
            // Berechtigung CONFIG
            var dbKey = "authentication.config";
            var input = content.createCheckbox(dbKey);
            input.checked = data[dbKey] == true;
            content.appendRow("{{.users.config.title}}", input);
            // Interaktion
            content.createInteraction();
            // Löschen
            if (data["defaultUser"] != true && id != "-") {
                var input = content.createInput("button", "delete", "{{.button.delete}}");
                input.className = "delete";
                input.setAttribute('onclick', 'javascript: savePopupData("' + dataType + '", "' + id + '", true, 0)');
                content.addInteraction(input);
            }
            // Abbrechen
            var input = content.createInput("button", "cancel", "{{.button.cancel}}");
            input.setAttribute("onclick", 'javascript: showElement("popup", false);');
            content.addInteraction(input);
            // Speichern
            var input = content.createInput("button", "save", "{{.button.save}}");
            input.setAttribute("onclick", 'javascript: savePopupData("' + dataType + '", "' + id + '", "false");');
            content.addInteraction(input);
            break;
        case "mapping":
            content.createHeadline("{{.mainMenu.item.mapping}}");
            if (BULK_EDIT == true) {
                var dbKey = "x-channels-start";
                var input = content.createInput("text", dbKey, data[dbKey]);
                // Set the value to the first selected channel
                var channels = getAllSelectedChannels();
                var channel = SERVER["xepg"]["epgMapping"][channels[0]];
                if (typeof channel !== "undefined") {
                    input.setAttribute("value", channel["x-channelID"]);
                }
                input.setAttribute("onchange", 'javascript: changeChannelNumbers("' + channels + '");');
                content.appendRow("{{.mapping.channelGroupStart.title}}", input);
            }
            // Aktiv 
            var dbKey = "x-active";
            var input = content.createCheckbox(dbKey);
            input.checked = data[dbKey];
            input.id = "active";
            //input.setAttribute("onchange", "javascript: this.className = 'changed'")
            input.setAttribute("onchange", "javascript: toggleChannelStatus('" + id + "', this)");
            content.appendRow("{{.mapping.active.title}}", input);
            // Kanalname 
            var dbKey = "x-name";
            var input = content.createInput("text", dbKey, data[dbKey]);
            input.setAttribute("onchange", "javascript: this.className = 'changed'");
            if (BULK_EDIT == true) {
                input.style.border = "solid 1px red";
                input.setAttribute("readonly", "true");
            }
            content.appendRow("{{.mapping.channelName.title}}", input);
            content.description("<span class='text-danger'>" + data["tvg-id"] + "</span> <span class='text-primary'>(" + data["x-epg"] + ")</span>");
            // Beschreibung 
            var dbKey = "x-description";
            var input = content.createInput("text", dbKey, data[dbKey]);
            input.setAttribute("placeholder", "{{.mapping.description.placeholder}}");
            input.setAttribute("onchange", "javascript: this.className = 'changed'");
            content.appendRow("{{.mapping.description.title}}", input);
            // Aktualisierung des Kanalnamens
            if (data.hasOwnProperty("_uuid.key")) {
                if (data["_uuid.key"] != "") {
                    var dbKey = "x-update-channel-name";
                    var input = content.createCheckbox(dbKey);
                    input.setAttribute("onchange", "javascript: this.className = 'changed'");
                    input.checked = data[dbKey];
                    content.appendRow("{{.mapping.updateChannelName.title}}", input);
                }
            }
            // Logo URL (Kanal) 
            var dbKey = "tvg-logo";
            var input = content.createInput("text", dbKey, data[dbKey]);
            input.setAttribute("onchange", "javascript: this.className = 'changed'");
            input.setAttribute("id", "channel-icon");
            content.appendRow("{{.mapping.channelLogo.title}}", input);
            // Aktualisierung des Kanallogos
            var dbKey = "x-update-channel-icon";
            var input = content.createCheckbox(dbKey);
            input.checked = data[dbKey];
            input.setAttribute("id", "update-icon");
            input.setAttribute("onchange", "javascript: this.className = 'changed'; changeChannelLogo('" + id + "');");
            content.appendRow("{{.mapping.updateChannelLogo.title}}", input);
            // Erweitern der EPG Kategorie
            var dbKey = "x-category";
            var text = ["-"];
            var values = [""];
            var epgCategories = SERVER["settings"]["epgCategories"];
            var categories = epgCategories.split("|");
            for (i = 0; i <= categories.length; i++) {
                var cat = categories[i];
                if (cat) {
                    var cat_split = cat.split(":");
                    text.push(cat_split[0]);
                    values.push(cat_split[1]);
                }
            }
            var select = content.createSelect(text, values, data[dbKey], dbKey);
            select.setAttribute("onchange", "javascript: this.className = 'changed'");
            content.appendRow("{{.mapping.epgCategory.title}}", select);
            // M3U Gruppentitel
            var dbKey = "x-group-title";
            var input = content.createInput("text", dbKey, data[dbKey]);
            input.setAttribute("onchange", "javascript: this.className = 'changed'");
            content.appendRow("{{.mapping.m3uGroupTitle.title}}", input);
            if (data["group-title"] != undefined) {
                content.description(data["group-title"]);
            }
            // XMLTV Datei
            var dbKey = "x-xmltv-file";
            var xmlFile = data[dbKey];
            var xmltv = new XMLTVFile();
            var select = xmltv.getFiles(data[dbKey]);
            select.setAttribute("name", dbKey);
            select.setAttribute("id", "popup-xmltv");
            select.setAttribute("onchange", "javascript: this.className = 'changed'; setXmltvChannel('" + id + "',this, '" + data["x-mapping"] + "');");
            content.appendRow("{{.mapping.xmltvFile.title}}", select);
            var file = data[dbKey];
            // XMLTV Mapping
            var dbKey = "x-mapping";
            var xmltv = new XMLTVFile();
            const currentXmlTvId = data[dbKey];
            const [xmlTvIdContainer, xmlTvIdInput, xmlTvIdDatalist] = xmltv.newXmlTvIdPicker(xmlFile, currentXmlTvId);
            xmlTvIdContainer.setAttribute('id', 'xmltv-id-picker-container');
            xmlTvIdInput.setAttribute('list', 'xmltv-id-picker-datalist');
            xmlTvIdInput.setAttribute('name', 'x-mapping'); // Should stay x-mapping as it will be used in donePopupData to make a server request
            xmlTvIdInput.setAttribute('id', 'xmltv-id-picker-input');
            xmlTvIdInput.setAttribute('onchange', `javascript: this.className = 'changed'; checkXmltvChannel('${id}', this, '${xmlFile}');`);
            xmlTvIdDatalist.setAttribute('id', 'xmltv-id-picker-datalist');
            // sortSelect(xmlTvIdDatalist); // TODO: Better sort before adding
            content.appendRow('{{.mapping.xmltvChannel.title}}', xmlTvIdContainer);
            // Extra PPV Data
            if (currentXmlTvId == "PPV") {
                var dbKey = "x-ppv-extra";
                var input = content.createInput("text", dbKey, data[dbKey]);
                input.setAttribute("onchange", "javascript: this.className = 'changed'");
                input.setAttribute("id", "ppv-extra");
                content.appendRow("{{.mapping.ppvextra.title}}", input);
            }
            var dbKey = "x-backup-channel-1";
            var xmltv = new XMLTVFile();
            const backup1XmlTvId = data[dbKey];
            const [xmlTvBackup1IdContainer, xmlTvBackup1IdInput, xmlTvBackup1IdDatalist] = xmltv.newM3uPicker(xmlFile, backup1XmlTvId);
            xmlTvBackup1IdContainer.setAttribute('id', 'm3u-id-picker-container-1');
            xmlTvBackup1IdInput.setAttribute('list', 'm3u-id-picker-datalist');
            xmlTvBackup1IdInput.setAttribute('name', dbKey); // Should stay x-mapping as it will be used in donePopupData to make a server request
            xmlTvBackup1IdInput.setAttribute("id", "backup-channel-1");
            xmlTvBackup1IdInput.setAttribute('onchange', `javascript: this.className = 'changed'; checkXmltvChannel('${id}', this, '${xmlFile}');`);
            xmlTvBackup1IdDatalist.setAttribute('id', 'm3u-id-picker-datalist');
            // sortSelect(xmlTvIdDatalist); // TODO: Better sort before adding
            content.appendRow('{{.mapping.backupChannel1.title}}', xmlTvBackup1IdContainer);
            var dbKey = "x-backup-channel-2";
            var xmltv = new XMLTVFile();
            const backup2XmlTvId = data[dbKey];
            const [xmlTvBackup2IdContainer, xmlTvBackup2IdInput, xmlTvBackup2IdDatalist] = xmltv.newM3uPicker(xmlFile, backup2XmlTvId);
            xmlTvBackup2IdContainer.setAttribute('id', 'xmltv-id-picker-container-2');
            xmlTvBackup2IdInput.setAttribute('list', 'm3u-id-picker-datalist');
            xmlTvBackup2IdInput.setAttribute('name', dbKey); // Should stay x-mapping as it will be used in donePopupData to make a server request
            xmlTvBackup2IdInput.setAttribute("id", "backup-channel-2");
            xmlTvBackup2IdInput.setAttribute('onchange', `javascript: this.className = 'changed'; checkXmltvChannel('${id}', this, '${xmlFile}');`);
            xmlTvBackup2IdDatalist.setAttribute('id', 'm3u-id-picker-datalist');
            content.appendRow("{{.mapping.backupChannel2.title}}", xmlTvBackup2IdContainer);
            var dbKey = "x-backup-channel-3";
            var xmltv = new XMLTVFile();
            const backup3XmlTvId = data[dbKey];
            const [xmlTvBackup3IdContainer, xmlTvBackup3IdInput, xmlTvBackup3IdDatalist] = xmltv.newM3uPicker(xmlFile, backup3XmlTvId);
            xmlTvBackup3IdContainer.setAttribute('id', 'xmltv-id-picker-container-3');
            xmlTvBackup3IdInput.setAttribute('list', 'm3u-id-picker-datalist');
            xmlTvBackup3IdInput.setAttribute('name', dbKey); // Should stay x-mapping as it will be used in donePopupData to make a server request
            xmlTvBackup3IdInput.setAttribute("id", "backup-channel-3");
            xmlTvBackup3IdInput.setAttribute('onchange', `javascript: this.className = 'changed'; checkXmltvChannel('${id}', this, '${xmlFile}');`);
            xmlTvBackup3IdDatalist.setAttribute('id', 'm3u-id-picker-datalist');
            content.appendRow("{{.mapping.backupChannel3.title}}", xmlTvBackup3IdContainer);
            // Interaktion
            content.createInteraction();
            var input = content.createInput("button", "cancel", "{{.button.probeChannel}}");
            input.setAttribute("onclick", 'javascript: probeChannel("' + data["url"] + '");');
            content.addInteraction(input);
            // Logo hochladen
            var input = content.createInput("button", "cancel", "{{.button.uploadLogo}}");
            input.setAttribute("onclick", 'javascript: uploadLogo();');
            content.addInteraction(input);
            // Abbrechen
            var input = content.createInput("button", "cancel", "{{.button.cancel}}");
            input.setAttribute("onclick", 'javascript: showElement("popup", false);');
            content.addInteraction(input);
            // Fertig
            var ids = new Array();
            ids = getAllSelectedChannels();
            if (ids.length == 0) {
                ids.push(id);
            }
            var input = content.createInput("button", "save", "{{.button.done}}");
            input.setAttribute("onclick", 'javascript: donePopupData("' + dataType + '", "' + ids + '", "false");');
            content.addInteraction(input);
            var probeDetails = document.createElement("div");
            probeDetails.id = "probeDetails";
            content.appendRow("{{.mapping.probeDetails.title}}", probeDetails);
            break;
        default:
            break;
    }
    enhanceSourcePopup(dataType);
    if (typeof enhanceFilterPopup == "function") {
        enhanceFilterPopup(dataType);
    }
    if (dataType == "users") {
        enhanceUsersPopup(id, data);
    }
    showPopUpElement('popup-custom');
}
class XMLTVFile {
    getFiles(set) {
        var fileIDs = getObjKeys(SERVER["xepg"]["xmltvMap"]);
        var values = new Array("-");
        var text = new Array("-");
        for (let i = 0; i < fileIDs.length; i++) {
            if (fileIDs[i] != "Threadfin Dummy") {
                values.push(getValueFromProviderFile(fileIDs[i], "xmltv", "file.threadfin"));
                text.push(getValueFromProviderFile(fileIDs[i], "xmltv", "name"));
            }
            else {
                values.push(fileIDs[i]);
                text.push(fileIDs[i]);
            }
        }
        var select = document.createElement("SELECT");
        for (let i = 0; i < text.length; i++) {
            var option = document.createElement("OPTION");
            option.setAttribute("value", values[i]);
            option.innerText = text[i];
            select.appendChild(option);
        }
        if (set != "") {
            select.value = set;
        }
        return select;
    }
    /**
   * @param xmlTvFile XML file path to get EPG from.
   * @param currentXmlTvId Current XMLTV ID to set initial input value to.
   * @returns Array of, sequentially:
   * 1) Container of the picker.
   * 2) Input field to type at and get choice from.
   * 3) Datalist containing every option.
   */
    newXmlTvIdPicker(xmlTvFile, currentXmlTvId) {
        const container = document.createElement('div');
        const input = document.createElement('input');
        input.setAttribute('type', 'text');
        // Initially, set value to '-' if input is empty
        input.value = (currentXmlTvId) ? currentXmlTvId : '-';
        // When input is focused, remove '-' from it
        input.addEventListener('focus', (evt) => {
            const target = evt.target;
            target.value = (target.value === '-') ? '' : target.value;
        });
        // When input lose focus or take a value, if it's empty, set value to '-'
        input.addEventListener('blur', setFallbackValue);
        input.addEventListener('change', setFallbackValue);
        function setFallbackValue(evt) {
            const target = evt.target;
            target.value = (target.value) ? target.value : '-';
        }
        container.appendChild(input);
        const datalist = document.createElement('datalist');
        const option = document.createElement('option');
        option.setAttribute('value', '-');
        option.innerText = '-';
        datalist.appendChild(option);
        const epg = SERVER['xepg']['xmltvMap'][xmlTvFile];
        if (epg) {
            const programIds = getOwnObjProps(epg);
            programIds.forEach((programId) => {
                const program = epg[programId];
                if (program.hasOwnProperty('display-name')) {
                    const option = document.createElement('option');
                    option.setAttribute('value', programId);
                    option.innerText = program["display-name"];
                    datalist.appendChild(option);
                }
                else {
                    const option = document.createElement('option');
                    option.setAttribute('value', programId);
                    option.innerText = '-';
                    datalist.appendChild(option);
                }
            });
        }
        container.appendChild(datalist);
        return [container, input, datalist];
    }
    newM3uPicker(xmlTvFile, currentXmlTvId) {
        const container = document.createElement('div');
        const input = document.createElement('input');
        input.setAttribute('type', 'text');
        // Initially, set value to '-' if input is empty
        input.value = (currentXmlTvId) ? currentXmlTvId : '-';
        // When input is focused, remove '-' from it
        input.addEventListener('focus', (evt) => {
            const target = evt.target;
            target.value = (target.value === '-') ? '' : target.value;
        });
        // When input lose focus or take a value, if it's empty, set value to '-'
        input.addEventListener('blur', setFallbackValue);
        input.addEventListener('change', setFallbackValue);
        function setFallbackValue(evt) {
            const target = evt.target;
            target.value = (target.value) ? target.value : '-';
        }
        container.appendChild(input);
        const datalist = document.createElement('datalist');
        const option = document.createElement('option');
        option.setAttribute('value', '-');
        option.innerText = '-';
        datalist.appendChild(option);
        const m3u = SERVER['xepg']['epgMapping'];
        if (m3u) {
            const programIds = getOwnObjProps(m3u);
            programIds.forEach((programId) => {
                const chanel = m3u[programId];
                if (chanel.hasOwnProperty('tvg-name')) {
                    const option = document.createElement('option');
                    option.setAttribute('value', chanel["tvg-name"]);
                    option.innerText = programId;
                    datalist.appendChild(option);
                }
                else {
                    const option = document.createElement('option');
                    option.setAttribute('value', chanel["tvg-name"]);
                    option.innerText = '-';
                    datalist.appendChild(option);
                }
            });
        }
        container.appendChild(datalist);
        return [container, input, datalist];
    }
    getPrograms(file, set, active) {
        //var fileIDs:string[] = getObjKeys(SERVER["xepg"]["xmltvMap"])
        var values = getObjKeys(SERVER["xepg"]["xmltvMap"][file]);
        var text = new Array();
        var displayName;
        var actives = getObjKeys(SERVER["data"]["StreamPreviewUI"]["activeStreams"]);
        var active_list = new Array();
        if (active == true) {
            for (let i = 0; i < actives.length; i++) {
                var names_split = SERVER["data"]["StreamPreviewUI"]["activeStreams"][actives[i]].split("[");
                displayName = names_split[0].trim();
                if (displayName != "") {
                    var object = { "value": displayName, "display": displayName };
                    active_list.push(object);
                }
            }
        }
        else {
            for (let i = 0; i < values.length; i++) {
                if (SERVER["xepg"]["xmltvMap"][file][values[i]].hasOwnProperty('display-name') == true) {
                    displayName = SERVER["xepg"]["xmltvMap"][file][values[i]]["display-name"];
                }
                else {
                    displayName = "-";
                }
                text[i] = displayName + " (" + values[i] + ")";
            }
        }
        text.unshift("-");
        values.unshift("-");
        var select = document.createElement("SELECT");
        for (let i = 0; i < text.length; i++) {
            var option = document.createElement("OPTION");
            option.setAttribute("value", values[i]);
            option.innerText = text[i];
            select.appendChild(option);
        }
        for (let i = 0; i < active_list.length; i++) {
            var option = document.createElement("OPTION");
            option.setAttribute("value", active_list[i]["value"]);
            option.innerText = active_list[i]["display"];
            select.appendChild(option);
        }
        if (set != "") {
            select.value = set;
        }
        if (select.value != set) {
            select.value = "-";
        }
        return select;
    }
}
function getValueFromProviderFile(file, fileType, key) {
    if (file == "Threadfin Dummy") {
        return file;
    }
    var fileID;
    var indicator = file.charAt(0);
    switch (indicator) {
        case "M":
            fileType = "m3u";
            fileID = file;
            break;
        case "H":
            fileType = "hdhr";
            fileID = file;
            break;
        case "X":
            fileType = "xmltv";
            fileID = file.substring(0, file.lastIndexOf('.'));
            break;
    }
    if (SERVER["settings"]["files"][fileType].hasOwnProperty(fileID) == true) {
        var data = SERVER["settings"]["files"][fileType][fileID];
        return data[key];
    }
    return;
}
function setXmltvChannel(epgMapId, xmlTvFileSelect) {
    const xmlTv = new XMLTVFile();
    const newXmlTvFile = xmlTvFileSelect.value;
    // Remove old XMLTV ID selection box
    const xmlTvIdPickerParent = document.getElementById('xmltv-id-picker-container').parentElement;
    xmlTvIdPickerParent.innerHTML = '';
    // Create new XMLTV ID selection box
    const tvgId = SERVER['xepg']['epgMapping'][epgMapId]['tvg-id'];
    const [xmlTvIdContainer, xmlTvIdInput, xmlTvIdDatalist] = xmlTv.newXmlTvIdPicker(newXmlTvFile, tvgId);
    xmlTvIdContainer.setAttribute('id', 'xmltv-id-picker-container');
    xmlTvIdInput.setAttribute('list', 'xmltv-id-picker-datalist');
    xmlTvIdInput.setAttribute('name', 'x-mapping'); // Should stay x-mapping as it will be used in donePopupData to make a server request
    xmlTvIdInput.setAttribute('id', 'xmltv-id-picker-input');
    xmlTvIdInput.setAttribute('onchange', `javascript: this.className = 'changed'; checkXmltvChannel('${epgMapId}', this.value, '${newXmlTvFile}');`);
    xmlTvIdInput.classList.add('changed');
    xmlTvIdDatalist.setAttribute('id', 'xmltv-id-picker-datalist');
    // Add new XMLTV ID selection box to it's parent
    xmlTvIdPickerParent.appendChild(xmlTvIdContainer);
    checkXmltvChannel(epgMapId, xmlTvIdInput.value, newXmlTvFile);
}
function checkPPV(title, element) {
    var value = element.value;
    console.log("DUMMY TYPE: " + value);
    if (value == "PPV") {
        var td = document.getElementById("x-ppv-extra").parentElement;
        td.innerHTML = "";
        var dbKey = "x-ppv-extra";
        var input = document.createElement("INPUT");
        input.setAttribute("type", "text");
        input.setAttribute("name", dbKey);
        // input.setAttribute("value", value)
        input.setAttribute("onchange", "javascript: this.className = 'changed'");
        input.setAttribute("id", "ppv-extra");
        var tr = document.createElement("TR");
        // Bezeichnung
        if (title.length != 0) {
            var td = document.createElement("TD");
            td.className = "left";
            td.innerHTML = title + ":";
        }
        // Content
        td.appendChild(element);
        this.table.appendChild(tr);
    }
}
function checkXmltvChannel(id, element, xmlFile) {
    var value = element.value;
    var bool;
    var checkbox = document.getElementById('active');
    var channel = SERVER["xepg"]["epgMapping"][id];
    var updateLogo;
    if (value == "-") {
        bool = false;
    }
    else {
        bool = true;
    }
    checkbox.checked = bool;
    checkbox.className = "changed";
    console.log(xmlFile);
    // Kanallogo aktualisieren
    /*
    updateLogo = (document.getElementById("update-icon") as HTMLInputElement).checked
    console.log(updateLogo);
    */
    if (xmlFile != "Threadfin Dummy" && bool == true) {
        //(document.getElementById("update-icon") as HTMLInputElement).checked = true;
        //(document.getElementById("update-icon") as HTMLInputElement).className = "changed";
        console.log("ID", id);
        changeChannelLogo(id);
        return;
    }
    if (xmlFile == "Threadfin Dummy") {
        document.getElementById("update-icon").checked = false;
        document.getElementById("update-icon").className = "changed";
    }
    return;
}
function changeChannelLogo(epgMapId) {
    const channel = SERVER['xepg']['epgMapping'][epgMapId];
    const xmlTvFileSelect = document.getElementById('popup-xmltv');
    const xmlTvFile = xmlTvFileSelect.options[xmlTvFileSelect.selectedIndex].value;
    const xmlTvIdInput = document.getElementById('xmltv-id-picker-input');
    const newXmlTvId = xmlTvIdInput.value;
    const updateLogo = !BULK_EDIT || document.getElementById('update-icon').checked;
    let logo;
    if (updateLogo == true && xmlTvFile != 'Threadfin Dummy') {
        if (SERVER['xepg']['xmltvMap'][xmlTvFile].hasOwnProperty(newXmlTvId)) {
            logo = SERVER['xepg']['xmltvMap'][xmlTvFile][newXmlTvId]['icon'];
        }
        else {
            logo = channel['tvg-logo'];
        }
    }
}
function savePopupData(dataType, id, remove, option) {
    var filterPopupValid = typeof validateFilterPopup != "function" || validateFilterPopup(dataType);
    if (remove != true && option == 0 && (!validateSourcePopup(dataType) || !filterPopupValid)) {
        return;
    }
    if (dataType == "mapping") {
        showElement("loading", true);
        var data = new Object();
        console.log("Save mapping data");
        cmd = "saveEpgMapping";
        data["epgMapping"] = SERVER["xepg"]["epgMapping"];
        console.log("SEND TO SERVER");
        var server = new Server(cmd);
        server.request(data);
        delete UNDO["epgMapping"];
        showElement("loading", false);
        return;
    }
    console.log("Save popup data");
    var div = document.getElementById("popup-custom");
    var inputs = div.getElementsByTagName("TABLE")[0].getElementsByTagName("INPUT");
    var selects = div.getElementsByTagName("TABLE")[0].getElementsByTagName("SELECT");
    var input = new Object();
    var confirmMsg;
    for (let i = 0; i < selects.length; i++) {
        var name;
        name = selects[i].name;
        var value = selects[i].value;
        switch (name) {
            case "tuner":
                input[name] = parseInt(value);
                break;
            default:
                input[name] = value;
                break;
        }
    }
    for (let i = 0; i < inputs.length; i++) {
        switch (inputs[i].type) {
            case "checkbox":
                name = inputs[i].name;
                input[name] = inputs[i].checked;
                break;
            case "text":
            case "hidden":
            case "password":
                name = inputs[i].name;
                switch (name) {
                    case "tuner":
                        input[name] = parseInt(inputs[i].value);
                        break;
                    default:
                        input[name] = inputs[i].value;
                        break;
                }
                break;
        }
    }
    var data = new Object();
    var cmd;
    if (remove == true) {
        input["delete"] = true;
    }
    switch (dataType) {
        case "users":
            confirmMsg = "Delete this user?";
            var userRequest = buildUserRequest(id, input, remove == true);
            cmd = userRequest.cmd;
            data = userRequest.data;
            break;
        case "m3u":
            confirmMsg = "Delete this playlist?";
            switch (option) {
                // Popup: Save
                case 0:
                    cmd = "saveFilesM3U";
                    break;
                // Popup: Update
                case 1:
                    cmd = "updateFileM3U";
                    break;
            }
            data["files"] = new Object;
            data["files"][dataType] = new Object;
            data["files"][dataType][id] = input;
            break;
        case "hdhr":
            confirmMsg = "Delete this HDHomeRun tuner?";
            switch (option) {
                // Popup: Save
                case 0:
                    cmd = "saveFilesHDHR";
                    break;
                // Popup: Update
                case 1:
                    cmd = "updateFileHDHR";
                    break;
            }
            data["files"] = new Object;
            data["files"][dataType] = new Object;
            data["files"][dataType][id] = input;
            break;
        case "xmltv":
            confirmMsg = "Delete this XMLTV file?";
            switch (option) {
                // Popup: Save
                case 0:
                    cmd = "saveFilesXMLTV";
                    break;
                // Popup: Update
                case 1:
                    cmd = "updateFileXMLTV";
                    break;
            }
            data["files"] = new Object;
            data["files"][dataType] = new Object;
            data["files"][dataType][id] = input;
            break;
        case "filter":
            confirmMsg = "Delete this filter?";
            cmd = "saveFilter";
            data["filter"] = new Object;
            data["filter"][id] = input;
            break;
        default:
            return;
            break;
    }
    if (remove == true) {
        if (!confirm(confirmMsg)) {
            showElement("popup", false);
            return;
        }
    }
    showElement("loading", true);
    console.log("SEND TO SERVER");
    beginSourceRequest(dataType, id, remove, option);
    if (typeof beginFilterRequest == "function") {
        beginFilterRequest(dataType, id, remove);
    }
    var server = new Server(cmd);
    server.request(data);
    showElement("loading", false);
}
function donePopupData(dataType, idsStr) {
    var ids = idsStr.split(',');
    var div = document.getElementById("popup-custom");
    var inputs = div.getElementsByClassName("changed");
    ids.forEach(id => {
        var input = new Object();
        input = SERVER["xepg"]["epgMapping"][id];
        console.log("INPUT: " + input);
        for (let i = 0; i < inputs.length; i++) {
            var name;
            var value;
            switch (inputs[i].tagName) {
                case "INPUT":
                    switch (inputs[i].type) {
                        case "checkbox":
                            name = inputs[i].name;
                            value = inputs[i].checked;
                            input[name] = value;
                            break;
                        case "text":
                            name = inputs[i].name;
                            value = inputs[i].value;
                            input[name] = value;
                            break;
                    }
                    break;
                case "SELECT":
                    name = inputs[i].name;
                    value = inputs[i].value;
                    input[name] = value;
                    break;
            }
            switch (name) {
                case "tvg-logo":
                    //(document.getElementById(id).childNodes[2].firstChild as HTMLElement).setAttribute("src", value)
                    break;
                case "x-channel-start":
                    document.getElementById(id).childNodes[3].firstChild.innerHTML = value;
                    break;
                case "x-name":
                    document.getElementById(id).childNodes[3].firstChild.innerHTML = value;
                    break;
                case "x-category":
                    var color = "white";
                    var catColorSettings = SERVER["settings"]["epgCategoriesColors"];
                    var colors_split = catColorSettings.split("|");
                    for (var ii = 0; ii < colors_split.length; ii++) {
                        var catsColor_split = colors_split[ii].split(":");
                        if (catsColor_split[0] == value) {
                            color = catsColor_split[1];
                        }
                    }
                    document.getElementById(id).childNodes[3].firstChild.style.borderColor = color;
                    break;
                case "x-group-title":
                    document.getElementById(id).childNodes[5].firstChild.innerHTML = value;
                    break;
                case "x-xmltv-file":
                    if (value != "Threadfin Dummy" && value != "-") {
                        value = getValueFromProviderFile(value, "xmltv", "name");
                    }
                    if (value == "-") {
                        input["x-active"] = false;
                    }
                    document.getElementById(id).childNodes[6].firstChild.innerHTML = value;
                    break;
                case "x-mapping":
                    if (value == "-") {
                        input["x-active"] = false;
                    }
                    document.getElementById(id).childNodes[7].firstChild.innerHTML = value;
                    break;
                case "x-backup-channel":
                    document.getElementById(id).childNodes[7].firstChild.innerHTML = value;
                    break;
                case "x-hide-channel":
                    document.getElementById(id).childNodes[7].firstChild.innerHTML = value;
                    break;
                default:
            }
            createSearchObj();
            searchInMapping();
        }
        if (input["x-active"] == false) {
            document.getElementById(id).className = "notActiveEPG";
        }
        else {
            document.getElementById(id).className = "activeEPG";
        }
        console.log(input["tvg-logo"]);
        document.getElementById(id).childNodes[2].firstChild.firstChild.setAttribute("src", input["tvg-logo"]);
    });
    showElement("popup", false);
    return;
}
function showPreview(element) {
    var div = document.getElementById("myStreamsBox");
    switch (element) {
        case false:
            div.className = "notVisible";
            return;
            break;
    }
    var streams = ["activeStreams", "inactiveStreams"];
    streams.forEach(preview => {
        var table = document.getElementById(preview);
        table.innerHTML = "";
        var obj = SERVER["data"]["StreamPreviewUI"][preview];
        var caption = document.createElement("CAPTION");
        var result = preview.replace(/([A-Z])/g, " $1");
        var finalResult = result.charAt(0).toUpperCase() + result.slice(1);
        caption.innerHTML = finalResult;
        table.appendChild(caption);
        var tbody = document.createElement("TBODY");
        table.appendChild(tbody);
        obj.slice(0, 1000).forEach(channel => {
            var tr = document.createElement("TR");
            var tdKey = document.createElement("TD");
            var tdVal = document.createElement("TD");
            tdKey.className = "tdKey";
            tdVal.className = "tdVal";
            switch (preview) {
                case "activeStreams":
                    tdKey.innerText = "Channel: (+)";
                    break;
                case "inactiveStreams":
                    tdKey.innerText = "Channel: (-)";
                    break;
            }
            tdVal.innerText = channel;
            tr.appendChild(tdKey);
            tr.appendChild(tdVal);
            tbody.appendChild(tr);
            table.appendChild(tr);
        });
    });
    // showElement("loading", false)
    div.className = "visible";
    return;
}
document.addEventListener("DOMContentLoaded", function () {
    var backToTopButton = document.createElement("button");
    backToTopButton.id = "back-to-top";
    backToTopButton.innerText = "Back to Top";
    backToTopButton.onclick = function () {
        window.scrollTo({ top: 0, behavior: 'smooth' });
    };
    document.body.appendChild(backToTopButton);
    window.addEventListener("scroll", function () {
        if (window.scrollY > 200) {
            backToTopButton.style.display = "block";
        }
        else {
            backToTopButton.style.display = "none";
        }
    });
});
