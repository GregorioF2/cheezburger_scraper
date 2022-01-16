package images

import (
	"net/http"
	"strconv"

	imagesController "propper/controllers/images"
	. "propper/types/errors"
)

func getImagesParameters(parameters map[string][]string) (int, error) {
	var err error
	var ammount uint64 = 10
	paramAmmount, ok := parameters["ammount"]

	if ok {
		ammount, err = strconv.ParseUint(paramAmmount[0], 10, 32)
		if err != nil {
			return 0, &InvalidParametersError{Err: "Error reading ammount parameter: " + err.Error()}
		}
	}
	if ammount == 0 {
		return 0, &InvalidParametersError{Err: "Error ammount parameter must be greater than 0: " + err.Error()}
	}
	return int(ammount), nil
}

func GetImages(w http.ResponseWriter, r *http.Request) {
	ammount, err := getImagesParameters(r.URL.Query())
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
	err = imagesController.GetImages(ammount)
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

	w.WriteHeader(http.StatusNoContent)
}
