package images

import (
	"fmt"
	"net/http"

	imagesController "propper/controllers/images"
	. "propper/types/errors"
)

func GetImages(w http.ResponseWriter, r *http.Request) {
	fmt.Println("On get images")
	err := imagesController.GetImages()
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
