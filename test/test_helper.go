package test

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/conductorone/baton-panda-doc/pkg/client"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
)

var (
	SystemRoles = []string{"Admin", "Member", "Collaborator", "Manager"}
)

// Custom RoundTripper for testing.
type TestRoundTripper struct {
	response *http.Response
	err      error
}

type MockRoundTripper struct {
	Response  *http.Response
	Err       error
	roundTrip func(*http.Request) (*http.Response, error)
}

func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTrip(req)
}

func (m *MockRoundTripper) SetRoundTrip(roundTrip func(*http.Request) (*http.Response, error)) {
	m.roundTrip = roundTrip
}

func (t *TestRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return t.response, t.err
}

// Helper function to create a test client with custom transport.
func NewTestClient(response *http.Response, err error) *client.PandaDocClient {
	transport := &TestRoundTripper{response: response, err: err}
	httpClient := &http.Client{Transport: transport}
	baseHttpClient := uhttp.NewBaseHttpClient(httpClient)
	return client.NewClient(baseHttpClient)
}

func ReadFile(fileName string) string {
	data, err := os.ReadFile("../../test/mock_responses/" + fileName)
	if err != nil {
		log.Fatal(err)
	}

	return string(data)
}

func GetUniqueTime() time.Time {
	result, _ := time.Parse(time.RFC3339, "2025-02-25T13:46:12Z")
	return result
}
