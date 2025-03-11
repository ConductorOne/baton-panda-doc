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

func TestPandaDocClient_ListWorkspaces(t *testing.T) {
	// Create a mock response.
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(test.ReadFile("mock_workspaces.json"))),
	}
	mockResponse.Header.Set("Content-Type", "application/json")
	// Create a test client with the mock response.
	testClient := test.NewTestClient(mockResponse, nil)

	// Call te get workspaces.
	ctx := context.Background()

	result, _, nextOptions, err := testClient.ListWorkspaces(ctx, client.PageOptions{
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

	for index, workspace := range result {
		expectedWorkspace := fmt.Sprintf("testWorkspace0%d", index+1)

		if expectedWorkspace != workspace.ID {
			t.Errorf("Unexpected workspace: got %+v, want %+v", workspace.ID, expectedWorkspace)
		}
	}

	// Check next options.
	if nextOptions == nil {
		t.Fatal("Expected non-nil nextOptions")
	}
}
