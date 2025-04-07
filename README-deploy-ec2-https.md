# 🚀 Despliegue de Servicio Web en AWS EC2 con Docker Compose y HTTPS

Este documento resume los pasos para desplegar un servicio web Dockerizado en una instancia EC2 de AWS con dominio personalizado y HTTPS gratuito usando Let's Encrypt.

---

## 🧩 Requisitos

- Cuenta de AWS
- Imagen Docker de tu app subida a Docker Hub
- Dominio propio (ej. `app.midominio.com`)
- Docker y Docker Compose instalados en EC2

---

## 1️⃣ Crear una instancia EC2 (Free Tier)

1. Entra a la [Consola EC2](https://console.aws.amazon.com/ec2).
2. Lanza una instancia con Amazon Linux 2 o Ubuntu.
3. Usa tipo `t2.micro` (Free Tier).
4. Abre los puertos:
   - 22 (SSH)
   - 80 (HTTP)
   - 443 (HTTPS)
5. Conéctate por SSH:
   ```bash
   ssh -i tu-clave.pem ubuntu@<IP_PUBLICA>
   ```

---

## 2️⃣ Instalar Docker y Docker Compose en EC2

```bash
# Para Ubuntu
sudo apt update
sudo apt install docker.io docker-compose -y
sudo usermod -aG docker $USER
newgrp docker
```

---

## 3️⃣ Comprar un dominio y apuntarlo a EC2

1. Compra un dominio (Namecheap, GoDaddy, etc.).
2. Apunta un **registro A** a la IP pública de tu instancia:

```
Tipo: A
Nombre: app
Valor: <IP_PUBLICA>
```

---

## 4️⃣ Estructura del proyecto

```
/tu-proyecto
├── docker-compose.yml
└── nginx/
    └── default.conf.template
```

---

## 5️⃣ docker-compose.yml

```yaml
version: '3.8'

services:
  app:
    image: tu_usuario/mi-servicio
    container_name: app
    expose:
      - 5000

  nginx:
    image: nginx:latest
    container_name: reverse-proxy
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/default.conf.template:/etc/nginx/templates/default.conf.template
      - certs:/etc/letsencrypt
      - certs-data:/data/letsencrypt
    environment:
      - NGINX_ENVSUBST_TEMPLATE_SUFFIX=.template
    depends_on:
      - app

  certbot:
    image: certbot/certbot
    container_name: certbot
    volumes:
      - certs:/etc/letsencrypt
      - certs-data:/data/letsencrypt
    entrypoint: "/bin/sh -c 'trap exit TERM; while :; do sleep 1d & wait $${!}; certbot renew --webroot -w /data/letsencrypt --quiet; done'"

volumes:
  certs:
  certs-data:
```

---

## 6️⃣ nginx/default.conf.template

```nginx
server {
    listen 80;
    server_name app.midominio.com;

    location /.well-known/acme-challenge/ {
        root /data/letsencrypt;
    }

    location / {
        return 301 https://$host$request_uri;
    }
}

server {
    listen 443 ssl;
    server_name app.midominio.com;

    ssl_certificate /etc/letsencrypt/live/app.midominio.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/app.midominio.com/privkey.pem;

    location / {
        proxy_pass http://app:5000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

---

## 7️⃣ Obtener el certificado SSL (una sola vez)

```bash
docker run --rm \
  -v certs:/etc/letsencrypt \
  -v certs-data:/data/letsencrypt \
  certbot/certbot certonly \
  --webroot \
  --webroot-path=/data/letsencrypt \
  -d app.midominio.com \
  --email tu-email@ejemplo.com \
  --agree-tos \
  --no-eff-email
```

---

## 8️⃣ Iniciar todos los servicios

```bash
docker compose up -d
```

---

## ✅ Resultado final

Tu app será accesible en:

```
https://app.midominio.com
```

El certificado se renovará automáticamente gracias al servicio `certbot`.

---

## 🧼 Buenas prácticas

- Configura alertas de uso en AWS para no exceder el Free Tier.
- Haz `docker compose pull` regularmente si tu imagen se actualiza en Docker Hub.
- Puedes usar Cloudflare para protección adicional DNS/SSL si lo deseas.

---
