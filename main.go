package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"

	config "propper/configs"
	middlewares "propper/middlewares"
	imagesRoutes "propper/routes/images"

	"github.com/gorilla/mux"
)

func reportStatus(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func genericHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func handleRequest() {
	mainRouter := mux.NewRouter().StrictSlash(true)
	mainRouter.HandleFunc("/status", reportStatus)

	imagesSubRoute := mainRouter.PathPrefix("/images").Subrouter()
	imagesSubRoute.Use(middlewares.SetCorsHeaders)
	imagesSubRoute.HandleFunc("/download", imagesRoutes.GetImages)
	imagesSubRoute.HandleFunc("/", genericHandler)
	imagesSubRoute.HandleFunc("/{.*}", genericHandler)

	fmt.Println("Running on " + config.PORT)
	log.Fatal(http.ListenAndServe(":"+config.PORT, mainRouter))

}
func main() {
	cores := runtime.NumCPU()
	runtime.GOMAXPROCS(cores)
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	handleRequest()
}
