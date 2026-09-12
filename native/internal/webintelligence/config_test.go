package webintelligence

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigBytesBuildsControllerAndHealth(t *testing.T) {
	controller, err := loadConfigBytes([]byte(`{
		"schema_version": 1,
		"budget": {"limit_micros": 1000000, "per_operation_micros": 250000},
		"providers": [
			{"id":"tinyfish-search-v1","capability":"search","metered":false,"priority":10},
			{"id":"tinyfish-research-v1","capability":"research","metered":true,"priority":20,"health":"degraded"}
		]
	}`))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	snapshot := controller.Snapshot()
	if snapshot.Budget.LimitMicros != 1000000 || snapshot.Budget.PerOperationMicros != 250000 {
		t.Fatalf("unexpected budget snapshot: %#v", snapshot.Budget)
	}
	if len(snapshot.Providers) != 2 {
		t.Fatalf("provider count = %d", len(snapshot.Providers))
	}
	if snapshot.Providers[0].ID != "tinyfish-research-v1" || snapshot.Providers[0].Health != HealthDegraded {
		t.Fatalf("unexpected first provider: %#v", snapshot.Providers[0])
	}
	if snapshot.Providers[1].ID != "tinyfish-search-v1" || snapshot.Providers[1].Health != HealthHealthy {
		t.Fatalf("unexpected second provider: %#v", snapshot.Providers[1])
	}
}

func TestLoadConfigBytesRejectsUnknownFields(t *testing.T) {
	_, err := loadConfigBytes([]byte(`{"schema_version":1,"budget":{},"providers":[],"secret":"nope"}`))
	if err == nil {
		t.Fatal("expected unknown field failure")
	}
}

func TestLoadConfigBytesRejectsTrailingData(t *testing.T) {
	_, err := loadConfigBytes([]byte(`{"schema_version":1,"budget":{},"providers":[]} {}`))
	if err == nil {
		t.Fatal("expected trailing data failure")
	}
}

func TestLoadConfigBytesRejectsUnsupportedVersion(t *testing.T) {
	_, err := loadConfigBytes([]byte(`{"schema_version":2,"budget":{},"providers":[]}`))
	if err == nil {
		t.Fatal("expected schema version failure")
	}
}

func TestLoadConfigBytesRejectsInvalidHealth(t *testing.T) {
	_, err := loadConfigBytes([]byte(`{
		"schema_version":1,
		"budget":{},
		"providers":[{"id":"tinyfish-search-v1","capability":"search","health":"unknown"}]
	}`))
	if err == nil {
		t.Fatal("expected invalid health failure")
	}
}

func TestLoadConfigBytesRejectsDuplicateProvider(t *testing.T) {
	_, err := loadConfigBytes([]byte(`{
		"schema_version":1,
		"budget":{},
		"providers":[
			{"id":"tinyfish-search-v1","capability":"search"},
			{"id":"tinyfish-search-v1","capability":"search"}
		]
	}`))
	if err == nil {
		t.Fatal("expected duplicate provider failure")
	}
}

func TestLoadConfigFileBoundsInput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "web-intelligence.json")
	body := make([]byte, MaxConfigBytes+1)
	if err := os.WriteFile(path, body, 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	_, err := LoadConfigFile(path)
	if err == nil {
		t.Fatal("expected size failure")
	}
}

func TestLoadFromEnvironmentAbsentIsNotConfigured(t *testing.T) {
	old, hadOld := os.LookupEnv(ConfigEnvironment)
	if err := os.Unsetenv(ConfigEnvironment); err != nil {
		t.Fatalf("unset env: %v", err)
	}
	t.Cleanup(func() {
		if hadOld {
			_ = os.Setenv(ConfigEnvironment, old)
		} else {
			_ = os.Unsetenv(ConfigEnvironment)
		}
	})
	controller, configured, err := LoadFromEnvironment()
	if err != nil {
		t.Fatalf("load absent config: %v", err)
	}
	if configured || controller != nil {
		t.Fatalf("expected unconfigured nil controller, got configured=%v controller=%v", configured, controller)
	}
}

func TestLoadConfigBytesBudgetFailurePreservesStableError(t *testing.T) {
	_, err := loadConfigBytes([]byte(`{
		"schema_version":1,
		"budget":{"limit_micros":100,"per_operation_micros":101},
		"providers":[]
	}`))
	if err == nil || !errors.Is(err, ErrInvalidOperation) {
		t.Fatalf("expected ErrInvalidOperation, got %v", err)
	}
}
