function getQueryParam(param) {
    const urlParams = new URLSearchParams(window.location.search);
    return urlParams.get(param);
}

function decryptPassword() {
    const service = getQueryParam("service");
    const passphrase = document.getElementById("passphrase").value;

    if (!passphrase) {
        alert("Por favor, ingrese la passphrase.");
        return;
    }

    fetch(`/protected/decrypt-password?service=${service}&passphrase=${passphrase}`)
        .then(response => response.json())
        .then(data => {
            const resultElement = document.getElementById("password-result");
            if (data.error) {
                resultElement.style.color = "red";
                resultElement.textContent = "Error: " + data.error;
                document.getElementById("edit-container").style.display = "none"
            } else {
                resultElement.style.color = "lime";
                resultElement.textContent = "Contraseña: " + data.password;
                document.getElementById("edit-button").disabled = false
            }
        })
        .catch(error => {
            document.getElementById("password-result").style.color = "red";
            document.getElementById("password-result").textContent = "Error en la solicitud.";
            document.getElementById("edit-container").style.display = "none"
        });
}

function showEditForm() {
    document.getElementById("edit-container").style.display = "block";
}

function updatePassword() {
    const service = getQueryParam("service");
    const newPassword = document.getElementById("new-password").value;
    const passphrase = document.getElementById("passphrase").value;

    if (!passphrase) {
        alert("Por favor, ingrese la passphrase.");
        return;
    }
    
    if (!newPassword) {
        alert("Por favor, ingrese la nueva contraseña.");
        return;
    }

    fetch(`/protected/update-password?service=${service}&newPassword=${newPassword}&passphrase=${passphrase}`, {
        method: "POST",
    })
    .then(response => response.json())
    .then(data => {
        console.log(data)
        if (data.success) {
            alert("Contraseña actualizada correctamente");
            document.getElementById("edit-container").style.display = "none";
            document.getElementById("password-update-result").textContent = "Contraseña actualizada";
            document.getElementById("password-result").textContent = "";
        } else {
            alert("Error al actualizar la contraseña");
        }
    })
    .catch(error => {
        alert("Error en la solicitud.");
    });
}

function goHome() {
    window.location.href = "/protected/home";
}