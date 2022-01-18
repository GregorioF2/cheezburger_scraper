package images

import (
	"encoding/json"
	"net/http"
	"strconv"

	imagesController "propper/controllers/images"
	. "propper/types/errors"
)

func getImagesParameters(parameters map[string][]string) (int, int, error) {
	var err error
	var amount uint64 = 10
	paramAmount, ok := parameters["amount"]

	if ok {
		amount, err = strconv.ParseUint(paramAmount[0], 10, 32)
		if err != nil {
			return 0, 0, &InvalidParametersError{Err: "Error reading 'amount' parameter: " + err.Error()}
		}
	}

	var threads uint64 = 1
	paramThreads, ok := parameters["threads"]
	if ok {
		threads, err = strconv.ParseUint(paramThreads[0], 10, 32)
		if err != nil {
			return 0, 0, &InvalidParametersError{Err: "Error reading 'threads' parameter: " + err.Error()}
		}
	}
	return int(amount), int(threads), nil
}

func GetImages(w http.ResponseWriter, r *http.Request) {
	var err error
	amount, threads, err := getImagesParameters(r.URL.Query())
	if err != nil {
		var responseError *ResponseError
		switch e := err.(type) {
		case *InvalidParametersError:
			responseError = &ResponseError{Err: e.Error(), StatusCode: http.StatusBadRequest}
		default:
			responseError = &ResponseError{Err: e.Error(), StatusCode: http.StatusInternalServerError}
		}
		http.Error(w, responseError.Error(), responseError.StatusCode)
		return
	}
	urls, err := imagesController.GetImages(amount, threads)
	if err != nil {
		var responseError *ResponseError
		switch e := err.(type) {
		case *InvalidParametersError:
			responseError = &ResponseError{Err: e.Error(), StatusCode: http.StatusBadRequest}
		case *NotFoundError:
			responseError = &ResponseError{Err: e.Error(), StatusCode: http.StatusNotFound}
		default:
			responseError = &ResponseError{Err: e.Error(), StatusCode: http.StatusInternalServerError}
		}
		http.Error(w, responseError.Error(), responseError.StatusCode)
		return
	}

	payload, err := json.Marshal(urls)
	if err != nil {
		responseError := &ResponseError{Err: "error encoding return payload", StatusCode: http.StatusInternalServerError}
		http.Error(w, responseError.Error(), responseError.StatusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(payload)
}
