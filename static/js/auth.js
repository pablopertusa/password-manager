document.getElementById('identityForm').addEventListener('submit', function(event) {
    event.preventDefault();

    const name = document.getElementById('name').value;
    console.log("activado")

    fetch('/login', {
        method: 'POST',
        body: new URLSearchParams({
            'name': name
        })
    })
    .then(response => response.json())
    .then(data => {
        if (data.token) {
            // Almacenar el token en el almacenamiento local
            localStorage.setItem('authToken', data.token);
            window.location.href = "/home";  // Redirigir a la página de inicio protegida
        } else {
            alert('Error al obtener el token');
        }
    })
    .catch(error => {
        alert('Error al enviar la solicitud: ' + error.message);
    });
});

// Función para hacer una solicitud autorizada y redirigir a la página recibida
async function makeAuthorizedRequestAndRedirect(url, method = 'GET', body = null) {
    // Obtén el token almacenado (puede estar en localStorage o en una cookie)
    const token = localStorage.getItem('authToken');

    if (!token) {
        alert('No estás autenticado');
        return;
    }

    // Configuración de la solicitud
    const headers = {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}` // Añadir el token JWT en los headers
    };

    // Si hay cuerpo (en caso de POST, PUT, DELETE), agregarlo a la solicitud
    if (body) {
        headers['Content-Type'] = 'application/json';
        body = JSON.stringify(body);
    }

    try {
        // Realiza la solicitud HTTP
        const response = await fetch(url, {
            method: method,
            headers: headers,
            body: body,
        });

        // Verifica si la respuesta fue exitosa
        if (!response.ok) {
            throw new Error('Error en la solicitud');
        }

        // Si la respuesta es un HTML, se redirige al contenido de la respuesta
        const htmlContent = await response.text();
        
        // Crea un nuevo documento HTML desde la respuesta y redirige
        const parser = new DOMParser();
        const doc = parser.parseFromString(htmlContent, 'text/html');
        
        // Redirige al usuario al URL de la respuesta o la URL que se obtiene
        window.location.href = doc.location.href;
        
    } catch (error) {
        console.error('Error haciendo la solicitud autorizada:', error);
        alert('Ocurrió un error en la solicitud.');
    }
}

