package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// MovieInfo struct representa la información de una reseña de película
type MovieInfo struct {
	Name       string `json:"name"`
	Stars      string `json:"stars"`
	Date       string `json:"date"`
	ReviewText string `json:"reviewText"`
	URL        string `json:"url"`
	Usefulness string `json:"usefulness"`
	ImageURL   string `json:"imageUrl"`
}

func getMoviesHandler(w http.ResponseWriter, r *http.Request, sortOrder string) {
	baseURL := "https://www.imdb.com/search/title/?groups=top_1000&view=simple&sort=%s&limit=10&start=0"
	URLMovies := fmt.Sprintf(baseURL, sortOrder)

	movieInfos := make([]MovieInfo, 0)

	// Obtener información de reseñas para cada película
	client := &http.Client{}
	req, err := http.NewRequest("GET", URLMovies, nil)
	if err != nil {
		handleError(w, "Error creating request", err)
		return
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36")
	response, err := client.Do(req)
	if err != nil {
		handleError(w, "GET failed with error", err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		handleError(w, "GET failed with response code", fmt.Errorf("%v", response.StatusCode))
		return
	}

	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		handleError(w, "Error creating document from response", err)
		return
	}

	doc.Find(".ipc-metadata-list-summary-item").Each(func(i int, s *goquery.Selection) {
		name := s.Find(".ipc-title__text").Text()

		starsStr, exists := s.Find(".ipc-rating-star.ipc-rating-star--imdb.ratingGroup--imdb-rating").Attr("aria-label")
		if !exists {
			log.Println("Error: Stars not found for", name)
			return
		}
		stars := extractStarsFromSVG(starsStr)

		date := s.Find(".dli-title-metadata-item").Eq(0).Text()

		reviewText := s.Find(".ipc-html-content-inner-div").Text()
		url, _ := s.Find(".ipc-lockup-overlay__screen").Parent().Attr("href")
		usefulness := s.Find(".sc-f24f1c5c-7.oCwmv").Text()

		// Obtener la URL de la imagen de la película
		imageHTML, _ := s.Find(".ipc-media").Find("img").Parent().Parent().Html()
		imageURL := extractImageURL(imageHTML)

		// Crear una instancia de MovieInfo y agregarla a la lista
		movieInfo := MovieInfo{
			Name:       name,
			Stars:      stars,
			Date:       date,
			ReviewText: reviewText,
			URL:        "https://www.imdb.com" + url,
			Usefulness: usefulness,
			ImageURL:   imageURL,
		}

		movieInfos = append(movieInfos, movieInfo)
	})

	var top10Movies []MovieInfo
	if len(movieInfos) >= 10 {
		top10Movies = movieInfos[:10]
	} else {
		top10Movies = movieInfos
	}

	// Responder en formato JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(top10Movies)
}

func getTop10MoviesHandler(w http.ResponseWriter, r *http.Request) {
	getMoviesHandler(w, r, "user_rating,desc")
}

func getTop10FavoritesMoviesHandler(w http.ResponseWriter, r *http.Request) {
	getMoviesHandler(w, r, "popularity,asc")
}

func handleError(w http.ResponseWriter, message string, err error) {
	log.Printf("%s: %v", message, err)
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}

func extractStarsFromSVG(svg string) string {
	return strings.Replace(svg, "IMDb rating: ", "", 1)
}

// Función auxiliar para extraer la URL de la imagen de una cadena HTML
func extractImageURL(html string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		log.Println("Error parsing HTML:", err)
		return ""
	}

	// Seleccionar el atributo src de la etiqueta img dentro de la clase .ipc-media
	imageURL, exists := doc.Find(".ipc-media img").Attr("src")
	if !exists {
		log.Println("Error: Image URL not found")
		return ""
	}

	return imageURL
}

func main() {
	// Configurar el manejador para la ruta "/getTop10Movies"
	http.HandleFunc("/getTop10Movies", getTop10MoviesHandler)
	// Configurar el manejador para la ruta "/getTop10OfWeek"
	http.HandleFunc("/getTop10FavoritesMovies", getTop10FavoritesMoviesHandler)

	// Iniciar el servidor en el puerto 8080
	fmt.Println("Server listening on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
