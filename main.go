package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)
// Struct — equivale a una clase en PHP pero más simple
type Respuesta struct {
	Status  string `json:"status"`
	Mensaje string `json:"mensaje"`
	Hora    string `json:"hora"`
}

type Servidor struct {
	Nombre  string `json:"nombre"`
	Version string `json:"version"`
	Uptime  string `json:"uptime"`
}

// Función que maneja el endpoint /health
func healthHandler(w http.ResponseWriter, r *http.Request) {
	// Solo aceptar GET
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	respuesta := Respuesta{
		Status:  "ok",
		Mensaje: "API corriendo correctamente",
		Hora:    time.Now().Format("2006-01-02 15:04:05"),
	}

	// Convertir struct a JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respuesta)
}

// Función que maneja el endpoint /info
func infoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	servidor := Servidor{
		Nombre:  "devops-api",
		Version: "1.1.0",
		Uptime:  time.Since(inicio).String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(servidor)
}

// Variable global para calcular uptime
var inicio = time.Now()
type Salud struct {
	CPU      string `json:"cpu"`
	Memoria  string `json:"memoria"`
	Disco    string `json:"disco"`
	Hostname string `json:"hostname"`
}

func sistemaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "desconocido"
	}

	// Leer memoria desde /proc/meminfo
	memInfo, err := os.ReadFile("/proc/meminfo")
	memoria := "no disponible"
	if err == nil {
		for _, linea := range strings.Split(string(memInfo), "\n") {
			if strings.HasPrefix(linea, "MemAvailable:") {
				memoria = strings.TrimSpace(strings.TrimPrefix(linea, "MemAvailable:"))
				break
			}
		}
	}

	salud := Salud{
		CPU:      fmt.Sprintf("%d núcleos", runtime.NumCPU()),
		Memoria:  memoria,
		Disco:    "ver /proc/diskstats",
		Hostname: hostname,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(salud)
}
func main() {
	// Registrar rutas
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/info", infoHandler)
        http.HandleFunc("/sistema", sistemaHandler)

	puerto := ":8080"
	fmt.Printf("Servidor iniciando en puerto %s\n", puerto)

	// Iniciar servidor — si falla, log.Fatal detiene el programa
	if err := http.ListenAndServe(puerto, nil); err != nil {
		log.Fatal("Error iniciando servidor:", err)
	}
}
