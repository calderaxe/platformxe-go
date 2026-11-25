// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================

package platformxe

import (
	"encoding/json"
	"fmt"
)

// PermissionsService handles permission checks, roles, policies, relationships,
// modules, audit, and federation operations.
type PermissionsService struct {
	client *Client
}

// Check performs a single permission check.
func (s *PermissionsService) Check(adminID, path, action string) (*PermissionCheckResult, error) {
	var result PermissionCheckResult
	err := s.client.doRequestTyped("POST", "/api/v1/permissions/check", map[string]interface{}{
		"adminId": adminID, "path": path, "action": action,
	}, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// CheckBatch performs multiple permission checks in a single request.
//
// The route answers a bare array (in request order). Before v1.11.0 this
// decoded the response as {"results": [...]}, so Results was always empty;
// it now reads the array, and still reads an object with "results".
func (s *PermissionsService) CheckBatch(checks []map[string]interface{}) (*BatchCheckResponse, error) {
	raw, err := s.client.doRequestOpts("POST", "/api/v1/permissions/check-batch", map[string]interface{}{"checks": checks}, nil, nil, callOptions{})
	if err != nil {
		return nil, err
	}
	var result BatchCheckResponse
	var source interface{} = raw
	target := interface{}(&result)
	if list, ok := raw["data"].([]interface{}); ok {
		source = list
		target = &result.Results
	}
	jsonBytes, err := json.Marshal(source)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}
	if err := json.Unmarshal(jsonBytes, target); err != nil {
		return nil, err
	}
	return &result, nil
}

// Resolve returns the full resolved permission set for an admin.
func (s *PermissionsService) Resolve(adminID string) (*ResolvedCapabilities, error) {
	var result ResolvedCapabilities
	err := s.client.doRequestTyped("GET", fmt.Sprintf("/api/v1/permissions/resolve/%s", adminID), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ListRoles returns all roles for the organization.
func (s *PermissionsService) ListRoles() (*RoleListResponse, error) {
	var result RoleListResponse
	err := s.client.doRequestTyped("GET", "/api/v1/permissions/roles", nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateRole creates a new role.
func (s *PermissionsService) CreateRole(name, description, model string) (*PermissionRole, error) {
	var result PermissionRole
	err := s.client.doRequestTyped("POST", "/api/v1/permissions/roles", map[string]interface{}{
		"name": name, "description": description, "model": model,
	}, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetRole returns a single role by ID.
func (s *PermissionsService) GetRole(roleID string) (*PermissionRole, error) {
	var result PermissionRole
	err := s.client.doRequestTyped("GET", fmt.Sprintf("/api/v1/permissions/roles/%s", roleID), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateRole updates a role by ID.
func (s *PermissionsService) UpdateRole(roleID string, updates map[string]interface{}) (*PermissionRole, error) {
	var result PermissionRole
	err := s.client.doRequestTyped("PATCH", fmt.Sprintf("/api/v1/permissions/roles/%s", roleID), updates, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteRole deletes a role by ID.
func (s *PermissionsService) DeleteRole(roleID string) (*DeleteResult, error) {
	var result DeleteResult
	err := s.client.doRequestTyped("DELETE", fmt.Sprintf("/api/v1/permissions/roles/%s", roleID), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetRoleCapabilities returns capabilities for a role (Simple model).
func (s *PermissionsService) GetRoleCapabilities(roleID string) (*RoleCapabilities, error) {
	var result RoleCapabilities
	err := s.client.doRequestTyped("GET", fmt.Sprintf("/api/v1/permissions/roles/%s/capabilities", roleID), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// SetRoleCapabilities sets capabilities for a role (Simple model).
func (s *PermissionsService) SetRoleCapabilities(roleID string, capabilities []string) (*RoleCapabilities, error) {
	var result RoleCapabilities
	err := s.client.doRequestTyped("PUT", fmt.Sprintf("/api/v1/permissions/roles/%s/capabilities", roleID), map[string]interface{}{"capabilities": capabilities}, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetRoleModulePermissions returns module permissions for a role (Full model).
func (s *PermissionsService) GetRoleModulePermissions(roleID string) (*RoleModulePermissions, error) {
	var result RoleModulePermissions
	err := s.client.doRequestTyped("GET", fmt.Sprintf("/api/v1/permissions/roles/%s/modules", roleID), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// SetRoleModulePermissions sets module permissions for a role (Full model).
func (s *PermissionsService) SetRoleModulePermissions(roleID string, modules []map[string]interface{}) (*RoleModulePermissions, error) {
	var result RoleModulePermissions
	err := s.client.doRequestTyped("PUT", fmt.Sprintf("/api/v1/permissions/roles/%s/modules", roleID), map[string]interface{}{"modules": modules}, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ListOverrides returns admin-level permission overrides. The API answers a
// bare array; it is wrapped in OverrideListResponse.Overrides.
func (s *PermissionsService) ListOverrides(adminID string) (*OverrideListResponse, error) {
	var overrides []PermissionOverride
	err := s.client.doRequestTypedArray("GET", fmt.Sprintf("/api/v1/permissions/overrides/%s", adminID), nil, nil, &overrides)
	if err != nil {
		return nil, err
	}
	if overrides == nil {
		overrides = []PermissionOverride{}
	}
	return &OverrideListResponse{Overrides: overrides}, nil
}

// CreateOverride creates or replaces the one permission override for
// (adminId, path, action). The API answers 201 on create and 200 when it
// replaced the existing override; a GRANT over a live DENY is refused with
// 409 OVERRIDE_DENY_ACTIVE unless input carries "replace": true. A DENY over
// a live DENY keeps the later expiry (no expiry wins).
func (s *PermissionsService) CreateOverride(input map[string]interface{}) (*PermissionOverride, error) {
	var result PermissionOverride
	err := s.client.doRequestTyped("POST", "/api/v1/permissions/overrides", input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteOverride removes a permission override.
func (s *PermissionsService) DeleteOverride(overrideID string) (*DeleteResult, error) {
	var result DeleteResult
	err := s.client.doRequestTyped("DELETE", fmt.Sprintf("/api/v1/permissions/overrides/remove/%s", overrideID), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ListPolicies returns all resource policies.
func (s *PermissionsService) ListPolicies() (*PolicyListResponse, error) {
	var result PolicyListResponse
	err := s.client.doRequestTyped("GET", "/api/v1/permissions/policies", nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// CreatePolicy creates a new resource policy.
func (s *PermissionsService) CreatePolicy(input map[string]interface{}) (*ResourcePolicy, error) {
	var result ResourcePolicy
	err := s.client.doRequestTyped("POST", "/api/v1/permissions/policies", input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdatePolicy updates a resource policy.
func (s *PermissionsService) UpdatePolicy(policyID string, updates map[string]interface{}) (*ResourcePolicy, error) {
	var result ResourcePolicy
	err := s.client.doRequestTyped("PATCH", fmt.Sprintf("/api/v1/permissions/policies/%s", policyID), updates, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeletePolicy deletes a resource policy.
func (s *PermissionsService) DeletePolicy(policyID string) (*DeleteResult, error) {
	var result DeleteResult
	err := s.client.doRequestTyped("DELETE", fmt.Sprintf("/api/v1/permissions/policies/%s", policyID), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ListRelationships returns ReBAC relationships with optional filters.
func (s *PermissionsService) ListRelationships(params map[string]string) (*RelationshipListResponse, error) {
	var result RelationshipListResponse
	err := s.client.doRequestTyped("GET", "/api/v1/permissions/relationships", nil, params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateRelationships creates, updates, or deletes ReBAC relationship tuples.
func (s *PermissionsService) UpdateRelationships(operations []map[string]interface{}) (*RelationshipUpdateResult, error) {
	var result RelationshipUpdateResult
	err := s.client.doRequestTyped("POST", "/api/v1/permissions/relationships", operations, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// RegisterModule registers a new permission module.
func (s *PermissionsService) RegisterModule(input map[string]interface{}) (*PermissionModule, error) {
	var result PermissionModule
	err := s.client.doRequestTyped("POST", "/api/v1/permissions/modules", input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ListModules returns all registered permission modules.
func (s *PermissionsService) ListModules() (*ModuleListResponse, error) {
	var result ModuleListResponse
	err := s.client.doRequestTyped("GET", "/api/v1/permissions/modules", nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetAuditLogs returns permission decision audit logs.
func (s *PermissionsService) GetAuditLogs(params map[string]string) (*AuditLogListResponse, error) {
	var result AuditLogListResponse
	err := s.client.doRequestTyped("GET", "/api/v1/permissions/audit", nil, params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ListChangeLogs returns permission mutation change logs.
func (s *PermissionsService) ListChangeLogs(params map[string]string) (*ChangeLogListResponse, error) {
	var result ChangeLogListResponse
	err := s.client.doRequestTyped("GET", "/api/v1/permissions/audit/changes", nil, params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ExportAudit exports audit logs.
func (s *PermissionsService) ExportAudit(params map[string]string) (*AuditExportResult, error) {
	var result AuditExportResult
	err := s.client.doRequestTyped("GET", "/api/v1/permissions/audit/export", nil, params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ShadowCheck performs a shadow permission check for testing policy changes.
func (s *PermissionsService) ShadowCheck(input map[string]interface{}) (*ShadowCheckResult, error) {
	var result ShadowCheckResult
	err := s.client.doRequestTyped("POST", "/api/v1/permissions/shadow-check", input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// --- Federation ---

// CreateFederationGroup creates a new federation group.
func (s *PermissionsService) CreateFederationGroup(name string) (*FederationGroup, error) {
	var result FederationGroup
	err := s.client.doRequestTyped("POST", "/api/v1/permissions/federation/groups", map[string]interface{}{"name": name}, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ListFederationGroups returns all federation groups.
func (s *PermissionsService) ListFederationGroups() (*FederationGroupListResponse, error) {
	var result FederationGroupListResponse
	err := s.client.doRequestTyped("GET", "/api/v1/permissions/federation/groups", nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetFederationGroup returns a federation group by ID.
func (s *PermissionsService) GetFederationGroup(groupID string) (*FederationGroup, error) {
	var result FederationGroup
	err := s.client.doRequestTyped("GET", fmt.Sprintf("/api/v1/permissions/federation/groups/%s", groupID), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteFederationGroup deletes a federation group.
func (s *PermissionsService) DeleteFederationGroup(groupID string) (*DeleteResult, error) {
	var result DeleteResult
	err := s.client.doRequestTyped("DELETE", fmt.Sprintf("/api/v1/permissions/federation/groups/%s", groupID), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// AddFederationMember invites an organization into a federation group.
// Only the group owner may invite. The membership is created with
// Status "INVITED": the invited organization must call
// AcceptFederationMembership with its own API key before any pull or push
// reaches it.
func (s *PermissionsService) AddFederationMember(groupID string, input map[string]interface{}) (*FederationMemberResult, error) {
	var result FederationMemberResult
	err := s.client.doRequestTyped("POST", fmt.Sprintf("/api/v1/permissions/federation/groups/%s/members", groupID), input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// AcceptFederationMembership accepts this organization's invitation to a
// federation group. Call it with the invited organization's own API key — the
// caller's organization is the one accepted. The returned membership is ACTIVE.
func (s *PermissionsService) AcceptFederationMembership(groupID string) (*FederationMemberResult, error) {
	var result FederationMemberResult
	err := s.client.doRequestTyped("POST", fmt.Sprintf("/api/v1/permissions/federation/groups/%s/members/accept", groupID), map[string]interface{}{}, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// RemoveFederationMember removes a membership from a federation group. The
// owner may remove any member; any other organization may pass its own orgID
// to leave the group or decline an invitation.
func (s *PermissionsService) RemoveFederationMember(groupID, orgID string) (*FederationRemoveResult, error) {
	var result FederationRemoveResult
	err := s.client.doRequestTyped("DELETE", fmt.Sprintf("/api/v1/permissions/federation/groups/%s/members?organizationId=%s", groupID, orgID), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// PullFederationModules pulls modules from a federation group into the target org.
func (s *PermissionsService) PullFederationModules(groupID string, targetOrgID string) (*FederationPullResult, error) {
	body := map[string]interface{}{}
	if targetOrgID != "" {
		body["targetOrgId"] = targetOrgID
	}
	var result FederationPullResult
	err := s.client.doRequestTyped("POST", fmt.Sprintf("/api/v1/permissions/federation/groups/%s/pull", groupID), body, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// PushFederationPermissions pushes resolved permissions to federation members.
func (s *PermissionsService) PushFederationPermissions(groupID string, adminIDs []string, targetOrgID string) (*FederationPushResult, error) {
	body := map[string]interface{}{"adminIds": adminIDs}
	if targetOrgID != "" {
		body["targetOrgId"] = targetOrgID
	}
	var result FederationPushResult
	err := s.client.doRequestTyped("POST", fmt.Sprintf("/api/v1/permissions/federation/groups/%s/push", groupID), body, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetFederationStatus returns the sync status of a federation group.
func (s *PermissionsService) GetFederationStatus(groupID string) (*FederationStatusResult, error) {
	var result FederationStatusResult
	err := s.client.doRequestTyped("GET", fmt.Sprintf("/api/v1/permissions/federation/groups/%s/status", groupID), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
