"use strict";
var navigationGroups = [
    { key: "overview", label: "Overview", items: ["overview"] },
    { key: "sources", label: "Sources", items: ["playlist", "xmltv"] },
    { key: "lineup", label: "Lineup", items: ["filter", "mapping"] },
    { key: "delivery", label: "Delivery", items: ["connections", "users"] },
    { key: "system", label: "System", items: ["activity", "settings", "log"] },
    { key: "account", label: "Account", items: ["logout"] },
];
// These indexes are the backend's established numeric menu contract. Keep this
// map explicit: the grouped navigation order must never become a routing API.
var legacyMenuIndexByDestination = {
    playlist: 0,
    xmltv: 1,
    filter: 2,
    mapping: 3,
    users: 4,
    settings: 5,
    log: 6,
    logout: 7,
};
var legacyDestinationByMenuIndex = {
    0: "playlist",
    1: "xmltv",
    2: "filter",
    3: "mapping",
    4: "users",
    5: "settings",
    6: "log",
    7: "logout",
};
var currentDestination;
function renderNavigation() {
    var navigation = document.getElementById("main-menu");
    if (!navigation) {
        return;
    }
    var navigationElement = navigation;
    initializeLegacyMenuItems();
    navigationElement.innerHTML = "";
    navigationGroups.forEach(group => {
        var items = group.items.filter(destination => navigationDestinationIsVisible(destination));
        if (items.length == 0) {
            return;
        }
        var groupElement = document.createElement("section");
        groupElement.className = "tf-navigation-group";
        groupElement.setAttribute("aria-labelledby", "navigation-group-" + group.key);
        var heading = document.createElement("h2");
        heading.id = "navigation-group-" + group.key;
        heading.className = "tf-navigation-group-label";
        heading.textContent = group.label;
        groupElement.appendChild(heading);
        var list = document.createElement("ul");
        list.className = "tf-navigation-list";
        items.forEach(destination => {
            var listItem = document.createElement("li");
            var button = document.createElement("button");
            button.type = "button";
            button.className = "tf-navigation-item";
            button.setAttribute("data-destination", destination);
            button.textContent = navigationDestinationLabel(destination);
            button.addEventListener("click", function () {
                openDestination(destination, true);
            });
            listItem.appendChild(button);
            list.appendChild(listItem);
        });
        groupElement.appendChild(list);
        navigationElement.appendChild(groupElement);
    });
    renderLegacyMenuAdapters(navigationElement);
    updateNavigationCurrentPage();
}
function initializeLegacyMenuItems() {
    menuItems.forEach(item => item.initializeTableHeader());
}
function renderLegacyMenuAdapters(navigation) {
    var adapters = document.createElement("div");
    adapters.id = "legacy-menu-adapters";
    adapters.hidden = true;
    Object.keys(legacyDestinationByMenuIndex).forEach(key => {
        var index = Number(key);
        var adapter = document.createElement("button");
        adapter.type = "button";
        adapter.id = key;
        adapter.setAttribute("data-legacy-menu-index", key);
        adapter.addEventListener("click", function () {
            openLegacyMenu(index);
        });
        adapters.appendChild(adapter);
    });
    navigation.appendChild(adapters);
}
function navigationDestinationIsVisible(destination) {
    if (destination != "users" && destination != "logout") {
        return true;
    }
    var server = SERVER;
    return server["settings"] && server["settings"]["authentication.web"] == true;
}
function navigationDestinationLabel(destination) {
    switch (destination) {
        case "overview":
            return "Overview";
        case "connections":
            return "Connections";
        case "activity":
            return "Activity";
        default:
            var index = legacyMenuIndexByDestination[destination];
            return menuItems[index].value;
    }
}
function openDestination(destination, addHistory) {
    var legacyIndex = legacyMenuIndexByDestination[destination];
    if (legacyIndex !== undefined) {
        openLegacyMenu(legacyIndex, false);
    }
    else {
        showDestinationHost(destination);
        setCurrentDestination(destination);
        dismissMobileNavigation();
        focusMainContent();
    }
    if (addHistory) {
        window.history.pushState({ threadfinDestination: destination }, "", "#" + destination);
    }
}
// This is the only adapter for numeric destinations sent by the backend.
function openLegacyMenu(index, replaceHistory = true) {
    var destination = legacyDestinationByMenuIndex[index];
    if (!destination) {
        console.warn("Ignoring unknown legacy menu index", index);
        return;
    }
    showDestinationHost("content");
    var content = new ShowContent(index);
    content.show();
    enableGroupSelection(".bulk");
    setCurrentDestination(destination);
    dismissMobileNavigation();
    focusMainContent();
    if (replaceHistory) {
        window.history.replaceState({ threadfinDestination: destination }, "", "#" + destination);
    }
}
function showDestinationHost(destination) {
    var hostByDestination = {
        overview: "overview-content",
        connections: "connections-content",
        activity: "activity-content",
        content: "content",
    };
    Object.keys(hostByDestination).forEach(key => {
        var host = document.getElementById(hostByDestination[key]);
        if (host) {
            host.hidden = key != destination;
        }
    });
    if (destination == "activity") {
        showPreview(true);
    }
}
function setCurrentDestination(destination) {
    currentDestination = destination;
    updateNavigationCurrentPage();
}
function updateNavigationCurrentPage() {
    var items = document.querySelectorAll("#main-menu [data-destination]");
    for (var i = 0; i < items.length; i++) {
        var item = items[i];
        if (item.getAttribute("data-destination") == currentDestination) {
            item.setAttribute("aria-current", "page");
        }
        else {
            item.removeAttribute("aria-current");
        }
    }
}
function dismissMobileNavigation() {
    var navigation = document.getElementById("main-menu");
    var toggle = document.querySelector(".tf-nav-toggle");
    if (!navigation) {
        return;
    }
    if (typeof bootstrap != "undefined" && bootstrap.Collapse) {
        var collapse = bootstrap.Collapse.getInstance(navigation);
        if (collapse) {
            collapse.hide();
        }
        else if (navigation.classList.contains("show")) {
            new bootstrap.Collapse(navigation).hide();
        }
    }
    if (toggle) {
        toggle.setAttribute("aria-expanded", "false");
    }
}
function focusMainContent() {
    var main = document.getElementById("main-content");
    if (main) {
        window.setTimeout(function () {
            main.focus();
        }, 0);
    }
}
function restoreDestinationFromHistory() {
    var destination = window.history.state && window.history.state.threadfinDestination;
    if (!destination && window.location.hash.length > 1) {
        destination = window.location.hash.slice(1);
    }
    if (navigationDestinationIsKnown(destination)) {
        openDestination(destination, false);
    }
}
function navigationDestinationIsKnown(destination) {
    return navigationGroups.some(group => group.items.indexOf(destination) != -1);
}
window.addEventListener("popstate", restoreDestinationFromHistory);
