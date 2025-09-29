package main

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
)

type TemperatureResponse struct {
	Location    string  `json:"location"`
	SensorID    string  `json:"sensor_id"`
	Value 		float64 `json:"value"`
}

var locationToSensorID = map[string]string{
	"Living Room": "1",
	"Bedroom":     "2",
	"Kitchen":     "3",
}

var sensorIDToLocation = map[string]string{
	"1": "Living Room",
	"2": "Bedroom",
	"3": "Kitchen",
}

// secureRandomFloat возвращает случайное число от 0.0 до 1.0
func secureRandomFloat() float64 {
	var b [8]byte
	_, err := rand.Read(b[:])
	if err != nil {
		panic(err) // в реальном коде обработать ошибку
	}
	u := binary.BigEndian.Uint64(b[:])
	// нормализуем к [0,1)
	return float64(u) / float64(math.MaxUint64)
}

func main() {
	http.HandleFunc("/temperature", temperatureHandler)
	http.HandleFunc("/temperature/", temperatureHandler)
	port := getEnv("PORT", ":8081")

	fmt.Printf("Server running on %s\n", port)

	log.Fatal(http.ListenAndServe(port, nil))
}

func temperatureHandler(w http.ResponseWriter, r *http.Request) {
	location := r.URL.Query().Get("location")
	sensorID := ""

	// Если не указан location, попробуем достать sensorId из URL (/temperature/<id>)
	if location == "" {
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) == 2 && pathParts[0] == "temperature" {
			sensorID = pathParts[1]
			if loc, ok := sensorIDToLocation[sensorID]; ok {
				location = loc
			} else {
				location = "Unknown"
			}
		}
	}

	// Если location есть, но sensorId пустой — подберём его
	if location != "" && sensorID == "" {
		if id, ok := locationToSensorID[location]; ok {
			sensorID = id
		} else {
			sensorID = "0"
		}
	}

	// Если оба пустые — Unknown
	if location == "" && sensorID == "" {
		location = "Unknown"
		sensorID = "0"
	}

	respondWithTemperature(w, location, sensorID)
}

func respondWithTemperature(w http.ResponseWriter, location, sensorID string) {
	// температура от 15 до 25 °C
	temp := 15 + secureRandomFloat()*10

	response := TemperatureResponse{
		Location:    location,
		SensorID:    sensorID,
		Value: math.Round(temp*10) / 10, // округлим до 0.1
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
