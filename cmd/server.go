package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type CEPJSONResponse struct {
	Localidade string `json:"localidade"`
	Erro       string `json:"erro"`
}

type Current struct {
	TempC float64 `json:"temp_c"`
}

type WeatherAPIResponse struct {
	Current Current `json:"current"`
}

type ServerResponse struct {
	City  string  `json:"city"`
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

func main() {
	client := &http.Client{}
	http.HandleFunc("/",  func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, client)
	})
	http.ListenAndServe(":8181", nil)
}

func handler(w http.ResponseWriter, r *http.Request, client *http.Client) {
	cep := r.URL.Query().Get("cep")

	_, err := strconv.Atoi(cep)
	if len(cep) != 8 || err != nil {
		http.Error(w, "Invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	req, err := client.Get(fmt.Sprintf("http://viacep.com.br/ws/%s/json/", cep))
	if err != nil {
		panic(err)
	}

	if req.StatusCode == http.StatusBadRequest {
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	defer req.Body.Close()

	body, err := io.ReadAll(req.Body)

	if err != nil {
		panic(err)
	}

	var cepResponse CEPJSONResponse
	err = json.Unmarshal(body, &cepResponse)
	if err != nil {
		panic(err)
	}
	if cepResponse.Erro == "true" {
		http.Error(w, "can not find zipcode", http.StatusNotFound)
		return
	}

	req2, err := client.Get(fmt.Sprintf("https://api.weatherapi.com/v1/current.json?q=%s&key=906fd34420aa45c2a20174551251802", cepResponse.Localidade))

	if err != nil {
		panic(err)
	}
	defer req2.Body.Close()

	body, err = io.ReadAll(req2.Body)
	if err != nil {
		panic(err)
	}
	var response WeatherAPIResponse
	err = json.Unmarshal(body, &response)

	if err != nil {
		panic(err)
	}

	serverResponse := ServerResponse{
		City:  cepResponse.Localidade,
		TempC: response.Current.TempC,
		TempF: (response.Current.TempC * 1.8) + 32,
		TempK: response.Current.TempC + 273.15,
	}

	json.NewEncoder(w).Encode(serverResponse)

}
