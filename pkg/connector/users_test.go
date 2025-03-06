package connector

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/conductorone/baton-panda-doc/pkg/client"
	"github.com/conductorone/baton-panda-doc/test"
)

// Test that client can fetch all users.
func TestPandaDocClient_ListUsers(t *testing.T) {
	// Create a mock response.
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(test.ReadFile("mock_users.json"))),
	}
	mockResponse.Header.Set("Content-Type", "application/json")
	// Create a test client with the mock response.
	testClient := test.NewTestClient(mockResponse, nil)

	// Call GetUsers.
	ctx := context.Background()

	result, _, nextOptions, err := testClient.ListUsers(ctx, client.PageOptions{
		Count: 50,
		Page:  1,
	})

	// Check for errors.
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify the result.
	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// Check count.
	if len(result) != 2 {
		t.Errorf("Expected Count to be 2, got %d", len(result))
	}

	for index, user := range result {
		expectedUserEmail := fmt.Sprintf("testUser0%d@test.com", index+1)

		if expectedUserEmail != user.Email {
			t.Errorf("Unexpected user: got %+v, want %+v", user.Email, expectedUserEmail)
		}
		workspaces := user.Workspaces
		if len(workspaces) < 1 {
			t.Errorf("Expected user to belong to at least one workspace")
		}
	}

	// Check next options.
	if nextOptions == nil {
		t.Fatal("Expected non-nil nextOptions")
	}
}
