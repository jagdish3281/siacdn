package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/thegreatdb/siacdn/backend/db"
	"github.com/thegreatdb/siacdn/backend/models"
	"github.com/thegreatdb/siacdn/backend/randstring"
)

type updateSiaNodeForm struct {
	MinioInstancesRequested int `json:"minio_instances_requested"`
	MinioInstancesActivated int `json:"minio_instances_activated"`
	// TODO: Other things? Maybe even status?
}

func (s *HTTPDServer) handleUpdateSiaNode(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	admin := r.URL.Query().Get("secret") == SiaCDNSecretKey
	var account *models.Account
	var err error
	if !admin {
		req := NewRequest(r, s.db)
		account, err = req.GetAccount()
		if err != nil && err != db.ErrNotFound {
			s.JsonErr(w, err.Error())
			return
		}
		if account == nil || err == db.ErrNotFound {
			s.JsonErr(w, "You must be authenticated to access this resource")
			return
		}
	}

	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		s.JsonErr(w, "Invalid (UUID) SiaNode ID: "+ps.ByName("id"))
		return
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.JsonErr(w, "Could not read data: "+err.Error())
		return
	}
	if err = r.Body.Close(); err != nil {
		s.JsonErr(w, "Could not read data: "+err.Error())
		return
	}

	var form updateSiaNodeForm
	if err = json.Unmarshal(data, &form); err != nil {
		s.JsonErr(w, "Could not decode JSON: "+err.Error())
		return
	}

	sn, err := s.db.GetSiaNode(id)
	if err != nil {
		s.JsonErr(w, "Could not get Sia node: "+err.Error())
		return
	}
	if account != nil && sn.AccountID != account.ID {
		s.JsonErr(w, "You must be authenticated under the correct account to modify this resource")
		return
	}

	if admin {
		sn.MinioInstancesActivated = form.MinioInstancesActivated
	} else {
		sn.MinioInstancesRequested = form.MinioInstancesRequested
	}
	if sn.MinioAccessKey == "" {
		sn.MinioAccessKey = randstring.NewFromUpper(20)
	}
	if sn.MinioSecretKey == "" {
		sn.MinioSecretKey = randstring.NewFromUpper(40)
	}

	if err = s.db.SaveSiaNode(sn); err != nil {
		s.JsonErr(w, "Could not save Sia node: "+err.Error())
		return
	}

	s.Json(w, map[string]interface{}{"sianode": sn})
}
