package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
)

// Endpoints for PandaDoc API
const (
	baseURL = "https://api.pandadoc.com/public/v1"

	// GET Endpoints
	allUsers      = "/users"
	allWorkspaces = "/workspaces"
)

type PandaDocClient struct {
	httpClient  *uhttp.BaseHttpClient
	pandaDocURL string
	domain      string
	token       string
}

type Option func(client *PandaDocClient)

func New(ctx context.Context, opts ...Option) (*PandaDocClient, error) {
	pandaDocClient := &PandaDocClient{
		httpClient:  &uhttp.BaseHttpClient{},
		pandaDocURL: baseURL,
		domain:      "",
		token:       "",
	}

	for _, opt := range opts {
		opt(pandaDocClient)
	}

	httpClient, err := uhttp.NewClient(ctx, uhttp.WithLogger(true, ctxzap.Extract(ctx)))

	if err != nil {
		return nil, err
	}

	cli, err := uhttp.NewBaseHttpClientWithContext(context.Background(), httpClient)
	if err != nil {
		return nil, err
	}

	dotIndex := strings.Index(baseURL, ".")
	if dotIndex == -1 {
		return nil, fmt.Errorf("invalid URL: %s", baseURL)
	}

	pDocURL := baseURL
	if pandaDocClient.domain == "eu" {
		baseURLCopy := baseURL
		pDocURL = strings.Replace(baseURLCopy, ".com", ".eu", 1)
	}

	if !isValidUrl(pDocURL) {
		return nil, fmt.Errorf("invalid URL: %s", pDocURL)
	}

	pandaDocClient.httpClient = cli
	pandaDocClient.pandaDocURL = pDocURL

	return pandaDocClient, nil
}

func NewClient(httpClient ...*uhttp.BaseHttpClient) *PandaDocClient {
	var wrapper = &uhttp.BaseHttpClient{}
	if httpClient != nil || len(httpClient) != 0 {
		wrapper = httpClient[0]
	}
	return &PandaDocClient{
		httpClient:  wrapper,
		pandaDocURL: "http://test.com",
		domain:      "",
		token:       "",
	}
}
func WithBearerToken(apiToken string) Option {
	return func(c *PandaDocClient) {
		c.token = apiToken
	}
}

func WithDomain(domain string) Option {
	return func(c *PandaDocClient) {
		c.domain = domain
	}
}

func (p *PandaDocClient) getToken() string {
	return p.token
}

func (p *PandaDocClient) GetDomain() string {
	return p.domain
}

func isValidUrl(urlBase string) bool {
	u, err := url.Parse(urlBase)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func (c *PandaDocClient) getResourcesFromAPI(
	ctx context.Context,
	urlAddress string,
	res any,
	reqOpt ...ReqOpt,
) (string, annotations.Annotations, error) {
	_, annotation, err := c.doRequest(ctx, http.MethodGet, urlAddress, &res, nil, reqOpt...)

	if err != nil {
		return "", nil, err
	}

	return "", annotation, nil
}

func (c *PandaDocClient) doRequest(
	ctx context.Context,
	method string,
	endpointUrl string,
	res interface{},
	body interface{},
	reqOpt ...ReqOpt,
) (http.Header, annotations.Annotations, error) {
	var (
		resp *http.Response
		err  error
	)

	urlAddress, err := url.Parse(endpointUrl)

	if err != nil {
		return nil, nil, err
	}

	for _, o := range reqOpt {
		o(urlAddress)
	}

	req, err := c.httpClient.NewRequest(
		ctx,
		method,
		urlAddress,
		uhttp.WithAcceptJSONHeader(),
		uhttp.WithContentTypeJSONHeader(),
		uhttp.WithHeader("Authorization", "API-Key "+c.getToken()),
		uhttp.WithJSONBody(body),
	)

	if err != nil {
		return nil, nil, err
	}

	switch method {
	case http.MethodGet, http.MethodPut, http.MethodPost:
		var doOptions []uhttp.DoOption
		if res != nil {
			doOptions = append(doOptions, uhttp.WithResponse(&res))
		}
		resp, err = c.httpClient.Do(req, doOptions...)

		if resp != nil {
			defer resp.Body.Close()
		}
	case http.MethodDelete:
		resp, err = c.httpClient.Do(req)
		if resp != nil {
			defer resp.Body.Close()
		}
	}

	if err != nil {
		return nil, nil, err
	}

	annotation := annotations.Annotations{}

	return resp.Header, annotation, nil
}

type UserResponse struct {
	Users []User `json:"results"`
	Total int    `json:"total"`
}

type WorkspaceResponse struct {
	Workspaces []Workspace `json:"results"`
	Total      int         `json:"total"`
}

func (c *PandaDocClient) ListUsers(ctx context.Context, opts PageOptions) ([]User, string, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)
	var res UserResponse

	queryUrl, err := url.JoinPath(c.pandaDocURL, allUsers)
	if err != nil {
		l.Error(fmt.Sprintf("Error creating url: %s", err))
		return nil, "", nil, err
	}

	pageToken, annotation, err := c.getResourcesFromAPI(ctx, queryUrl, &res, WithPage(opts.Page), WithPageLimit(opts.Count))

	if err != nil {
		l.Error(fmt.Sprintf("Error getting resources: %s", err))
		return nil, "", nil, err
	}

	return res.Users, pageToken, annotation, nil
}

func (c *PandaDocClient) ListWorkspaces(ctx context.Context, opts PageOptions) ([]Workspace, string, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)
	var res WorkspaceResponse

	queryUrl, err := url.JoinPath(c.pandaDocURL, allWorkspaces)
	if err != nil {
		l.Error(fmt.Sprintf("Error creating url: %s", err))
		return nil, "", nil, err
	}

	pageToken, annotation, err := c.getResourcesFromAPI(ctx, queryUrl, &res, WithPage(opts.Page), WithPageLimit(opts.Count))

	if err != nil {
		l.Error(fmt.Sprintf("Error getting resources: %s", err))
		return nil, "", nil, err
	}

	return res.Workspaces, pageToken, annotation, nil
}
