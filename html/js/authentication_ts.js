"use strict";
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
    inputs.forEach(input => {
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
