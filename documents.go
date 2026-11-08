// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-09-15):
// CreateFolder returned *DocumentFolder, but the real endpoint
// (POST /api/v1/storage/fixed/folders) returns only {folderId}, leaving
// every field zero-valued on unmarshal. Called with {ownerType, ownerId} it
// DID create the folder, so its v1 signature
// CreateFolder(map[string]interface{}) is kept with the corrected
// *CreateFolderResult; CreateFolderForOwner(ownerType, ownerID) is the typed
// form. ListFolders called GET /folders with no params; there is no
// list-all-folders route (GET /folders requires ?ownerType=&ownerId= and
// returns ONE folder), so every call 400'd. It is kept as a Deprecated
// method returning ErrListFoldersUnsupported without sending a request —
// deleting an exported method would break v1 consumers' compilation — and
// GetFolderByOwner(ownerType, ownerId) is the replacement.
// RequestOverride/ProcessOverride kept their untyped-map signatures (so
// existing callers who pass the correct keys are unaffected) but their
// *DeletionOverride return type is corrected to match the real response
// (requestedById/overrideReason/etc., not requestedBy/reason/reviewedBy/
// reviewNote, which don't exist in the real API). ProcessOverride's
// "execute" action returns a DIFFERENT shape ({success, documentId}, not a
// DeletionOverride) — use the new ExecuteOverride for that action instead
// of ProcessOverride, which cannot represent that response correctly.
// =============================================================================

package platformxe

import (
	"errors"
	"fmt"
)

// DocumentsService handles fixed-storage documents, folders, and deletion overrides.
type DocumentsService struct {
	client *Client
}

// ListDocuments returns documents with optional filters.
func (s *DocumentsService) ListDocuments(params map[string]string) (*DocumentListResponse, error) {
	var result DocumentListResponse
	err := s.client.doRequestTyped("GET", "/api/v1/storage/fixed/documents", nil, params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetDocument returns a document by ID.
func (s *DocumentsService) GetDocument(id string) (*Document, error) {
	var result Document
	err := s.client.doRequestTyped("GET", fmt.Sprintf("/api/v1/storage/fixed/documents/%s", id), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateDocument creates a new fixed-storage document.
func (s *DocumentsService) CreateDocument(input map[string]interface{}) (*Document, error) {
	var result Document
	err := s.client.doRequestTyped("POST", "/api/v1/storage/fixed/documents", input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateDocument updates a fixed-storage document.
func (s *DocumentsService) UpdateDocument(documentID string, input map[string]interface{}) (*Document, error) {
	var result Document
	err := s.client.doRequestTyped("PATCH", fmt.Sprintf("/api/v1/storage/fixed/documents/%s", documentID), input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteDocument deletes a fixed-storage document.
func (s *DocumentsService) DeleteDocument(documentID string) (*DeleteResult, error) {
	var result DeleteResult
	err := s.client.doRequestTyped("DELETE", fmt.Sprintf("/api/v1/storage/fixed/documents/%s", documentID), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetFolderByOwner returns the folder for a given owner — one folder per
// (ownerType, ownerId) pair. ownerType must be one of "ADMIN", "AGENT",
// "PARTNER", "SYSTEM".
func (s *DocumentsService) GetFolderByOwner(ownerType, ownerID string) (*DocumentFolder, error) {
	var result DocumentFolder
	params := map[string]string{"ownerType": ownerType, "ownerId": ownerID}
	err := s.client.doRequestTyped("GET", "/api/v1/storage/fixed/folders", nil, params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetFolder returns a folder by ID.
func (s *DocumentsService) GetFolder(id string) (*DocumentFolder, error) {
	var result DocumentFolder
	err := s.client.doRequestTyped("GET", fmt.Sprintf("/api/v1/storage/fixed/folders/%s", id), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ErrListFoldersUnsupported is returned by the deprecated ListFolders.
var ErrListFoldersUnsupported = errors.New("platformxe: ListFolders is not supported by the API — use GetFolderByOwner")

// ListFolders is kept for v1 compatibility only. The API has no
// list-folders route (GET /api/v1/storage/fixed/folders requires
// ownerType+ownerId and returns ONE folder), so this never sends a request
// and always returns ErrListFoldersUnsupported.
//
// Deprecated: use GetFolderByOwner.
func (s *DocumentsService) ListFolders() (*FolderListResponse, error) {
	return nil, ErrListFoldersUnsupported
}

// CreateFolder creates (or idempotently returns) a storage folder for an
// owner. body must include "ownerType" (one of "ADMIN", "AGENT", "PARTNER",
// "SYSTEM") and "ownerId". Prefer the typed CreateFolderForOwner.
func (s *DocumentsService) CreateFolder(body map[string]interface{}) (*CreateFolderResult, error) {
	var result CreateFolderResult
	err := s.client.doRequestTyped("POST", "/api/v1/storage/fixed/folders", body, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateFolderForOwner creates (or idempotently returns) the storage folder
// for (ownerType, ownerID). ownerType must be one of "ADMIN", "AGENT",
// "PARTNER", "SYSTEM".
func (s *DocumentsService) CreateFolderForOwner(ownerType, ownerID string) (*CreateFolderResult, error) {
	return s.CreateFolder(map[string]interface{}{"ownerType": ownerType, "ownerId": ownerID})
}

// RequestOverride requests a deletion override for a protected document.
// input must include "documentId" and "overrideReason" (string, minimum
// 500 characters); optional keys: "authorityReference", "authorityType",
// "authorityDocument", "operatorId".
func (s *DocumentsService) RequestOverride(input map[string]interface{}) (*DeletionOverride, error) {
	var result DeletionOverride
	err := s.client.doRequestTyped("POST", "/api/v1/storage/fixed/overrides", input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ProcessOverride approves, rejects, or witnesses a deletion override
// request. input must include "action" ("approve", "reject", or "witness"
// — for "execute", use ExecuteOverride instead, which returns the correct
// response shape); "rejectionReason" is required when action is "reject".
// Optional key: "operatorId".
func (s *DocumentsService) ProcessOverride(overrideID string, input map[string]interface{}) (*DeletionOverride, error) {
	var result DeletionOverride
	err := s.client.doRequestTyped("PATCH", fmt.Sprintf("/api/v1/storage/fixed/overrides/%s", overrideID), input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ExecuteOverride executes an approved deletion override, soft-deleting the
// document as part of this call — there is no need to retry
// DeleteDocument() afterward. Returns a shape distinct from
// DeletionOverride's fields; operatorID may be empty.
func (s *DocumentsService) ExecuteOverride(overrideID, operatorID string) (*ExecuteOverrideResult, error) {
	body := map[string]interface{}{"action": "execute"}
	if operatorID != "" {
		body["operatorId"] = operatorID
	}
	var result ExecuteOverrideResult
	err := s.client.doRequestTyped("PATCH", fmt.Sprintf("/api/v1/storage/fixed/overrides/%s", overrideID), body, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
