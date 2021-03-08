package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	
	"andrewwormald/apollo/bookings"
)

type PingRequest struct {}

type PingResponse struct {}

func HandlePing(api bookings.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			_, _ = w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var req PingRequest
		err = json.Unmarshal(b, &req)
		if err != nil {
			_, _ = w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = api.Ping(r.Context())
		if err != nil {
			_, _ = w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var resp PingResponse
		_ = err
	
		respBody, err := json.Marshal(resp)
		if err != nil {
			_, _ = w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		
		_, err = w.Write(respBody)
		if err != nil {
			_, _ = w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func RegisterHandlers(api bookings.Client) {
	http.HandleFunc("/bookings/ping", HandlePing(api))
}
