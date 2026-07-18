package middleware

import "testing"

func TestAgentOSManagedKeyBodyIsOmittedFromAudit(t *testing.T) {
	route := "POST /api/v1/admin/users/:id/api-keys"
	if _, ok := auditBodyOmittedRoutes[route]; !ok {
		t.Fatal("托管 Key 创建请求体必须整体从审计中省略")
	}
	if auditActionOverrides[route] != "admin.users.api_keys.create" {
		t.Fatal("托管 Key 创建必须保留稳定审计动作")
	}
}
