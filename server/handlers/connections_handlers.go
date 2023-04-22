package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/layer5io/meshery/server/models"
)

func (h *Handler) SaveConnection(w http.ResponseWriter, req *http.Request, _ *models.Preference, _ *models.User, provider models.Provider) {
	bd, err := io.ReadAll(req.Body)
	if err != nil {
		h.log.Error(fmt.Errorf("error reading request body: %v", err))
		http.Error(w, "unable to read result data", http.StatusInternalServerError)
		return
	}

	connection := models.ConnectionPayload{}
	err = json.Unmarshal(bd, &connection)
	if err != nil {
		h.log.Error(fmt.Errorf("error unmarshal request body: %v", err))
		http.Error(w, "unable to parse connection data", http.StatusInternalServerError)
		return
	}

	err = provider.SaveConnection(req, &connection, "", false)
	if err != nil {
		h.log.Error(fmt.Errorf("error saving connection: %v", err))
		http.Error(w, "unable to save connection", http.StatusInternalServerError)
		return
	}

	h.log.Info("connection saved successfully")
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) GetConnections(w http.ResponseWriter, req *http.Request, _ *models.Preference, user *models.User, provider models.Provider) {
	q := req.URL.Query()
	pathSegments := strings.Split(req.URL.Path, "/")
	connectionKind := pathSegments[len(pathSegments)-1]
	page, _ := strconv.Atoi(q.Get("page"))
	order := q.Get("order")
	search := q.Get("search")
	pageSize, _ := strconv.Atoi(q.Get("page_size"))

	if pageSize > 25 {
		pageSize = 25
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if page < 0 {
		page = 0
	}
	if order == "" {
		order = "created_at desc"
	}

	h.log.Debug(fmt.Sprintf("page: %d, page size: %d, search: %s, order: %s", page+1, pageSize, search, order))

	connectionsPage, err := provider.GetConnections(req, user.ID, page, pageSize, order, search, connectionKind)
	if err != nil {
		h.log.Error(fmt.Errorf("error getting user connections: %v", err))
		http.Error(w, "unable to get user connections", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(connectionsPage); err != nil {
		h.log.Error(fmt.Errorf("error encoding user connections: %v", err))
		http.Error(w, "unable to encode user connections", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) UpdateConnection(w http.ResponseWriter, req *http.Request, _ *models.Preference, _ *models.User, provider models.Provider) {
	bd, err := io.ReadAll(req.Body)
	if err != nil {
		h.log.Error(fmt.Errorf("error reading request body: %v", err))
		http.Error(w, "unable to read connection data", http.StatusInternalServerError)
		return
	}

	connection := &models.Connection{}
	err = json.Unmarshal(bd, connection)
	if err != nil {
		h.log.Error(fmt.Errorf("error unmarshal request body: %v", err))
		http.Error(w, "unable to parse connection data", http.StatusInternalServerError)
		return
	}

	_, err = provider.UpdateConnection(req, connection)
	if err != nil {
		h.log.Error(fmt.Errorf("error getting user connection: %v", err))
		http.Error(w, "unable to get user connection", http.StatusInternalServerError)
		return
	}

	h.log.Info("connection updated successfully")
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) DeleteConnection(w http.ResponseWriter, req *http.Request, _ *models.Preference, _ *models.User, provider models.Provider) {
	q := req.URL.Query()

	connectionID := uuid.FromStringOrNil(q.Get("connection_id"))
	_, err := provider.DeleteConnection(req, connectionID)
	if err != nil {
		h.log.Error(fmt.Errorf("error deleting user connection: %v", err))
		http.Error(w, "unable to delete user connection", http.StatusInternalServerError)
		return
	}

	h.log.Info("connection deleted successfully")
	w.WriteHeader(http.StatusOK)
}
