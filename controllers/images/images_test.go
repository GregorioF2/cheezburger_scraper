package images_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	config "propper/configs"
	"strings"
	"testing"

	controller "propper/controllers/images"
	utils "propper/test/utils"
)

var testDirectory = "../../test"
var downloadsDirectory = testDirectory + "/downloads"

func returnImageHandler(w http.ResponseWriter, r *http.Request) {
	fileBytes, err := ioutil.ReadFile(testDirectory + "/data/test_image.jpg")
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

func returnHtmlHandler(content string) http.Handler {
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
	mux.HandleFunc("/download/image", returnImageHandler)

	config.CARD_IMG_SELECTOR = "img"
	config.MIN_CARDS_PER_PAGE = 5
	config.SITE_URL = ts.URL
	config.DOWNLOADS_SAVE_DIR = downloadsDirectory

	return ts, mux
}

func beforeAll() {
	_, err := os.Stat(downloadsDirectory)
	if os.IsNotExist(err) {
		os.Mkdir(downloadsDirectory, 0755)
	}
}

func afterAll() {
	os.RemoveAll(downloadsDirectory)
}

func cleanUpDownloads() {
	os.RemoveAll(downloadsDirectory)
	os.Mkdir(downloadsDirectory, 0755)
}

func checkIfDownloadsAreOk(t *testing.T, numberOfDownloads int) {
	dirs, err := os.ReadDir(downloadsDirectory)
	if err != nil {
		t.Error(err)
		return
	}
	if len(dirs) > 1 {
		t.Error("Downloads directory dirty with more than one download")
		return
	}
	imagesDir := dirs[0].Name()
	files, err := os.ReadDir(fmt.Sprintf("%s/%s", downloadsDirectory, imagesDir))
	if err != nil {
		t.Error(err)
		return
	}
	if !utils.Assert(t, len(files), numberOfDownloads, "Invalid number of downloaded images") {
		return
	}
	filesNames := []string{}
	for _, file := range files {
		filesNames = append(filesNames, file.Name())
	}

	for i := 0; i < numberOfDownloads; i += 1 {
		fileName := fmt.Sprintf("%d.jpg", i+1)
		if !utils.Contains(filesNames, fileName) {
			t.Error("Missing expected download with name: ", fileName)
			return
		}
	}
}

func TestMain(m *testing.M) {
	beforeAll()
	code := m.Run()
	afterAll()
	os.Exit(code)
}

func TestRetriveOneImage(t *testing.T) {
	ts, _ := setupCommonServer()
	defer cleanUpDownloads()
	defer ts.Close()
	ammount := 1
	threads := 1
	_, err := controller.GetImages(ammount, threads)
	if err != nil {
		t.Error("Error getting images: ", err)
	}
	checkIfDownloadsAreOk(t, ammount)
}

func TestRetriveMultipleImages(t *testing.T) {
	ts, _ := setupCommonServer()
	defer cleanUpDownloads()
	defer ts.Close()
	ammount := 100
	threads := 1
	_, err := controller.GetImages(ammount, threads)
	if err != nil {
		t.Error("Error getting images: ", err)
	}
	checkIfDownloadsAreOk(t, ammount)
}

func TestRetriveMultipleImagesWithMultipleThreads(t *testing.T) {
	ts, _ := setupCommonServer()
	defer cleanUpDownloads()
	defer ts.Close()
	ammount := 100
	threads := 2
	_, err := controller.GetImages(ammount, threads)
	if err != nil {
		t.Error("Error getting images: ", err)
	}
	checkIfDownloadsAreOk(t, ammount)
}
