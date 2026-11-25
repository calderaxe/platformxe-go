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
//
// CHANGELOG (2026-11-18, v1.12.0 pending): the tenant category registry and
// settings: ListCategories, CreateCategory, UpdateCategory,
// DeactivateCategory (/api/v1/storage/fixed/categories), GetSettings and
// UpdateSettings (/api/v1/storage/fixed/settings), with DocumentCategory,
// CreateDocumentCategoryInput and DocumentStorageSettings.
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

// ----------------------------------------------------------------------------
// Category registry and settings (v1.12.0)
// ----------------------------------------------------------------------------
//
// The tenant category registry (/api/v1/storage/fixed/categories) and its
// enforcement switch (/api/v1/storage/fixed/settings) belong to the key's
// caller service within its organisation: another tenant using the same
// caller name never sees or changes them. A field of the wrong type is a 400
// BAD_REQUEST *APIError naming the field; a key that normalises to nothing is
// 400 VALIDATION_ERROR; a duplicate key on create is 409 CONFLICT; an unknown
// key on update or deactivate is 404.

const (
	documentCategoriesPath = "/api/v1/storage/fixed/categories"
	documentSettingsPath   = "/api/v1/storage/fixed/settings"
)

// The retention types and who may delete (DocumentCategory.RetentionType and
// DeletableBy; a value outside them is 400).
const (
	RetentionTypeNone      = "NONE"
	RetentionTypeStandard  = "STANDARD"
	RetentionTypeExtended  = "EXTENDED"
	RetentionTypeLegal     = "LEGAL"
	RetentionTypePermanent = "PERMANENT"

	DeletableByOwner      = "OWNER"
	DeletableByAdmin      = "ADMIN"
	DeletableBySuperAdmin = "SUPER_ADMIN"
	DeletableByNone       = "NONE"
)

// DocumentCategory is one registered category. The retention fields are nil
// when the category falls back to the platform default.
type DocumentCategory struct {
	ID               string  `json:"id"`
	CallerService    string  `json:"callerService"`
	Key              string  `json:"key"`
	Label            string  `json:"label"`
	Description      *string `json:"description"`
	SortOrder        int     `json:"sortOrder"`
	Color            *string `json:"color"`
	IsActive         bool    `json:"isActive"`
	RetentionType    *string `json:"retentionType"`
	RetentionDays    *int    `json:"retentionDays"`
	DeletableBy      *string `json:"deletableBy"`
	RequiresOverride bool    `json:"requiresOverride"`
	RequiresWitness  bool    `json:"requiresWitness"`
	CreatedAt        string  `json:"createdAt"`
	UpdatedAt        string  `json:"updatedAt"`
}

// CreateDocumentCategoryInput is the body of CreateCategory. Key and Label are
// required; the key is stored in UPPER_SNAKE. A nil field is not sent.
type CreateDocumentCategoryInput struct {
	Key              string  `json:"key"`
	Label            string  `json:"label"`
	Description      *string `json:"description,omitempty"`
	SortOrder        *int    `json:"sortOrder,omitempty"`
	Color            *string `json:"color,omitempty"`
	RetentionType    *string `json:"retentionType,omitempty"`
	RetentionDays    *int    `json:"retentionDays,omitempty"`
	DeletableBy      *string `json:"deletableBy,omitempty"`
	RequiresOverride *bool   `json:"requiresOverride,omitempty"`
	RequiresWitness  *bool   `json:"requiresWitness,omitempty"`
}

// DocumentCategoryListResult is ListCategories' answer.
type DocumentCategoryListResult struct {
	Categories []DocumentCategory `json:"categories"`
}

// documentCategoryResult wraps one category ({"category": {...}}).
type documentCategoryResult struct {
	Category DocumentCategory `json:"category"`
}

// DocumentStorageSettings are the fixed-storage settings of the key's caller
// service. EnforceCategories (off by default) refuses a document whose
// category is not registered and active (INVALID_CATEGORY).
type DocumentStorageSettings struct {
	EnforceCategories bool `json:"enforceCategories"`
}

type documentSettingsResult struct {
	Settings DocumentStorageSettings `json:"settings"`
}

// ListCategories lists the registered categories, active ones only unless
// includeInactive (GET /api/v1/storage/fixed/categories).
func (s *DocumentsService) ListCategories(includeInactive bool) (*DocumentCategoryListResult, error) {
	var q map[string]string
	if includeInactive {
		q = map[string]string{"includeInactive": "true"}
	}
	var result DocumentCategoryListResult
	if err := s.client.doRequestTyped("GET", documentCategoriesPath, nil, q, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateCategory registers a category (POST, 201).
func (s *DocumentsService) CreateCategory(input CreateDocumentCategoryInput) (*DocumentCategory, error) {
	var result documentCategoryResult
	if err := s.client.doRequestTyped("POST", documentCategoriesPath, input, nil, &result); err != nil {
		return nil, err
	}
	return &result.Category, nil
}

// UpdateCategory changes the fields given in fields, by their JSON names
// (label, description, sortOrder, color, isActive, retentionType,
// retentionDays, deletableBy, requiresOverride, requiresWitness). A nil value
// is sent as null and clears description, color, retentionType, retentionDays
// or deletableBy (back to the platform default); the other fields do not take
// null (400). PATCH /api/v1/storage/fixed/categories?key=.
func (s *DocumentsService) UpdateCategory(key string, fields map[string]interface{}) (*DocumentCategory, error) {
	if fields == nil {
		fields = map[string]interface{}{}
	}
	var result documentCategoryResult
	if err := s.client.doRequestTyped("PATCH", documentCategoriesPath, fields, map[string]string{"key": key}, &result); err != nil {
		return nil, err
	}
	return &result.Category, nil
}

// DeactivateCategory soft-disables a category; its documents keep it
// (DELETE /api/v1/storage/fixed/categories?key=).
func (s *DocumentsService) DeactivateCategory(key string) (*DocumentCategory, error) {
	var result documentCategoryResult
	if err := s.client.doRequestTyped("DELETE", documentCategoriesPath, nil, map[string]string{"key": key}, &result); err != nil {
		return nil, err
	}
	return &result.Category, nil
}

// GetSettings reads the fixed-storage settings (GET /api/v1/storage/fixed/settings).
func (s *DocumentsService) GetSettings() (*DocumentStorageSettings, error) {
	var result documentSettingsResult
	if err := s.client.doRequestTyped("GET", documentSettingsPath, nil, nil, &result); err != nil {
		return nil, err
	}
	return &result.Settings, nil
}

// UpdateSettings turns category enforcement on or off
// (PATCH /api/v1/storage/fixed/settings {"enforceCategories": bool}).
func (s *DocumentsService) UpdateSettings(enforceCategories bool) (*DocumentStorageSettings, error) {
	var result documentSettingsResult
	body := map[string]interface{}{"enforceCategories": enforceCategories}
	if err := s.client.doRequestTyped("PATCH", documentSettingsPath, body, nil, &result); err != nil {
		return nil, err
	}
	return &result.Settings, nil
}
