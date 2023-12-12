package controllergeo

import (
	"encoding/json"
	"log"
	"net/http"

	"studentgit.kata.academy/Nikolai/selenium/internal/infrastructure/responder"
	"studentgit.kata.academy/Nikolai/selenium/internal/modules/geoservis/repository"
	"studentgit.kata.academy/Nikolai/selenium/internal/modules/geoservis/servis"
)

type GeoServiceController interface {
	SearchAPI(w http.ResponseWriter, r *http.Request)
	GeocodeAPI(w http.ResponseWriter, r *http.Request)
}

type GeoController struct {
	responder.Responder
	dadataService  servis.DadataService
	geoRepo		repository.GeoRepository
}

func NewGeoController(dadataService servis.DadataService,responder responder.Responder, geoRep repository.GeoRepository) *GeoController {

	return &GeoController{dadataService: dadataService, Responder: responder,geoRepo: geoRep}
}



func(gc *GeoController) SearchAPI(w http.ResponseWriter, r *http.Request) {

    var requestBody RequestAddressSearch
	var addresses []Address

    if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		gc.Responder.ErrorBadRequest(w,err)
		log.Println("Decoder Body")
		return
    }

	// Перед отправкой запросов в API приложение должно проверять наличие адреса в базе данных.

    // Проверяем наличие похожих адресов в базе данных
    
    if  OK, err := gc.geoRepo.SearchInHistory(requestBody.Query);OK {

		if err!=nil{
			log.Printf("SearchInHistory %s",err)
		}

		addreses , err := gc.geoRepo.FindAddressByQueryAndHistory(requestBody.Query)
		if err!=nil{
			log.Printf("FindAddressByQueryAndHistory %s",err)
		}

		if len(addreses) > 0{
			address := Address{
				GeoLat: addreses[0].GeoLat,
				GeoLon: addreses[0].GeoLon,
				Result: addreses[0].Region,
			}
			addresses = append(addresses, address)
	
		w.WriteHeader(http.StatusOK)
		gc.OutputJSON(w,ResponseAddress{Addresses: addresses})
		return
		}
    }

	queryID, err := gc.geoRepo.InsertSearchHistory(requestBody.Query)
		if err != nil {
		log.Println(err)
		return
		}

        // Если похожие адреса не найдены, обращаемся к сервису Dadata
		respData, err := gc.dadataService.SearchAddress(requestBody.Query)
		if err != nil {
			gc.Responder.ErrorInternal(w,err)
			log.Println("RespData")
			log.Println(err)
		return
		}

		addresID, err := gc.geoRepo.InsertAddress(respData[0].Region,respData[0].GeoLat,respData[0].GeoLon)
		if err != nil {
		log.Println(err)
		return
	}
	
		for _, adres := range respData {
			address := Address{
				GeoLat: adres.GeoLat,
				GeoLon: adres.GeoLon,
				Result: adres.Region,
			}
			addresses = append(addresses, address)
		}
	
	
		w.WriteHeader(http.StatusOK)
		gc.OutputJSON(w,ResponseAddress{Addresses: addresses})

		err = gc.geoRepo.InsertHistorySearchAddress(queryID,addresID)
		if err != nil {
			log.Println(err)
			return	
		}

}


func(gc *GeoController) GeocodeAPI(w http.ResponseWriter, r *http.Request){

	var requestBody RequestAddressGeocode

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		gc.Responder.ErrorBadRequest(w,err)
		log.Println("Decoder Body")
		return
	}

	respData, err := gc.dadataService.GeocodeAddress(requestBody.Lat,requestBody.Lng)

	if err != nil {
		gc.Responder.ErrorInternal(w,err)
		log.Println("RespData")
		log.Println(err)
		return
	}


	var addresses []Address

	for _, suggestion := range respData.Suggestions {
		address := Address{
			GeoLat: suggestion.Data.GeoLat,
			GeoLon: suggestion.Data.GeoLon,
			Result: suggestion.Data.Result,
		}
		addresses = append(addresses, address)
		break
	}

	if len(addresses) == 0{
		log.Println("len")
		return
	}



	w.WriteHeader(http.StatusOK)
	gc.OutputJSON(w,ResponseAddress{Addresses: addresses})

}