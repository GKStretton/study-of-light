package gooapi

import (
	"encoding/json"
	"net/http"
)

// updateHandler handles incoming state updates from goo
func (g *connectedGooApi) updateHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	defer r.Body.Close()

	if g.stateUpdateCallback == nil {
		httpErr(&w, http.StatusInternalServerError, "state update callback not set")
		return
	}

	_, err := g.verifyInternalRequest(r)
	if err != nil {
		httpErr(&w, http.StatusUnauthorized, "failed to verify internal (goo) request: %v", err)
		return
	}

	// unmarshal data
	var state GooStateUpdate
	err = json.NewDecoder(r.Body).Decode(&state)
	if err != nil {
		httpErr(&w, http.StatusBadRequest, "failed to unmarshal state update: %v", err)
		return
	}

	g.stateUpdateCallback(state)
	l.Debugf("received state update, sent to callback: %+v", state)

	w.WriteHeader(http.StatusOK)
}
