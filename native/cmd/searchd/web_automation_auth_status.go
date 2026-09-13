package main

import "net/http"

func (s server) webAutomationAuthStatus(w http.ResponseWriter, _ *http.Request) {
	response := map[string]any{
		"schema_version":              1,
		"configured":                  s.webAutomationAuthConfigured,
		"management_scope":            "development-control-plane",
		"credentials_exposed":          false,
		"profile_ids_exposed":          false,
		"credential_item_ids_exposed": false,
		"production_approved":          false,
	}
	if s.webAutomationAuth != nil {
		response["snapshot"] = s.webAutomationAuth.Snapshot()
	}
	writeAPIV1JSON(w, http.StatusOK, response)
}
