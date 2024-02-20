package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/handlers"
)

var imdbHost = "https://www.imdb.com"

type MediaInfo struct {
	Name       string `json:"name"`
	Stars      string `json:"stars"`
	Date       string `json:"date"`
	ReviewText string `json:"reviewText"`
	URL        string `json:"url"`
	Usefulness string `json:"usefulness"`
	ImageURL   string `json:"imageUrl"`
}

func extractStarsFromSVG(svg string) string {
	parts := strings.SplitN(svg, ": ", 2)
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

func extractImageURL(html string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		log.Println("Error parsing HTML:", err)
		return ""
	}
	imageURL, exists := doc.Find(".ipc-media img").Attr("src")
	if !exists {
		log.Println("Error: Image URL not found")
		return ""
	}

	return imageURL
}

func getSearchHandler(w http.ResponseWriter, r *http.Request, sortOrder string, mediaType string) {
	queryParams := r.URL.Query()
	limitStr := queryParams.Get("limit")
	language := queryParams.Get("language")

	title := queryParams.Get("title")
	encodedTitle := url.QueryEscape(title)
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	URLSearch := ""
	if mediaType == "video-game" {

		if encodedTitle != "" {
			mediaSearch := "?title_type=video_game&release_date=1900-01-01,2012-01-01&sort=%s&title=%s"
			baseURL := imdbHost + "/search/title/" + mediaSearch
			URLSearch = fmt.Sprintf(baseURL, sortOrder, encodedTitle)
		} else {
			mediaSearch := "?title_type=video_game&release_date=1900-01-01,2012-01-01&sort=%s"
			baseURL := imdbHost + "/search/title/" + mediaSearch
			URLSearch = fmt.Sprintf(baseURL, sortOrder)
		}
		log.Println(URLSearch)
	} else {
		mediaSearch := "?groups=top_1000&view=simple&sort=%s&limit=%s"
		baseURL := imdbHost + "/search/title/" + mediaSearch
		URLSearch = fmt.Sprintf(baseURL, sortOrder, limit)
		log.Println(URLSearch)
	}

	MediaInfos := make([]MediaInfo, 0)

	client := &http.Client{}
	req, err := http.NewRequest("GET", URLSearch, nil)
	if err != nil {
		handleError(w, "Error creating request", err)
		return
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36")
	req.Header.Set("Accept-Language", language+";q=0.5")
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

		imageHTML, _ := s.Find(".ipc-media").Find("img").Parent().Parent().Html()
		imageURL := extractImageURL(imageHTML)

		mediaInfo := MediaInfo{
			Name:       name,
			Stars:      stars,
			Date:       date,
			ReviewText: reviewText,
			URL:        imdbHost + url,
			Usefulness: usefulness,
			ImageURL:   imageURL,
		}

		MediaInfos = append(MediaInfos, mediaInfo)
	})

	var topMovies []MediaInfo
	if len(MediaInfos) >= limit {
		topMovies = MediaInfos[:limit]
	} else {
		topMovies = MediaInfos
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(topMovies)
}

func moviesTop(w http.ResponseWriter, r *http.Request) {
	getSearchHandler(w, r, "user_rating,desc", "video")
}

func moviesFavorites(w http.ResponseWriter, r *http.Request) {
	getSearchHandler(w, r, "popularity,desc", "video")
}
func moviesReleases(w http.ResponseWriter, r *http.Request) {
	getSearchHandler(w, r, "release_date,desc", "video")
}

func moviesCompany(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	company := queryParams.Get("company")
	getSearchHandler(w, r, "num_votes,desc&feature,tv_series,short,tv_movie,tv_miniseries,tv_short,tv_special,tv_episode&companies="+company, "video")
}

// Games
func searchGameByTitle(w http.ResponseWriter, r *http.Request) {
	getSearchHandler(w, r, "moviemeter,asc", "video-game")
}
func searchGameByTop(w http.ResponseWriter, r *http.Request) {
	getSearchHandler(w, r, "user_rating,desc", "video-game")
}
func handleError(w http.ResponseWriter, message string, err error) {
	log.Printf("%s: %v", message, err)
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}
func main() {
	http.HandleFunc("/imdb/movies/top", moviesTop)
	http.HandleFunc("/imdb/movies/favorites", moviesFavorites)
	http.HandleFunc("/imdb/movies/releases", moviesReleases)
	http.HandleFunc("/imdb/movies/company", moviesCompany)

	//Games
	http.HandleFunc("/imdb/games/title", searchGameByTitle)
	http.HandleFunc("/imdb/games/top", searchGameByTop)
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "HEAD", "OPTIONS", "POST", "PUT"}),
		handlers.AllowedHeaders([]string{"Content-Type"}),
	)(http.DefaultServeMux)
	fmt.Println("Server listening on :8080...")
	log.Fatal(http.ListenAndServe(":8080", corsHandler))
}
