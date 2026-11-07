// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================

package platformxe

import "fmt"

// StorageService handles media file storage operations.
type StorageService struct {
	client *Client
}

// GetProcessor returns the storage processor configuration.
func (s *StorageService) GetProcessor() (*ProcessorConfig, error) {
	var result ProcessorConfig
	err := s.client.doRequestTyped("GET", "/api/v1/storage/processor", nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateProcessor updates the storage processor configuration.
func (s *StorageService) UpdateProcessor(input map[string]interface{}) (*ProcessorConfig, error) {
	var result ProcessorConfig
	err := s.client.doRequestTyped("PUT", "/api/v1/storage/processor", input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// SignUpload returns a signed upload URL for direct file upload.
func (s *StorageService) SignUpload(filename, contentType string) (*SignUploadResult, error) {
	var result SignUploadResult
	err := s.client.doRequestTyped("POST", "/api/v1/storage/media/sign-upload", map[string]interface{}{
		"filename": filename, "contentType": contentType,
	}, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// RegisterURL registers an externally hosted file URL.
func (s *StorageService) RegisterURL(fileURL string) (*StorageFile, error) {
	var result StorageFile
	err := s.client.doRequestTyped("POST", "/api/v1/storage/media/register", map[string]interface{}{"url": fileURL}, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ListFiles returns stored media files with optional filters.
func (s *StorageService) ListFiles(params map[string]string) (*StorageFileListResponse, error) {
	var result StorageFileListResponse
	err := s.client.doRequestTyped("GET", "/api/v1/storage/media/files", nil, params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetFile returns a media file by ID.
func (s *StorageService) GetFile(id string) (*StorageFile, error) {
	var result StorageFile
	err := s.client.doRequestTyped("GET", fmt.Sprintf("/api/v1/storage/media/files/%s", id), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteFile deletes a media file by ID.
func (s *StorageService) DeleteFile(id string) (*DeleteResult, error) {
	var result DeleteResult
	err := s.client.doRequestTyped("DELETE", fmt.Sprintf("/api/v1/storage/media/files/%s", id), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ReorderFiles updates the display order of media files.
func (s *StorageService) ReorderFiles(fileIds []string) (*ReorderFilesResult, error) {
	var result ReorderFilesResult
	err := s.client.doRequestTyped("POST", "/api/v1/storage/media/files/reorder", map[string]interface{}{"fileIds": fileIds}, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Upload uploads a single file (base64-encoded payload).
func (s *StorageService) Upload(input map[string]interface{}) (*UploadResult, error) {
	var result UploadResult
	err := s.client.doRequestTyped("POST", "/api/v1/storage/media/upload", input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// BatchUpload uploads multiple files in a single request (base64-encoded payloads).
func (s *StorageService) BatchUpload(input map[string]interface{}) (*BatchUploadResult, error) {
	var result BatchUploadResult
	err := s.client.doRequestTyped("POST", "/api/v1/storage/media/upload/batch", input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
