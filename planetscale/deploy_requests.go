package planetscale

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// DeployRequestReview posts a review to a deploy request.
type DeployRequestReview struct {
	ID    string `json:"id"`
	Body  string `json:"body"`
	State string `json:"state"`
}

// PerformDeployRequest is a request for approving and deploying a deploy request.
// NOTE: We deviate from naming convention here because we have a data model
// named DeployRequest already.
type PerformDeployRequest struct {
	Organization string `json:"-"`
	Database     string `json:"-"`
	Number       uint64 `json:"-"`
}

// GetDeployRequest encapsulates the request for getting a single deploy
// request.
type GetDeployRequestRequest struct {
	Organization string `json:"-"`
	Database     string `json:"-"`
	Number       uint64 `json:"-"`
}

// ListDeployRequestsRequest gets the deploy requests for a specific database
// branch.
type ListDeployRequestsRequest struct {
	Organization string
	Database     string
}

// DeployRequest encapsulates the request to deploy a database branch's schema
// to a production branch
type DeployRequest struct {
	ID string `json:"id"`

	Branch     string `json:"branch"`
	IntoBranch string `json:"into_branch"`

	Number uint64 `json:"number"`

	DeployabilityErrors string `json:"deployability_errors"`
	DeploymentState     string `json:"deployment_state"`

	State string `json:"state"`

	Ready    bool `json:"ready"`
	Approved bool `json:"approved"`
	Deployed bool `json:"deployed"`

	Notes string `json:"notes"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	ClosedAt  *time.Time `json:"closed_at"`
}

type CancelDeployRequest struct {
	Organization string `json:"-"`
	Database     string `json:"-"`
	Number       uint64 `json:"-"`
}

type CreateDeployRequestRequest struct {
	Organization string `json:"-"`
	Database     string `json:"-"`
	Branch       string `json:"branch"`
	IntoBranch   string `json:"into_branch"`
	Notes        string `json:"notes"`
}

type ReviewDeployRequestRequest struct {
	Organization string `json:"-"`
	Database     string `json:"-"`
	Number       uint64 `json:"-"`
	Body         string `json:"body"`
	State        string `json:"state"`
}

// DeployRequestsService is an interface for communicating with the PlanetScale
// deploy requests API.
type DeployRequestsService interface {
	List(context.Context, *ListDeployRequestsRequest) ([]*DeployRequest, error)
	Create(context.Context, *CreateDeployRequestRequest) (*DeployRequest, error)
	Get(context.Context, *GetDeployRequestRequest) (*DeployRequest, error)
	Deploy(context.Context, *PerformDeployRequest) (*DeployRequest, error)
	CancelDeploy(context.Context, *CancelDeployRequest) (*DeployRequest, error)
	Close(context.Context, *CloseDeployRequestRequest) (*DeployRequest, error)
	CreateReview(context.Context, *ReviewDeployRequestRequest) (*DeployRequestReview, error)
}

type CloseDeployRequestRequest struct {
	Organization string `json:"-"`
	Database     string `json:"-"`
	Number       uint64 `json:"-"`
}

type deployRequestsService struct {
	client *Client
}

var _ DeployRequestsService = &deployRequestsService{}

func NewDeployRequestsService(client *Client) *deployRequestsService {
	return &deployRequestsService{
		client: client,
	}
}

// Get fetches a single deploy request.
func (d *deployRequestsService) Get(ctx context.Context, getReq *GetDeployRequestRequest) (*DeployRequest, error) {
	req, err := d.client.newRequest(http.MethodGet, deployRequestAPIPath(getReq.Organization, getReq.Database, getReq.Number), nil)
	if err != nil {
		return nil, errors.Wrap(err, "error creating http request")
	}

	res, err := d.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	dr := &DeployRequest{}
	err = json.NewDecoder(res.Body).Decode(dr)
	if err != nil {
		return nil, err
	}

	return dr, nil
}

type CloseRequest struct {
	State string `json:"state"`
}

// Close closes a deploy request
func (d *deployRequestsService) Close(ctx context.Context, closeReq *CloseDeployRequestRequest) (*DeployRequest, error) {
	updateReq := &CloseRequest{
		State: "closed",
	}

	req, err := d.client.newRequest(http.MethodPatch, deployRequestAPIPath(closeReq.Organization, closeReq.Database, closeReq.Number), updateReq)
	if err != nil {
		return nil, errors.Wrap(err, "error creating http request")
	}

	res, err := d.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	dr := &DeployRequest{}
	err = json.NewDecoder(res.Body).Decode(dr)
	if err != nil {
		return nil, err
	}

	return dr, nil
}

// Deploy approves and executes a specific deploy request.
func (d *deployRequestsService) Deploy(ctx context.Context, deployReq *PerformDeployRequest) (*DeployRequest, error) {
	path := deployRequestActionAPIPath(deployReq.Organization, deployReq.Database, deployReq.Number, "deploy")
	req, err := d.client.newRequest(http.MethodPost, path, deployReq)
	if err != nil {
		return nil, errors.Wrap(err, "error creating http request")
	}

	res, err := d.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	dr := &DeployRequest{}
	err = json.NewDecoder(res.Body).Decode(dr)
	if err != nil {
		return nil, err
	}

	return dr, nil
}

type deployRequestsResponse struct {
	DeployRequests []*DeployRequest `json:"data"`
}

func (d *deployRequestsService) Create(ctx context.Context, createReq *CreateDeployRequestRequest) (*DeployRequest, error) {
	path := deployRequestsAPIPath(createReq.Organization, createReq.Database)
	req, err := d.client.newRequest(http.MethodPost, path, createReq)
	if err != nil {
		return nil, err
	}

	res, err := d.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	dr := &DeployRequest{}
	err = json.NewDecoder(res.Body).Decode(dr)
	if err != nil {
		return nil, err
	}

	return dr, nil
}

// CancelDeploy cancels a queued deploy request.
func (d *deployRequestsService) CancelDeploy(ctx context.Context, deployReq *CancelDeployRequest) (*DeployRequest, error) {
	path := deployRequestActionAPIPath(deployReq.Organization, deployReq.Database, deployReq.Number, "cancel")
	req, err := d.client.newRequest(http.MethodPost, path, deployReq)
	if err != nil {
		return nil, errors.Wrap(err, "error creating http request")
	}

	res, err := d.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	dr := &DeployRequest{}
	err = json.NewDecoder(res.Body).Decode(dr)
	if err != nil {
		return nil, err
	}

	return dr, nil
}

func (d *deployRequestsService) List(ctx context.Context, listReq *ListDeployRequestsRequest) ([]*DeployRequest, error) {
	req, err := d.client.newRequest(http.MethodGet, deployRequestsAPIPath(listReq.Organization, listReq.Database), nil)
	if err != nil {
		return nil, errors.Wrap(err, "error creating http request")
	}

	res, err := d.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	deployRequests := &deployRequestsResponse{}
	err = json.NewDecoder(res.Body).Decode(&deployRequests)

	if err != nil {
		return nil, err
	}

	return deployRequests.DeployRequests, nil
}

func (d *deployRequestsService) CreateReview(ctx context.Context, reviewReq *ReviewDeployRequestRequest) (*DeployRequestReview, error) {
	req, err := d.client.newRequest(http.MethodGet, deployRequestActionAPIPath(reviewReq.Organization, reviewReq.Database, reviewReq.Number, "reviews"), reviewReq)
	if err != nil {
		return nil, errors.Wrap(err, "error creating http request")
	}

	res, err := d.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	drr := &DeployRequestReview{}
	err = json.NewDecoder(res.Body).Decode(drr)
	if err != nil {
		return nil, err
	}

	return drr, nil
}

func deployRequestsAPIPath(org, db string) string {
	return fmt.Sprintf("%s/%s/deploy-requests", databasesAPIPath(org), db)
}

// deployRequestAPIPath gets the base path for accessing a single deploy request
func deployRequestAPIPath(org string, db string, number uint64) string {
	return fmt.Sprintf("%s/%s/deploy-requests/%d", databasesAPIPath(org), db, number)
}

func deployRequestActionAPIPath(org string, db string, number uint64, path string) string {
	return fmt.Sprintf("%s/%s", deployRequestAPIPath(org, db, number), path)
}
