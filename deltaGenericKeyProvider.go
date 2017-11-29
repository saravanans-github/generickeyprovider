package main

import (
	"helper/encode"
	"io/ioutil"
	"log"
	"middleware"
	"net/http"
	"net/http/httputil"
)

const _KEY = "a674a66870be1eba1fee7adb3f3dd37f"
const _IV = "1f408836a6e7c0295bbad005a71a5532"
const _URI = "sdk://test123"
const _SPEKE_UA = "ssaravanan_generickeyprovider"

type SpekeResponseType struct {
	Cpix   string `xml:"license"`
	Status string `json:"status"`
}

func main() {
	startServer()
}

func startServer() {
	resource := []middleware.ResourceType{
		middleware.ResourceType{
			Path:    "/getKeyAndIv",
			Method:  "GET",
			Handler: getKeyAndIv(middleware.IsRequestValid(sendGenericResponse(http.HandlerFunc(final))))},
		middleware.ResourceType{
			Path:    "/speke/v1.0/copyProtection",
			Method:  "POST",
			Handler: getKeyAndIv(middleware.IsRequestValid(sendSpekeResponse(http.HandlerFunc(final))))}}

	config := middleware.ConfigType{Port: 8080, Path: "/delta", Resources: resource}

	middleware.StartServer(config)
}

func final(w http.ResponseWriter, r *http.Request) {
	log.Println("Executing finalHandler")
}

func getKeyAndIv(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// TURN THIS ON/OFF TO ENABLE/DISABLE HTTP DEBUGGING
		dump, err := httputil.DumpRequest(r, true)
		if err != nil {
			log.Fatalln(err)
			message, status := middleware.GetErrorResponse(500, "Server unable to read body.")
			http.Error(w, message, status)
		}

		log.Printf("%q", dump)

		next.ServeHTTP(w, r)
	})
}

func sendGenericResponse(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Writing response headers...")
		log.Println("	ContentType header set")
		w.Header().Set("Content-Type", "application/octet-stream")
		log.Println("	Location header set")
		w.Header().Set("Location", _URI)
		log.Println("Writing response headers... DONE")

		log.Println("Writing response body...")
		if _, err := w.Write(encode.HexStringToBin(_KEY + _IV)); err != nil {
			log.Panicf("Writing response body... FAILED \n [%s]", err.Error())
		}
		log.Println("Writing response body... DONE")
	})
}

func sendSpekeResponse(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
			message, status := middleware.GetErrorResponse(500, "Server unable to read body.")
			http.Error(w, message, status)
		}
		ioutil.NopCloser(r.Body)

		if len(body) == 0 {
			message, status := middleware.GetErrorResponse(404, "Bad request.")
			http.Error(w, message, status)
		}

		log.Println("Writing response headers...")
		log.Println("	ContentType header set")
		w.Header().Set("Content-Type", "application/xml")
		log.Println("	Location header set")
		w.Header().Set("Speke-User-Agent", _SPEKE_UA)
		log.Println("Writing response headers... DONE")

		log.Println("Writing response body...")
		if _, err := w.Write(encode.HexStringToBin(_KEY + _IV)); err != nil {
			log.Panicf("Writing response body... FAILED \n [%s]", err.Error())
		}
		log.Println("Writing response body... DONE")

		next.ServeHTTP(w, r)
	})
}
