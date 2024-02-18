# Usa la imagen oficial de Go como base
FROM golang:latest

# Establece el directorio de trabajo dentro del contenedor
WORKDIR /app

# Copia los archivos necesarios para compilar y ejecutar la aplicación
COPY . .

# Compila la aplicación
RUN go build -o main .

# Expone el puerto en el que la aplicación se ejecutará
EXPOSE 8080

# Comando para ejecutar la aplicación
CMD ["./main"]
