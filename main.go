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

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ── Métricas personalizadas ──
var (
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_requests_total",
			Help: "Total de requests por endpoint",
		},
		[]string{"endpoint", "method", "status"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "api_request_duration_seconds",
			Help:    "Duración de requests en segundos",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint"},
	)
)

func init() {
	prometheus.MustRegister(requestsTotal)
	prometheus.MustRegister(requestDuration)
}

// ── Middleware para medir cada request ──
func metricsMiddleware(endpoint string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		inicio := time.Now()
		next(w, r)
		duration := time.Since(inicio).Seconds()

		requestsTotal.WithLabelValues(endpoint, r.Method, "200").Inc()
		requestDuration.WithLabelValues(endpoint).Observe(duration)
	}
}

// ── Structs ──
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

type Salud struct {
	CPU      string `json:"cpu"`
	Memoria  string `json:"memoria"`
	Disco    string `json:"disco"`
	Hostname string `json:"hostname"`
}

var inicio = time.Now()

// ── Handlers ──
func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	respuesta := Respuesta{
		Status:  "ok",
		Mensaje: "API corriendo correctamente",
		Hora:    time.Now().Format("2006-01-02 15:04:05"),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respuesta)
}

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

func sistemaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "desconocido"
	}

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
	// Rutas con middleware de métricas
	http.HandleFunc("/health",  metricsMiddleware("/health", healthHandler))
	http.HandleFunc("/info",    metricsMiddleware("/info", infoHandler))
	http.HandleFunc("/sistema", metricsMiddleware("/sistema", sistemaHandler))

	// Endpoint de métricas para Prometheus
	http.Handle("/metrics", promhttp.Handler())

	puerto := ":8080"
	fmt.Printf("Servidor iniciando en puerto %s\n", puerto)

	if err := http.ListenAndServe(puerto, nil); err != nil {
		log.Fatal("Error iniciando servidor:", err)
	}
}
