package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	
	"andrewwormald/apollo/users"
)

type SetRequest struct {
	U users.User
}

type SetResponse struct {
	UserID int64
}

func HandleSet(api users.API) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			_, _ = w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var req SetRequest
		err = json.Unmarshal(b, &req)
		if err != nil {
			_, _ = w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		userID, err := api.Set(r.Context(), req.U)
		if err != nil {
			_, _ = w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var resp SetResponse
		resp.UserID, _ = userID, err
	
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
