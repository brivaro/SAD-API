package main

import (
    "database/sql"      // Para conexión con la base de datos
    "net/http"          // Importamos el paquete para manejar las respuestas HTTP
    "strconv"  			// Importamos strconv para convertir el ID de string a int
    "log"			    // Importamos log para realizar el sistema de logging
    "os"			    // Importamos os para guardar el logger en la carpeta de la app
    "time"			    // Importamos time para guardar la fecha de inicio de sesiones del logger
    // "fmt"			// Este lo he quitado, lo he usado para debugging
    "github.com/gin-gonic/gin"
    _"github.com/lib/pq"
)

var db *sql.DB

func initDB() {
    // Cadena de conexión
    // Esto es una práctica insegura, poner el user y pass en el código
    // Se podrí­a hacer con variables de entorno, os.Getenv
    connStr := "user=user dbname=mydatabase password=pass host=db sslmode=disable"

    // Abre la conexión
    tempdb, err := sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal(err)
    }
    
    db = tempdb

    // Verifica la conexión
    err = db.Ping()
    if err != nil {
        log.Fatal("No se pudo conectar a la base de datos:", err)
    }
    log.Println("ConexiÃ³n exitosa a la base de datos")

    // Crear la tabla si no existe
    createTableSQL := `CREATE TABLE IF NOT EXISTS todos (
        id SERIAL PRIMARY KEY,
        task TEXT NOT NULL,
        completed BOOLEAN NOT NULL DEFAULT FALSE
    );`
    _, err = db.Exec(createTableSQL)
    if err != nil {
        log.Fatal("Error creando la tabla:", err)
    }
}

// Definimos una estructura 'toDo' que representará una tarea en nuestra lista
type toDo struct {
    ID        int    `json:"id"`        // ID de la tarea, entero
    Task      string `json:"task"`      // La descripción de la tarea, string
    Completed bool   `json:"completed"` // Indicador de si la tarea completada o no, boolean
}

// Inicializamos una lista de tareas (en este caso, dos tareas predefinidas)
var toDos = []toDo{
    {ID: 1, Task: "Learn Golang", Completed: false},        // Tarea 1
    {ID: 2, Task: "Build a REST API", Completed: false},    // Tarea 2
}

func main() {
    ////////////////////////////////////
    // Inicializamos la base de datos
    initDB()

    // Verifica si ya existen tareas en la tabla
    var count int
    err := db.QueryRow("SELECT COUNT(*) FROM todos").Scan(&count)
    if err != nil {
        log.Println("Error checking task count:", err)
        return
    }

    // Si no hay tareas, inserta las predefinidas
    if count == 0 {
        // Inserta cada tarea en la base de datos
        for _, todo := range toDos {
            _, err := db.Exec("INSERT INTO todos (id, task, completed) VALUES ($1, $2, $3)", todo.ID, todo.Task, todo.Completed)
            if err != nil {
                log.Println("Error inserting task:", err)
            } else {
                log.Printf("Inserted task: %s\n", todo.Task)
            }
        }
    } else {
        log.Println("BD exists. Tasks already exist in the database, skipping initialization.")
    }

    ////////////////////////////////////
    // Preparar log file and router
    // Configurar el logger para registrar en un archivo
    file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    // Establecer el archivo como salida del logger
    log.SetOutput(file)

    // Registrar la fecha y hora de inicio de sesiÃ³n
    log.Println("======== Inicio de sesiÃ³n: ", time.Now().Format("2006-01-02 15:04:05"), "========")

    // Creamos un nuevo router de Gin, que serÃ¡ el manejador de las rutas y peticiones HTTP
    router := gin.Default()

    //////////////////////////////////// GET

    // Definimos la ruta para obtener todas las tareas ('GET' en '/toDos')
    router.GET("/toDos", func(c *gin.Context) {
        log.Println("Received request to get all tasks")
        rows, err := db.Query("SELECT id, task, completed FROM todos")
        if err != nil {
            log.Println("Error querying tasks:", err)
            // Respondemos con un JSON que contiene la lista de tareas (toDos)
            c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving tasks"})
            return
        }
        defer rows.Close()

        var toDos []toDo
        // Devolver JSON de las tareas
        for rows.Next() {
            var todo toDo
            if err := rows.Scan(&todo.ID, &todo.Task, &todo.Completed); err != nil {
                log.Println("Error scanning task:", err)
                c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving tasks"})
                return
            }
            toDos = append(toDos, todo)
        }

        c.IndentedJSON(http.StatusOK, toDos)
    })

    //////////////////////////////////// POST

    // Definimos la ruta para añadir una nueva tarea ('POST' en '/toDos')
    router.POST("/toDos", func(c *gin.Context) {
        var newToDo toDo // Creamos una variable para almacenar la nueva tarea
        log.Println("Received a new task creation request")

        // BindJSON se encarga de enlazar los datos JSON enviados en la solicitud al nuevo 'toDo'
        if err := c.BindJSON(&newToDo); err != nil {
            log.Println("Error binding JSON:", err)
            // Si hay un error en el proceso, simplemente se sale de la funciÃ³n
            c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid task format -- ID is int, Task is string, Completed is bool"})
            return
        }

        // Comprobar si el ID ya existe en la base de datos
        var exists bool
        err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM todos WHERE id = $1)", newToDo.ID).Scan(&exists)
        if err != nil {
            log.Println("Error checking if task exists:", err)
            c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error checking task existence"})
            return
        }

        if exists {
            log.Println("Task with the same ID already exists:", newToDo.ID)
            c.IndentedJSON(http.StatusConflict, gin.H{"message": "Task with this ID already exists"})
            return
        }

        // Inserta la nueva tarea en la base de datos
        _, err = db.Exec("INSERT INTO todos (id, task, completed) VALUES ($1, $2, $3)", newToDo.ID, newToDo.Task, newToDo.Completed)
        if err != nil {
            log.Println("Error inserting new task:", err)
            c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error creating task"})
            return
        }

        log.Println("Added new task:", newToDo)
        c.IndentedJSON(http.StatusCreated, newToDo)
    })

    //////////////////////////////////// PUT

    // Definimos la ruta para actualizar una tarea existente ('PUT' en '/toDos/:id')
    router.PUT("/toDos/:id", func(c *gin.Context) {
        var updatedToDo toDo       // Creamos una variable para almacenar la tarea actualizada
        id := c.Param("id")        // Obtenemos el parÃ¡metro 'id' de la URL (el identificador de la tarea)
        log.Println("Received a new task update request")

        // Enlazamos los datos del cuerpo de la solicitud a 'updatedToDo'
        if err := c.BindJSON(&updatedToDo); err != nil {
            log.Println("Error binding JSON:", err)
            c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid task format -- ID is int, Task is string, Completed is bool"})
            return
        }

        // Convertimos el id de string a int
        intID, _ := strconv.Atoi(id)

        // Antes de actualizar la BD compruebo que ese ID exista
        // Comprobamos que la tarea con ese ID exista
        var exists bool
        err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM todos WHERE id = $1)", intID).Scan(&exists)
        if err != nil {
            log.Println("Error checking if task exists:", err)
            c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error checking task existence"})
            return
        }
        
        if !exists {
            log.Println("Task with ID:", intID, "not found")
            c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Task not found"})
            return
        }

        // Actualizamos la tarea en la base de datos
        _, err = db.Exec("UPDATE todos SET task = $1, completed = $2 WHERE id = $3", updatedToDo.Task, updatedToDo.Completed, intID)
        if err != nil {
            log.Println("Error updating task:", err)
            c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error updating task"})
            return
        }

        log.Printf("Updated task with ID: %d\n", intID)
        c.IndentedJSON(http.StatusOK, updatedToDo)
    })

    //////////////////////////////////// DELETE

    // Definimos la ruta para eliminar una tarea ('DELETE' en '/toDos/:id')
    router.DELETE("/toDos/:id", func(c *gin.Context) {
        id := c.Param("id")  // Obtenemos el parÃ¡metro 'id' de la URL
        log.Println("Received a new task deletion request")

        // Convertimos el id de string a int
        intID, err := strconv.Atoi(id)
        if err != nil {
            log.Println("Error converting ID to int:", err)
            c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid ID"})
            return
        }

        // Eliminamos la tarea de la base de datos
        _, err = db.Exec("DELETE FROM todos WHERE id = $1", intID)
        if err != nil {
            log.Println("Error deleting task $1: $2", intID, err)
            c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error deleting task"})
            return
        }

        log.Printf("Deleted task with ID: %d\n", intID)
        c.IndentedJSON(http.StatusOK, gin.H{"message": "toDo deleted"})
    })

    // Ejecutamos el servidor en el puerto 8080. '0.0.0.0' significa que será accesible desde cualquier dirección IP
    router.Run("0.0.0.0:8080")
}
