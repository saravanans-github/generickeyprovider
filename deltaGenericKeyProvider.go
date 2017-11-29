package main

import (
	"encoding/xml"
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
const _CPIX_URN = "urn:dashif:org:cpix"
const _PSKC_URN = "urn:ietf:params:xml:ns:keyprov:pskc"
const _SPEKE_URN = "urn:aws:amazon:com:speke"

type SpekeResponseType struct {
	XMLName        xml.Name         `xml:"cpix:CPIX"`
	Id             string           `xml:"id,attr"`
	Cpix           string           `xml:"xmlns:cpix,attr"`
	Pskc           string           `xml:"xmlns:pskc,attr"`
	Speke          string           `xml:"xmlns:speke,attr"`
	ContentKeyList []ContentKeyType `xml:"cpix:ContentKeyList>cpix:ContentKey"`
}

type ContentKeyType struct {
	Kid        string `xml:"kid,attr"`
	ExplicitIV string `xml:"explicitIV,attr"`
	Data       string `xml:"cpix:Data>pskc:Secret>pskc:PlainValue"`
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

		log.Println("Creating Static Speke XML body...")
		response, err := buildStaticSpekeResponse()
		if err != nil {
			log.Panicf("Creating Static Speke XML body... FAILED \n [%s]", err.Error())
		}
		log.Println("Creating Static Speke XML body... DONE")

		log.Println("Writing response body...")
		if _, err := w.Write(response); err != nil {
			log.Panicf("Writing response body... FAILED \n [%s]", err.Error())
		}
		log.Println("Writing response body... DONE")

		next.ServeHTTP(w, r)
	})
}

func buildStaticSpekeResponse() ([]byte, error) {
	//5dGAgwGuUYu4dHeHtNlxJw==
	spekeResponse, err := xml.Marshal(SpekeResponseType{Id: "123", Cpix: "abc", Pskc: "123", Speke: "123",
		ContentKeyList: []ContentKeyType{
			ContentKeyType{
				Kid:        "b4453f69-75ef-415b-9160-1ca699013871",
				ExplicitIV: "5dGAgwGuUYu4dHeHtNlxJw==",
				Data:       "5dGAgwGuUYu4dHeHtNlxJw=="}}})

	if err != nil {
		return nil, err
	}

	return spekeResponse, nil
}
