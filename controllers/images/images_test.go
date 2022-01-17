package images_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	config "propper/configs"
	"sort"
	"strings"
	"testing"

	controller "propper/controllers/images"
)

func assertInt(t *testing.T, expected int, actual int, msg string) {
	if expected != actual {
		t.Error(msg, fmt.Sprintf("expected : %v, got : %v", expected, actual))
	}
}

func assertFloat(t *testing.T, expected float64, actual float64, msg string) {
	if expected != actual {
		t.Error(msg, fmt.Sprintf("expected : %v, got : %v", expected, actual))
	}
}

func assertFloats(t *testing.T, expected []float64, actual []float64, msg string) {
	if len(expected) != len(actual) {
		t.Error(msg, fmt.Sprintf("expected : %v, got : %v", expected, actual))
		return
	}
	sort.Float64s(expected)
	sort.Float64s(actual)
	for i, expectedValue := range expected {
		if expectedValue != actual[i] {
			t.Error(msg, fmt.Sprintf("expected : %v, got : %v", expected, actual))
			return
		}
	}
}

func handleReturnImage(w http.ResponseWriter, r *http.Request) {
	fileBytes, err := ioutil.ReadFile("testdata/test_image.jpg")
	if err != nil {
		panic(err)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(fileBytes)
	return
}

func testHtml(imagesNumber int, url string) string {
	res := `
	<body>
	<p id="content" onclick="changeText()">Original content.</p>`
	for i := 0; i < imagesNumber; i += 1 {
		res += fmt.Sprintf(`
		<img src="%s">`, url)
	}
	res += `</body>
	`
	return res
}

func writeHTML(content string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		io.WriteString(w, strings.TrimSpace(content))
	})
}

func setupCommonServer() (*httptest.Server, *http.ServeMux) {
	mux := http.NewServeMux()
	ts := httptest.NewServer(mux)
	url := fmt.Sprintf("%s/download/image", ts.URL)
	mux.HandleFunc("/page/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		io.WriteString(w, strings.TrimSpace(testHtml(5, url)))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		io.WriteString(w, strings.TrimSpace(testHtml(5, url)))
	})
	mux.HandleFunc("/download/image", handleReturnImage)

	config.CARD_IMG_SELECTOR = "img"
	config.MIN_CARDS_PER_PAGE = 5
	config.SITE_URL = ts.URL
	config.DOWNLOADS_SAVE_DIR = "./testdata/downloads"

	return ts, mux
}

func beforeAll() {
	_, err := os.Stat("testdata/downloads")
	if os.IsNotExist(err) {
		os.Mkdir("testdata/downloads", 0755)
	}
}

func afterAll() {
	os.RemoveAll("testdata/downloads")
}

func cleanUpDownloads() {
	fmt.Println("Clean up")
	os.RemoveAll("testdata/downloads")
	os.Mkdir("testdata/downloads", 0755)
}

func TestMain(m *testing.M) {
	beforeAll()
	code := m.Run()
	afterAll()
	os.Exit(code)
}

func TestRetriveOneImage(t *testing.T) {
	ts, _ := setupCommonServer()
	cleanUpDownloads()
	defer ts.Close()
	err := controller.GetImages(1, 1)
	if err != nil {
		t.Error("Error getting images: ", err)
	}
}
