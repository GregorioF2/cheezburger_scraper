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

	errors "propper/types/errors"
)

var testDirectory = "../../test"
var downloadsDirectory = testDirectory + "/downloads"

func testHtml(imagesNumber int, url string) string {
	res := `
	<body>
	<p id="content"">Original content.</p>`
	for i := 0; i < imagesNumber; i += 1 {
		res += fmt.Sprintf(`
		<img src="%s">`, url)
	}
	res += `</body>
	`
	return res
}

func imageHandler(w http.ResponseWriter, r *http.Request) {
	fileBytes, err := ioutil.ReadFile(testDirectory + "/data/test_image.jpg")
	if err != nil {
		panic(err)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(fileBytes)
	return
}

func returnHtmlHandler(content string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		io.WriteString(w, strings.TrimSpace(content))
	}
}

func setupServerWithBlankBody() (*httptest.Server, *http.ServeMux) {
	mux := http.NewServeMux()
	ts := httptest.NewServer(mux)
	url := fmt.Sprintf("%s/download/image", ts.URL)
	mux.HandleFunc("/page/{id}", returnHtmlHandler(testHtml(0, url)))
	mux.HandleFunc("/", returnHtmlHandler(testHtml(0, url)))
	mux.HandleFunc("/download/image", imageHandler)

	config.CARD_IMG_SELECTOR = "img"
	config.MIN_CARDS_PER_PAGE = 5
	config.SITE_URL = ts.URL
	config.DOWNLOADS_SAVE_DIR = downloadsDirectory
	config.SLEEP_TIME = 0

	return ts, mux
}

func setupServerErrorImgSrc() (*httptest.Server, *http.ServeMux) {
	mux := http.NewServeMux()
	ts := httptest.NewServer(mux)
	mux.HandleFunc("/page/{id}", returnHtmlHandler(testHtml(5, "invalid imag src")))
	mux.HandleFunc("/", returnHtmlHandler(testHtml(5, "invalid imag src")))
	mux.HandleFunc("/download/image", imageHandler)

	config.CARD_IMG_SELECTOR = "img"
	config.MIN_CARDS_PER_PAGE = 5
	config.SITE_URL = ts.URL
	config.DOWNLOADS_SAVE_DIR = downloadsDirectory
	config.SLEEP_TIME = 0

	return ts, mux
}

func setupCommonServer() (*httptest.Server, *http.ServeMux) {
	mux := http.NewServeMux()
	ts := httptest.NewServer(mux)
	url := fmt.Sprintf("%s/download/image", ts.URL)
	mux.HandleFunc("/page/{id}", returnHtmlHandler(testHtml(5, url)))
	mux.HandleFunc("/", returnHtmlHandler(testHtml(5, url)))
	mux.HandleFunc("/download/image", imageHandler)

	config.CARD_IMG_SELECTOR = "img"
	config.MIN_CARDS_PER_PAGE = 5
	config.SITE_URL = ts.URL
	config.DOWNLOADS_SAVE_DIR = downloadsDirectory
	config.SLEEP_TIME = 0

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
	ammount := 20
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
	ammount := 20
	threads := 2
	_, err := controller.GetImages(ammount, threads)
	if err != nil {
		t.Error("Error getting images: ", err)
	}
	checkIfDownloadsAreOk(t, ammount)
}

func TestErrorOnInvalidSiteURL(t *testing.T) {
	ts, _ := setupCommonServer()
	config.SITE_URL = "invalid site url"
	defer cleanUpDownloads()
	defer ts.Close()
	ammount := 1
	threads := 1
	_, err := controller.GetImages(ammount, threads)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	switch e := err.(type) {
	case *errors.ConnectionError:
		return
	default:
		t.Error("Expected error has invalid type. ConnectionError was expected. Error recived: ", e.Error())
	}
}

func TestErrorOnInvalidSaveDirectory(t *testing.T) {
	ts, _ := setupCommonServer()
	config.DOWNLOADS_SAVE_DIR = "invalid save directory selector"
	defer cleanUpDownloads()
	defer ts.Close()
	ammount := 1
	threads := 1
	_, err := controller.GetImages(ammount, threads)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	switch e := err.(type) {
	case *errors.InternalServerError:
		return
	default:
		t.Error("Expected error has invalid type. InternalServerError was expected. Error recived: ", e.Error())
	}
}

func TestErrorOnInvalidImageSrc(t *testing.T) {
	ts, _ := setupServerErrorImgSrc()
	defer cleanUpDownloads()
	defer ts.Close()
	ammount := 1
	threads := 1
	_, err := controller.GetImages(ammount, threads)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	switch e := err.(type) {
	case *errors.ConnectionError:
		return
	default:
		t.Error("Expected error has invalid type. ConnectionError was expected. Error recived: ", e.Error())
	}
}

func TestErrorOnBodyWithNoImages(t *testing.T) {
	ts, _ := setupServerWithBlankBody()
	defer cleanUpDownloads()
	defer ts.Close()
	ammount := 1
	threads := 1
	_, err := controller.GetImages(ammount, threads)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	switch e := err.(type) {
	case *errors.NotFoundError:
		return
	default:
		t.Error("Expected error has invalid type. NotFoundError was expected. Error recived: ", e.Error())
	}
}
