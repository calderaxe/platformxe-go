// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-10-08, v1.7.0 pending — calderasuite/caldera-xadmin#29):
// - Created. KnowledgeService.UpsertDocument / DeleteDocument /
//   ListDocuments / Search for PUT / DELETE / GET
//   /api/v1/ai/knowledge/documents and POST /api/v1/ai/knowledge/search —
//   scope ai:knowledge, 1,200 requests/hr/key. Typed request / response
//   structs in types.go mirror @caldera/platformxe-types 3.6.0. Errors are
//   *APIError with the closed ErrCodeAi* set.
//
// CHANGELOG (2026-10-09, v1.7.0 pending — final review):
// - UpsertDocument waits up to 65 s and Search up to 35 s: both embed inline
//   and run longer than the client's 10 s default.
// =============================================================================

package platformxe

import (
	"strconv"
	"time"
)

// KnowledgeService indexes, lists, deletes and searches the calling
// service's knowledge documents.
type KnowledgeService struct {
	client *Client
}

// UpsertDocument creates or replaces one document (text ≤ 400,000
// characters, markdown or plain). Chunks are replaced only when the content
// changed.
func (s *KnowledgeService) UpsertDocument(input KnowledgeUpsertDocumentInput) (*KnowledgeUpsertDocumentResult, error) {
	var result KnowledgeUpsertDocumentResult
	// The upsert embeds inline (route limit 60 s).
	if err := s.client.doRequestTypedOpts("PUT", "/api/v1/ai/knowledge/documents", input, nil, nil, callOptions{timeout: 65 * time.Second}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteDocument deletes one document; Deleted is false when it was absent.
func (s *KnowledgeService) DeleteDocument(collection, sourceRef string) (*KnowledgeDeleteDocumentResult, error) {
	var result KnowledgeDeleteDocumentResult
	params := map[string]string{"collection": collection, "sourceRef": sourceRef}
	if err := s.client.doRequestTyped("DELETE", "/api/v1/ai/knowledge/documents", nil, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListDocuments returns one page of a collection's documents (no text);
// pass NextCursor back as Cursor for the next page.
func (s *KnowledgeService) ListDocuments(query KnowledgeListQuery) (*KnowledgeDocumentList, error) {
	params := map[string]string{"collection": query.Collection}
	if query.Cursor != "" {
		params["cursor"] = query.Cursor
	}
	if query.Limit > 0 {
		params["limit"] = strconv.Itoa(query.Limit)
	}
	var result KnowledgeDocumentList
	if err := s.client.doRequestTyped("GET", "/api/v1/ai/knowledge/documents", nil, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Search runs a hybrid search over the given collections. Each hit carries
// SourceRef + Ordinal for citation.
func (s *KnowledgeService) Search(input KnowledgeSearchInput) (*KnowledgeSearchResult, error) {
	if input.ACLTags == nil {
		input.ACLTags = []string{}
	}
	var result KnowledgeSearchResult
	// The search embeds the query (route limit 30 s).
	if err := s.client.doRequestTypedOpts("POST", "/api/v1/ai/knowledge/search", input, nil, nil, callOptions{timeout: 35 * time.Second}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
