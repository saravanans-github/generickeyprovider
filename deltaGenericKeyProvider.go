package main

import (
	"encoding/xml"
	"helper/encode"
	"io/ioutil"
	"log"
	"middleware"
	"net/http"
)

const _KEY = "a674a66870be1eba1fee7adb3f3dd37f"
const _IV = "1f408836a6e7c0295bbad005a71a5532"
const _URI = "sdk://test123"
const _SPEKE_UA = "ssaravanan_generickeyprovider"
const _CPIX_URN = "urn:dashif:org:cpix"
const _PSKC_URN = "urn:ietf:params:xml:ns:keyprov:pskc"
const _SPEKE_URN = "urn:aws:amazon:com:speke"
const _FAIRPLAY_URIEXTXKEY = "skd://thisIsNotNeededForFairPlay"
const _FAIRPLAY_KEYFORMAT = "com.apple.streamingkeydelivery"
const _FAIRPLAY_KEYFORMATVERSIONS = "1"

type CpixRequestType struct {
	XMLName        xml.Name         `xml:"CPIX"`
	Id             string           `xml:"id,attr"`
	Cpix           string           `xml:"xmlns:cpix,attr"`
	Pskc           string           `xml:"xmlns:pskc,attr"`
	Speke          string           `xml:"xmlns:speke,attr"`
	ContentKeyList []ContentKeyType `xml:"ContentKeyList>ContentKey"`
	DRMSystemList  []DRMSystemType  `xml:"DRMSystemList>DRMSystem,omitempty"`
}

type CpixResponseType struct {
	XMLName        xml.Name         `xml:"cpix:CPIX"`
	Id             string           `xml:"id,attr"`
	Cpix           string           `xml:"xmlns:cpix,attr"`
	Pskc           string           `xml:"xmlns:pskc,attr"`
	Speke          string           `xml:"xmlns:speke,attr"`
	ContentKeyList []ContentKeyType `xml:"cpix:ContentKeyList>cpix:ContentKey"`
	DRMSystemList  []DRMSystemType  `xml:"cpix:DRMSystemList>cpix:DRMSystem,omitempty"`
}

type ContentKeyType struct {
	Kid        string `xml:"kid,attr"`
	ExplicitIV string `xml:"explicitIV,attr"`
	Data       string `xml:"cpix:Data>pskc:Secret>pskc:PlainValue"`
}

type DRMSystemType struct {
	Kid               string `xml:"kid,attr"`
	SystemId          string `xml:"systemId,attr"`
	URIExtXKey        string `xml:"cpix:URIExtXKey,omitempty"`
	KeyFormat         string `xml:"speke:KeyFormat,omitempty"`
	KeyFormatVersions string `xml:"speke:KeyFormatVersions,omitempty"`
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
		/*
			dump, err := httputil.DumpRequest(r, true)
			if err != nil {
				log.Fatalln(err)
				message, status := middleware.GetErrorResponse(500, "Server unable to read body.")
				http.Error(w, message, status)
			}

			log.Printf("%q", dump)
		*/

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

		log.Println(string(body))
		log.Println("Marshalling request into XML object...")
		var requestInXML CpixRequestType
		err = xml.Unmarshal(body, &requestInXML)
		if err != nil {
			log.Fatalf("Marshalling request into XML object... FAILED [%s]", err.Error())
		}
		log.Println("Marshalling request into XML object... DONE")

		log.Println("Writing response headers...")
		log.Println("	ContentType header set")
		w.Header().Set("Content-Type", "application/xml")
		log.Println("	Location header set")
		w.Header().Set("Speke-User-Agent", _SPEKE_UA)
		log.Println("Writing response headers... DONE")

		log.Println("Creating Static Speke XML body...")
		response, err := buildStaticSpekeResponse(requestInXML.Id, requestInXML.ContentKeyList, requestInXML.DRMSystemList)
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

func buildStaticSpekeResponse(id string, contentKeys []ContentKeyType, drmSystems []DRMSystemType) ([]byte, error) {

	var resContentKeys = make([]ContentKeyType, len(contentKeys))

	// Set the same static key & iv for each kid in the request
	// Ideally we will want to create a different key and iv for every different kid
	log.Printf("length of content keys %d", len(contentKeys))
	for i := 0; i < len(contentKeys); i++ {
		resContentKeys[i].Kid = contentKeys[i].Kid
		resContentKeys[i].Data = encode.BytesToBase64(encode.HexStringToBin(_KEY))
		resContentKeys[i].ExplicitIV = encode.BytesToBase64(encode.HexStringToBin(_IV))
	}

	// Now we set DRM specific data statically
	// Ideally we'll want to pull this of a config
	var resDrmSystems = make([]DRMSystemType, len(drmSystems))

	log.Printf("length of drm systems %d", len(drmSystems))
	// Set the same static key & iv for each kid in the request
	// Ideally we will want to create a different key and iv for every different kid
	for i := 0; i < len(drmSystems); i++ {
		resDrmSystems[i].Kid = drmSystems[i].Kid
		resDrmSystems[i].SystemId = drmSystems[i].SystemId
		resDrmSystems[i].URIExtXKey = encode.BytesToBase64([]byte(_FAIRPLAY_URIEXTXKEY))
		resDrmSystems[i].KeyFormat = encode.BytesToBase64([]byte(_FAIRPLAY_KEYFORMAT))
		resDrmSystems[i].KeyFormatVersions = encode.BytesToBase64([]byte(_FAIRPLAY_KEYFORMATVERSIONS))
	}

	spekeResponse, err := xml.Marshal(CpixResponseType{Id: id, Cpix: _CPIX_URN, Pskc: _PSKC_URN, Speke: _SPEKE_URN,
		ContentKeyList: resContentKeys,
		DRMSystemList:  resDrmSystems})

	if err != nil {
		return nil, err
	}

	return spekeResponse, nil
}
