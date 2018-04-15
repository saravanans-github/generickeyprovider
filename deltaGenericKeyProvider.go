package main

import (
	pb "WidevineCencHeader"
	"encoding/xml"
	"helper/encode"
	"io/ioutil"
	"log"
	"middleware"
	"net/http"

	"github.com/golang/protobuf/proto"
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

const _FAIRPLAY_SYSTEM_ID = "98ee5596-cd3e-a20d-163a-e382420c6eff"
const _WIDEVINE_SYSTEM_ID = "edef8ba9-79d6-4ace-a3c8-27dcd51d21ed"
const _PLAYREADY_SYSTEM_ID = "9a04f079-9840-4286-ab92-e65be0885f95"
const _CENC_SYSTEM_ID = ""

const _WIDEVINE_PROVIDER = "widevine_test"
const _WIDEVINE_TRACKTYPE = "SD"

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
	Pssh              string `xml:"cpix:PSSH,omitempty"`
	ProtectionHeader  string `xml:"speke:ProtectionHeader,omitempty"`
}

type empty struct{}

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
		defer next.ServeHTTP(w, r)

		log.Println("Reading request body...")
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("Reading request body... FAILED [%s]", err.Error())
			message, status := middleware.GetErrorResponse(500, "Server unable to read body. "+err.Error())
			http.Error(w, message, status)
		}
		ioutil.NopCloser(r.Body)

		if len(body) == 0 {
			message, status := middleware.GetErrorResponse(400, "Bad request. Body is empty.")
			http.Error(w, message, status)
			return
		}

		log.Println("Marshalling request into XML object...")
		var requestInXML CpixRequestType
		err = xml.Unmarshal(body, &requestInXML)
		if err != nil {
			log.Printf("Marshalling request into XML object... FAILED [%s]", err.Error())
			message, status := middleware.GetErrorResponse(400, "Bad request. "+err.Error())
			http.Error(w, message, status)
			return
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
			log.Printf("Creating Static Speke XML body... FAILED \n [%s]", err.Error())
			message, status := middleware.GetErrorResponse(400, "Bad request. "+err.Error())
			http.Error(w, message, status)
			return
		}
		log.Println("Creating Static Speke XML body... DONE")

		log.Println("Writing response body...")
		if _, err := w.Write(response); err != nil {
			log.Printf("Writing response body... FAILED \n [%s]", err.Error())
			message, status := middleware.GetErrorResponse(400, "Bad request. "+err.Error())
			http.Error(w, message, status)
			return
		}
		log.Println("Writing response body... DONE")
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
	len := len(drmSystems)
	resDrmSystems := make([]DRMSystemType, len)
	sem := make(chan empty, len) // semaphore pattern

	// Here we use the semaphore pattern to parallelize the response for each drm system
	for i, drmSystem := range drmSystems {
		go func(i int, drmSystem DRMSystemType) {
			log.Println(drmSystem.SystemId)
			switch drmSystem.SystemId {
			case _FAIRPLAY_SYSTEM_ID:
				resDrmSystems[i].URIExtXKey = encode.BytesToBase64([]byte(_FAIRPLAY_URIEXTXKEY))
				resDrmSystems[i].KeyFormat = encode.BytesToBase64([]byte(_FAIRPLAY_KEYFORMAT))
				resDrmSystems[i].KeyFormatVersions = encode.BytesToBase64([]byte(_FAIRPLAY_KEYFORMATVERSIONS))
				break
			case _WIDEVINE_SYSTEM_ID:
				// TODO: implement proper (HTTP) error handling
				// data, err := generateProtoBuf([]byte(drmSystem.Kid), []byte(id), _WIDEVINE_PROVIDER, _WIDEVINE_TRACKTYPE)
				// if err != nil {
				// 	log.Fatal(err)
				// }

				data := []byte{
					0x00, 0x00, 0x00, 0x44, 0x70, 0x73, 0x73, 0x68, // BMFF box header (68 bytes, 'pssh')
					0x01, 0x00, 0x00, 0x00, // Full box header (version = 1, flags = 0)
					// 0x10, 0x77, 0xef, 0xec, 0xc0, 0xb2, 0x4d, 0x02, // SystemID
					// 0xac, 0xe3, 0x3c, 0x1e, 0x52, 0xe2, 0xfb, 0x4b,
					0xed, 0xef, 0x8b, 0xa9, 0x79, 0xd6, 0x4a, 0xce,
					0xa3, 0xc8, 0x27, 0xdc, 0xd5, 0x1d, 0x21, 0xed,
					0x00, 0x00, 0x00, 0x01, // KID_count (1)
					0x00, 0x00, 0xbe, 0xcb, 0xba, 0xdd, 0x44, 0x26,
					0x11, 0x56, 0x8e, 0x4e, 0xfd, 0x58, 0x27, 0x70,
					0x00, 0x00, 0x00, 0x00} // Size of Data (0)

				resDrmSystems[i].Pssh = encode.BytesToBase64(data)
				break
			}
			resDrmSystems[i].Kid = drmSystems[i].Kid
			resDrmSystems[i].SystemId = drmSystems[i].SystemId

			sem <- empty{}
		}(i, drmSystem)
	}
	// wait for goroutines to finish
	for i := 0; i < len; i++ {
		<-sem
	}

	spekeResponse, err := xml.Marshal(CpixResponseType{Id: id, Cpix: _CPIX_URN, Pskc: _PSKC_URN, Speke: _SPEKE_URN,
		ContentKeyList: resContentKeys,
		DRMSystemList:  resDrmSystems})

	if err != nil {
		return nil, err
	}

	return spekeResponse, nil
}

func generateProtoBuf(keyId []byte, contentId []byte, provider string, trackType string) ([]byte, error) {

	key_id := [][]byte{keyId}
	algorithm_value := pb.WidevineCencHeader_AESCTR
	policy := string("")

	pssh := &pb.WidevineCencHeader{
		Algorithm:           &algorithm_value,
		KeyId:               key_id,
		Provider:            &provider,
		ContentId:           contentId,
		TrackTypeDeprecated: &trackType,
		Policy:              &policy}

	data, err := proto.Marshal(pssh)
	if err != nil {
		log.Fatal("marshaling error: ", err)

		return nil, err
	}

	return data, nil
}
