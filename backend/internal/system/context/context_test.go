/*
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package context

import (
	"context"
	"testing"
)

func TestGetTraceID_WithNilContext(t *testing.T) {
	traceID := GetTraceID(context.TODO())
	if traceID == "" {
		t.Error("Expected non-empty trace ID, got empty string")
	}
}

func TestGetTraceID_WithEmptyContext(t *testing.T) {
	ctx := context.Background()
	traceID := GetTraceID(ctx)
	if traceID == "" {
		t.Error("Expected non-empty trace ID, got empty string")
	}
}

func TestGetTraceID_WithExistingTraceID(t *testing.T) {
	expectedID := "test-trace-id-123"
	ctx := WithTraceID(context.Background(), expectedID)

	traceID := GetTraceID(ctx)
	if traceID != expectedID {
		t.Errorf("Expected trace ID %s, got %s", expectedID, traceID)
	}
}

func TestWithTraceID(t *testing.T) {
	expectedID := "custom-trace-id"
	ctx := WithTraceID(context.Background(), expectedID)

	traceID := GetTraceID(ctx)
	if traceID != expectedID {
		t.Errorf("Expected trace ID %s, got %s", expectedID, traceID)
	}
}

func TestWithTraceID_NilContext(t *testing.T) {
	expectedID := "custom-trace-id"
	ctx := WithTraceID(context.TODO(), expectedID)

	traceID := GetTraceID(ctx)
	if traceID != expectedID {
		t.Errorf("Expected trace ID %s, got %s", expectedID, traceID)
	}
}

func TestEnsureTraceID_CreatesNew(t *testing.T) {
	ctx := context.Background()
	ctx = EnsureTraceID(ctx)

	traceID := GetTraceID(ctx)
	if traceID == "" {
		t.Error("Expected non-empty trace ID after EnsureTraceID")
	}
}

func TestEnsureTraceID_PreservesExisting(t *testing.T) {
	expectedID := "existing-trace-id"
	ctx := WithTraceID(context.Background(), expectedID)
	ctx = EnsureTraceID(ctx)

	traceID := GetTraceID(ctx)
	if traceID != expectedID {
		t.Errorf("Expected trace ID %s to be preserved, got %s", expectedID, traceID)
	}
}

func TestEnsureTraceID_NilContext(t *testing.T) {
	ctx := EnsureTraceID(context.TODO())

	traceID := GetTraceID(ctx)
	if traceID == "" {
		t.Error("Expected non-empty trace ID after EnsureTraceID with nil context")
	}
}

func TestGenerateUUID_Format(t *testing.T) {
	uuid := generateUUID()

	// UUID should be in format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	if len(uuid) != 36 {
		t.Errorf("Expected UUID length of 36, got %d", len(uuid))
	}

	// Check for dashes at correct positions
	if uuid[8] != '-' || uuid[13] != '-' || uuid[18] != '-' || uuid[23] != '-' {
		t.Errorf("UUID format incorrect: %s", uuid)
	}
}

func TestGenerateUUID_Uniqueness(t *testing.T) {
	uuid1 := generateUUID()
	uuid2 := generateUUID()

	if uuid1 == uuid2 {
		t.Error("Expected different UUIDs, got same value")
	}
}

func TestGetTraceID_WithNilContextExplicit(t *testing.T) {
	// Test with explicitly nil context
	traceID := GetTraceID(context.TODO())
	if traceID == "" {
		t.Error("Expected non-empty trace ID with nil context, got empty string")
	}
}

func TestGetTraceID_WithEmptyTraceIDValue(t *testing.T) {
	// Test with empty trace ID value in context
	ctx := context.WithValue(context.Background(), TraceIDKey, "")
	traceID := GetTraceID(ctx)
	if traceID == "" {
		t.Error("Expected non-empty trace ID when context has empty trace ID, got empty string")
	}
}

func TestGetTraceID_WithWrongType(t *testing.T) {
	// Test with wrong type in context (not a string)
	ctx := context.WithValue(context.Background(), TraceIDKey, 12345)
	traceID := GetTraceID(ctx)
	if traceID == "" {
		t.Error("Expected non-empty trace ID when context has wrong type, got empty string")
	}
}

func TestWithTraceID_WithNilContextExplicit(t *testing.T) {
	// Test with explicitly nil context
	expectedID := "test-id-nil-ctx"
	ctx := WithTraceID(context.TODO(), expectedID)

	traceID := GetTraceID(ctx)
	if traceID != expectedID {
		t.Errorf("Expected trace ID %s with nil context, got %s", expectedID, traceID)
	}
}

func TestWithTraceID_EmptyString(t *testing.T) {
	// Test setting empty trace ID
	ctx := WithTraceID(context.Background(), "")

	// Should still set it in context even if empty
	if ctx == nil {
		t.Error("Expected non-nil context")
	}

	// GetTraceID should generate a new one since it's empty
	traceID := GetTraceID(ctx)
	if traceID == "" {
		t.Error("Expected GetTraceID to generate new ID when stored value is empty")
	}
}

func TestEnsureTraceID_WithNilContextExplicit(t *testing.T) {
	// Test with explicitly nil context
	ctx := EnsureTraceID(context.TODO())

	if ctx == nil {
		t.Error("Expected non-nil context")
	}

	traceID := GetTraceID(ctx)
	if traceID == "" {
		t.Error("Expected non-empty trace ID after EnsureTraceID with nil context")
	}
}

func TestEnsureTraceID_WithEmptyTraceID(t *testing.T) {
	// Test with context containing empty trace ID
	ctx := context.WithValue(context.Background(), TraceIDKey, "")
	ctx = EnsureTraceID(ctx)

	traceID := GetTraceID(ctx)
	if traceID == "" {
		t.Error("Expected EnsureTraceID to generate new ID when empty")
	}
}

func TestEnsureTraceID_WithWrongType(t *testing.T) {
	// Test with context containing wrong type for trace ID
	ctx := context.WithValue(context.Background(), TraceIDKey, 999)
	ctx = EnsureTraceID(ctx)

	traceID := GetTraceID(ctx)
	if traceID == "" {
		t.Error("Expected EnsureTraceID to generate new ID when type is wrong")
	}

	// Verify it's now a string in the context
	if _, ok := ctx.Value(TraceIDKey).(string); !ok {
		t.Error("Expected trace ID to be a string after EnsureTraceID")
	}
}

func TestGenerateUUID_Version4(t *testing.T) {
	uuid := generateUUID()

	// Check version bits (should be 0100 for version 4)
	// The version is in the 7th byte (after 3rd dash at position 14-15)
	// Format: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx where 4 is version
	if uuid[14] != '4' {
		t.Errorf("Expected UUID version 4, got version %c", uuid[14])
	}

	// Check variant bits (should be 10xx for RFC 4122)
	// The variant is in the 9th byte (after 4th dash at position 19)
	// Valid values are 8, 9, a, b (binary 10xx)
	variantChar := uuid[19]
	validVariants := map[byte]bool{'8': true, '9': true, 'a': true, 'b': true}
	if !validVariants[variantChar] {
		t.Errorf("Expected UUID variant to be 8, 9, a, or b, got %c", variantChar)
	}
}

func TestGenerateUUID_MultipleGenerations(t *testing.T) {
	// Generate many UUIDs to ensure no collisions
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		uuid := generateUUID()
		if seen[uuid] {
			t.Errorf("Duplicate UUID generated: %s", uuid)
		}
		seen[uuid] = true
	}
}
