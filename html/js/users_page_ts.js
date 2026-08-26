"use strict";
var USER_PERMISSION_DEFINITIONS = [
    { key: "authentication.web", label: "WEB", description: "Sign in to the Threadfin web interface and manage this user's access." },
    { key: "authentication.pms", label: "PMS", description: "Authenticate DVR discovery and access used by Plex, Emby, or Jellyfin." },
    { key: "authentication.m3u", label: "M3U", description: "Authenticate requests for the generated M3U playlist." },
    { key: "authentication.xml", label: "XML", description: "Authenticate requests for the generated XMLTV guide." },
    { key: "authentication.api", label: "API", description: "Authenticate access to Threadfin's API commands." },
];
function userActions(data) {
    var defaultUser = Boolean(data && data.defaultUser);
    return { canDelete: !defaultUser, webLocked: defaultUser };
}
function userRowValues(record) {
    var data = record && record.data && typeof record.data == "object" ? record.data : {};
    var row = { username: String(data.username || "") };
    USER_PERMISSION_DEFINITIONS.forEach(permission => {
        row[permission.key] = data[permission.key] === true;
    });
    row.defaultUser = data.defaultUser === true;
    return row;
}
function buildUserRequest(id, input, remove) {
    var values = {};
    Object.keys(input || {}).forEach(key => { values[key] = input[key]; });
    if (remove) {
        values.delete = true;
    }
    if (id == "-") {
        return { cmd: "saveNewUser", data: { userData: values } };
    }
    var users = {};
    users[id] = values;
    return { cmd: "saveUserData", data: { userData: users } };
}
function enhanceUsersPopup(id, data) {
    var popup = document.getElementById("popup-custom");
    if (!popup) {
        return;
    }
    var username = popup.querySelector('[name="username"]');
    var password = popup.querySelector('[name="password"]');
    var confirmPassword = popup.querySelector('[name="confirm"]');
    if (username) {
        username.autocomplete = "username";
    }
    if (password) {
        password.autocomplete = "new-password";
    }
    if (confirmPassword) {
        confirmPassword.autocomplete = "new-password";
    }
    var actions = userActions(data);
    var web = popup.querySelector('[name="authentication.web"]');
    if (web && actions.webLocked) {
        web.disabled = true;
        web.setAttribute("aria-describedby", "default-user-web-note");
        var note = document.createElement("p");
        note.id = "default-user-web-note";
        note.className = "tf-field-note";
        note.textContent = "WEB access is required for the default user.";
        web.parentElement.appendChild(note);
    }
}
function appendUserPermissionDetails(identity, values) {
    var details = document.createElement("details");
    details.className = "tf-user-permission-details";
    var summary = document.createElement("summary");
    summary.textContent = "Permissions";
    details.appendChild(summary);
    var list = document.createElement("dl");
    USER_PERMISSION_DEFINITIONS.forEach(permission => {
        var term = document.createElement("dt");
        term.textContent = permission.label;
        var description = document.createElement("dd");
        description.textContent = (values[permission.key] ? "Allowed. " : "Denied. ") + permission.description;
        list.appendChild(term);
        list.appendChild(description);
    });
    details.appendChild(list);
    identity.appendChild(details);
}
function renderUsersPage(doc) {
    doc.innerHTML = "";
    var root = document.createElement("div");
    root.className = "tf-users";
    var header = document.createElement("header");
    header.className = "tf-admin-header";
    var headingGroup = document.createElement("div");
    var heading = document.createElement("h1");
    heading.textContent = "Users";
    var purpose = document.createElement("p");
    purpose.textContent = "Control which Threadfin interfaces and generated endpoints each account may use. Passwords are available only while creating or editing an account.";
    headingGroup.appendChild(heading);
    headingGroup.appendChild(purpose);
    var add = document.createElement("button");
    add.type = "button";
    add.className = "tf-primary-action";
    add.textContent = "New user";
    add.addEventListener("click", () => openPopUp("users", undefined));
    header.appendChild(headingGroup);
    header.appendChild(add);
    root.appendChild(header);
    var explanations = document.createElement("dl");
    explanations.className = "tf-permission-guide";
    USER_PERMISSION_DEFINITIONS.forEach(permission => {
        var term = document.createElement("dt");
        term.textContent = permission.label;
        var description = document.createElement("dd");
        description.textContent = permission.description;
        explanations.appendChild(term);
        explanations.appendChild(description);
    });
    root.appendChild(explanations);
    var scroll = document.createElement("div");
    scroll.className = "tf-admin-table-scroll";
    var table = document.createElement("table");
    table.className = "tf-users-table";
    var caption = document.createElement("caption");
    caption.textContent = "Threadfin users and endpoint permissions";
    table.appendChild(caption);
    var head = document.createElement("thead");
    var headRow = document.createElement("tr");
    ["Username"].concat(USER_PERMISSION_DEFINITIONS.map(permission => permission.label), ["Actions"]).forEach(label => {
        var th = document.createElement("th");
        th.scope = "col";
        th.textContent = label;
        headRow.appendChild(th);
    });
    head.appendChild(headRow);
    table.appendChild(head);
    var body = document.createElement("tbody");
    var users = SERVER && SERVER["users"] && typeof SERVER["users"] == "object" ? SERVER["users"] : {};
    Object.keys(users).sort((left, right) => userRowValues(users[left]).username.localeCompare(userRowValues(users[right]).username)).forEach(userID => {
        var values = userRowValues(users[userID]);
        var tr = document.createElement("tr");
        var username = document.createElement("th");
        username.scope = "row";
        username.className = "tf-user-identity";
        username.textContent = values.username || "Unnamed user";
        if (values.defaultUser) {
            var defaultLabel = document.createElement("span");
            defaultLabel.className = "tf-user-default";
            defaultLabel.textContent = "Default user";
            username.appendChild(defaultLabel);
        }
        appendUserPermissionDetails(username, values);
        tr.appendChild(username);
        USER_PERMISSION_DEFINITIONS.forEach(permission => {
            var td = document.createElement("td");
            td.className = "tf-user-permission-cell";
            td.textContent = values[permission.key] ? "Allowed" : "Denied";
            td.setAttribute("data-allowed", values[permission.key] ? "true" : "false");
            tr.appendChild(td);
        });
        var actions = document.createElement("td");
        actions.className = "tf-user-actions";
        var edit = document.createElement("button");
        edit.type = "button";
        edit.id = userID;
        edit.textContent = "Edit";
        edit.addEventListener("click", () => openPopUp("users", edit));
        actions.appendChild(edit);
        tr.appendChild(actions);
        body.appendChild(tr);
    });
    table.appendChild(body);
    scroll.appendChild(table);
    root.appendChild(scroll);
    doc.appendChild(root);
}
