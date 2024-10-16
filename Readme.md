---
# To-Do Application with Go and Docker

## Descripción

Esta es una aplicación web de lista de tareas (To-Do) desarrollada en **Go** utilizando el framework **Gin**. La aplicación permite crear, actualizar, eliminar y listar tareas. Está diseñada para ser ejecutada en contenedores Docker, con orquestación mediante **Docker Compose** y una base de datos PostgreSQL.

### Funcionalidades CRUD:
1. Obtener (leer) todas las tareas.
2. Crear una nueva tarea.
3. Actualizar una tarea existente.
4. Eliminar una tarea.

## Requisitos

Para ejecutar esta aplicación, necesitarás tener instalados los siguientes componentes en tu máquina:

- **Docker**: Para ejecutar contenedores.
- **Docker Compose**: Para orquestar múltiples servicios (como la aplicación web y la base de datos).

## Estructura del Proyecto

- **`main.go`**: Contiene el código fuente de la aplicación Go.
- **`Dockerfile`**: Archivo para construir la imagen Docker de la aplicación Go.
- **`docker-compose.yml`**: Orquestación de servicios para levantar la aplicación y la base de datos PostgreSQL.
- **`app.log`**: Archivo de logs generado automáticamente cuando se ejecuta la aplicación (guarda todas las actividades importantes con marca de tiempo).
  
## Instalación

### Paso 1: Clonar el repositorio

Clona este repositorio en tu máquina local:

```bash
git clone <URL_DEL_REPOSITORIO>
cd <NOMBRE_DEL_REPOSITORIO>
```

### Paso 2: Construcción y ejecución de la aplicación con Docker Compose

Para iniciar la aplicación y la base de datos usando Docker Compose, simplemente ejecuta:

```bash
docker compose up --build
```

Este comando hace lo siguiente:

1. **Construye** la imagen Docker de la aplicación Go a partir del `Dockerfile`.
2. **Levanta** los servicios definidos en el archivo `docker-compose.yml`:
   - **web**: El servicio de la aplicación de tareas.
   - **db**: Un contenedor PostgreSQL para almacenar las tareas.
   
Una vez que Docker Compose haya levantado los contenedores, la aplicación estará disponible en `http://localhost:8080`.

### Paso 3: Verificar la aplicación

Puedes acceder a la aplicación en tu navegador o utilizar herramientas como **curl** o **Postman** para interactuar con la API.

- **Obtener todas las tareas (GET /toDos)**

  ```bash
  curl -X GET http://localhost:8080/toDos
  ```

- **Crear una nueva tarea (POST /toDos)**

  ```bash
  curl -X POST http://localhost:8080/toDos -H 'Content-Type: application/json' -d '{"id": 3, "task": "Learn Docker", "completed": false}'
  ```

- **Actualizar una tarea existente (PUT /toDos/:id)**

  ```bash
  curl -X PUT http://localhost:8080/toDos/3 -H 'Content-Type: application/json' -d '{"id": 3, "task": "Learn Docker", "completed": true}'
  ```

- **Eliminar una tarea (DELETE /toDos/:id)**

  ```bash
  curl -X DELETE http://localhost:8080/toDos/3
  ```

## Detalles del Código

### `main.go`

Este archivo contiene la implementación principal de la aplicación. Se utilizan los siguientes componentes:

- **Gin**: Un framework para crear APIs REST de forma sencilla.
- **Log**: Se configura un logger que guarda todas las peticiones y acciones en un archivo llamado `app.log`. Al inicio de cada sesión, se registra la fecha y hora para poder distinguir entre sesiones de ejecución.

Las rutas principales incluyen:

1. **GET `/toDos`**: Obtiene todas las tareas.
2. **POST `/toDos`**: Crea una nueva tarea.
3. **PUT `/toDos/:id`**: Actualiza una tarea existente.
4. **DELETE `/toDos/:id`**: Elimina una tarea específica.

### Dockerfile

El `Dockerfile` sigue una estrategia de **multi-stage build** para optimizar el tamaño de la imagen final. Los pasos son los siguientes:

1. La primera etapa **compila** el código Go.
2. La segunda etapa crea una **imagen ligera** basada en Alpine Linux que contiene solo el binario resultante.

```dockerfile
# Etapa 1: construir la aplicación Go
FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main

# Etapa 2: imagen mínima para producción
FROM alpine
WORKDIR /app
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
```

### Docker Compose

El archivo `docker-compose.yml` configura dos servicios:

1. **web**: La aplicación Go.
2. **db**: Un contenedor PostgreSQL para almacenar los datos.

```yaml
services:
  web:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      - db

  db:
    image: postgres
    environment:
      POSTGRES_USER: user
      POSTGRES_PASS: pass
      POSTGRES_DB_NAME: mydatabase
    volumes:
      - db-data:/var/lib/postgresql/data

volumes:
  db-data:
```

### Volúmenes

El servicio de base de datos utiliza un volumen para persistir los datos incluso si el contenedor de la base de datos se detiene o se elimina.

## Logging

La aplicación crea un archivo de logs (`app.log`) en el directorio principal de la aplicación. Cada vez que se inicia la aplicación, se registra la fecha y hora de inicio de la sesión.

## Container Seminario1-web-1 (servicio web) y Seminario1-db-1 (servicio db)
Si abrimos bash en el contenedor donde estamos ejecutando nuestra aplicación GO, podremos observar mediante el comando ```ls``` que solamente tenemos el binario de nuestra app (main) y el archivo de logs (app.log).

```bash
docker exec -it <id_contenedor> /bin/sh
```

## Cierre de la Aplicación

Para detener los servicios de los contenedores, ejecuta:

```bash
docker compose stop
```

Para volver a crear/iniciar los servicios de los contenedores, ejecuta:
```bash
docker compose up
```
o
```bash
docker compose start
```

Para detener la aplicación y eliminar los contenedores, ejecuta:

```bash
docker compose down
```

Esto detendrá y eliminará todos los contenedores creados por Docker Compose, pero los datos de la base de datos permanecerán debido al uso del volumen persistente.

---
