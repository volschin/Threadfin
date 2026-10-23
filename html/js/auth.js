// <stdin>
var themePreference = matchMedia("(prefers-color-scheme: dark)");
var chosenTheme = "system";
try {
  chosenTheme = localStorage.getItem("threadfin-theme") || "system";
} catch {
}
if (!["system", "light", "dark"].includes(chosenTheme)) chosenTheme = "system";
function applyTheme() {
  document.documentElement.dataset.bsTheme = chosenTheme === "system" ? themePreference.matches ? "dark" : "light" : chosenTheme;
}
applyTheme();
themePreference.addEventListener("change", applyTheme);
var themeControl = document.createElement("select");
themeControl.className = "tf-theme-control form-select";
themeControl.setAttribute("aria-label", "Color theme");
for (const [value, label] of [["system", "System theme"], ["light", "Light theme"], ["dark", "Dark theme"]]) {
  const option = new Option(label, value);
  themeControl.add(option);
}
themeControl.value = chosenTheme;
themeControl.addEventListener("change", () => {
  chosenTheme = themeControl.value;
  try {
    localStorage.setItem("threadfin-theme", chosenTheme);
  } catch {
  }
  applyTheme();
});
(document.querySelector(".tf-sidebar-header") || document.querySelector(".tf-auth-card header") || document.querySelector("#headline")).append(themeControl);
function login(_event) {
  var username = document.getElementById("username");
  var password = document.getElementById("password");
  var confirmPassword = document.getElementById("confirm");
  var message = document.getElementById("err");
  var inputs = [username, password];
  if (confirmPassword) {
    inputs.push(confirmPassword);
  }
  var firstInvalid = null;
  inputs.forEach((input) => {
    input.style.borderColor = "";
    input.setAttribute("aria-invalid", "false");
    if (input.value.length == 0 && !firstInvalid) {
      firstInvalid = input;
    }
  });
  if (firstInvalid) {
    firstInvalid.style.borderColor = "red";
    firstInvalid.setAttribute("aria-invalid", "true");
    message.textContent = "{{.alert.missingInput}}";
    firstInvalid.focus();
    return false;
  }
  if (confirmPassword && confirmPassword.value != password.value) {
    password.style.borderColor = "red";
    confirmPassword.style.borderColor = "red";
    password.setAttribute("aria-invalid", "true");
    confirmPassword.setAttribute("aria-invalid", "true");
    message.textContent = "{{.account.failed}}";
    password.focus();
    return false;
  }
  message.textContent = "";
  return true;
}
document.getElementById("authentication").addEventListener("submit", (event) => {
  if (!login(event)) event.preventDefault();
});
