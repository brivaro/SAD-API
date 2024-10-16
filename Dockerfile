# Este Dockerfile usa multi-stage build. 
# La primera etapa compila la aplicación.
# La segunda crea una imagen ligera basada en Alpine Linux.

# Etapa 1: construir la aplicación Go
FROM golang:alpine AS builder

# Configurar directorio de trabajo
WORKDIR /app

# Copiar los archivos go.mod y go.sum, y descargamos dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copiar el código de la aplicación
COPY . .

# Construir la aplicación
RUN go build -o main

# Etapa 2: imagen mínima para producción
FROM alpine

# Crear un directorio de trabajo
WORKDIR /app

# Copiar la aplicación construida desde la primera etapa
COPY --from=builder /app/main .

# Exponer el puerto de la aplicación
EXPOSE 8080

# Comando de ejecución que incluye wait-for-it para esperar la DB
CMD ["./main"]
