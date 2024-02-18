package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetTop10MoviesHandler(t *testing.T) {
	// Caso de prueba 1: Verificar que se obtenga una respuesta exitosa (código de estado 200)
	req1, err := http.NewRequest("GET", "/getTop10Movies", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr1 := httptest.NewRecorder()
	handler1 := http.HandlerFunc(getTop10MoviesHandler)
	handler1.ServeHTTP(rr1, req1)
	if status := rr1.Code; status != http.StatusOK {
		t.Errorf("Handler devolvió un código de estado incorrecto: esperado %v pero obtuvo %v",
			http.StatusOK, status)
	}

	// Caso de prueba 2: Verificar que la respuesta es un JSON válido y se puede analizar correctamente
	var response2 []MovieInfo
	err = json.Unmarshal(rr1.Body.Bytes(), &response2)
	if err != nil {
		t.Errorf("Error al analizar la respuesta JSON: %v", err)
	}

	// Caso de prueba 3: Verificar que la longitud de la respuesta sea de 10 o menos
	if len(response2) > 10 {
		t.Errorf("La respuesta debería contener como máximo 10 películas, pero contiene %d", len(response2))
	}

	// Caso de prueba 4: Verificar que el contenido de cada película sea válido (puedes agregar más casos según tu estructura)
	for _, movie := range response2 {
		if movie.Name == "" || movie.Stars == "" || movie.Date == "" || movie.URL == "" {
			t.Errorf("La información de la película no es válida: %+v", movie)
		}
	}
}
