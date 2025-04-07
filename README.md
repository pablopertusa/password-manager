# Gestor de Contraseñas en Go

Esto es un **proyecto personal**. Es un gestor de contraseñas seguro desarrollado en **Golang**, con autenticación y encriptación de datos. Permite almacenar, recuperar, crear y gestionar contraseñas de forma segura. Es una forma de no depender de servicios externos para gestionar tus contraseñas. Si quieres disponer de este servicio en cualquier momento, solo tienes que desplejarlo en la nube de tu proveedor favorito y listo. Es suficiente con la capa gratuita de la mayoría de ellos. La UI es sencilla.

## Características
- Autenticación de usuario.
- Cifrado de contraseñas.
- Interfaz web para gestionar contraseñas.
- API protegida.
- Creación de contraseñas seguras.
- Despliegue con Docker.

## Requisitos
- **Docker** instalado en el sistema.

## Construcción y Ejecución con Docker
Para ejecutar la aplicación en un contenedor Docker, sigue estos pasos:

1. **Clona este repositorio:**
   ```sh
   git clone https://github.com/pablopertusa/password-manager.git
   cd password-manager
   ```

2. **Construye la imagen de Docker:**
   ```sh
   docker build -t password-manager .
   ```

3. **Ejecuta el contenedor:**
   ```sh
   docker run -p 2727:2727 --name password-manager -d password-manager
   ```

4. **Accede a la aplicación:**
   - Abre tu navegador y ve a:  
     **http://localhost:2727**

## Estructura del Proyecto
```
/
├── utils/          # Funciones auxiliares
├── static/         # Archivos estáticos (JS, CSS, etc.)
├── .env            # Variables de entorno
├── go.mod          # Dependencias del proyecto
├── go.sum          # Hashes de dependencias
├── main.go         # Punto de entrada del servidor
├── Dockerfile      # Definición de la imagen Docker
└── README.md
```

## Variables de Entorno
Asegúrate de configurar el archivo **.env** antes de ejecutar la aplicación. Ejemplo:
```
USER_NAME=yourUser
JWT_KEY=yourKey
```

## Contribuciones
Si deseas contribuir a este proyecto, ¡eres bienvenido! Puedes hacer un **fork**, crear una rama y abrir un **pull request**.