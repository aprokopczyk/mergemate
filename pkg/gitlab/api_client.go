package gitlab

import (
	"errors"
	"github.com/go-resty/resty/v2"
	"log"
	"sort"
	"strconv"
	"time"
)

const projectIdParam = "projectId"
const branchIdParam = "branchId"
const sourceBranchParam = "source_branch"
const targetBranchParam = "target_branch"
const removeSourceBranchParam = "remove_source_branch"
const titleParam = "title"
const tokenHeader = "PRIVATE-TOKEN"
const mergeRequestIdParam = "merge_request_iid"
const mergeWhenPipelineSucceeds = "merge_when_pipeline_succeeds"
const shouldRemoveSourceBranch = "should_remove_source_branch"
const sha = "sha"
const skipCi = "skip_ci"
const includeDivergedCommits = "include_diverged_commits_count"
const includeRebaseInProgress = "include_rebase_in_progress"
const MergeRequestsEndpoint = "/api/v4/projects/{" + projectIdParam + "}/merge_requests"
const MergeRequestsMergeEndpoint = "/api/v4/projects/{" + projectIdParam + "}/merge_requests/{" + mergeRequestIdParam + "}/merge"
const MergeRequestsDetailsEndpoint = "/api/v4/projects/{" + projectIdParam + "}/merge_requests/{" + mergeRequestIdParam + "}"
const MergeRequestsRebaseEndpoint = "/api/v4/projects/{" + projectIdParam + "}/merge_requests/{" + mergeRequestIdParam + "}/rebase"
const MergeRequestsEventsEndpoint = "/api/v4/projects/{" + projectIdParam + "}/merge_requests/{" + mergeRequestIdParam + "}/notes"
const MergeRequestsPipelinesEndpoint = "/api/v4/projects/{" + projectIdParam + "}/merge_requests/{" + mergeRequestIdParam + "}/pipelines"
const BranchesEndpoint = "/api/v4/projects/{" + projectIdParam + "}/repository/branches"
const DeleteBranchEndpoint = "/api/v4/projects/{" + projectIdParam + "}/repository/branches/{" + branchIdParam + "}"

type ApiClient struct {
	resty       *resty.Client
	projectName string
	userName    string
	apiToken    string
}
type MergeRequestDetails struct {
	Id                        int    `json:"id"`
	Iid                       int    `json:"iid"`
	Title                     string `json:"title"`
	State                     string `json:"state"`
	TargetBranch              string `json:"target_branch"`
	SourceBranch              string `json:"source_branch"`
	MergeWhenPipelineSucceeds bool   `json:"merge_when_pipeline_succeeds"`
	MergeStatus               string `json:"merge_status"`
	DetailedMergeStatus       string `json:"detailed_merge_status"`
	HasConflicts              bool   `json:"has_conflicts"`
	ShouldRemoveSourceBranch  bool   `json:"should_remove_source_branch"`
	CommitsBehind             int    `json:"diverged_commits_count"`
	Sha                       string `json:"sha"`
	RebaseInProgress          bool   `json:"rebase_in_progress"`
	RebaseError               string `json:"merge_error"`
}

type MergeRequestNote struct {
	MergeRequestIid int    `json:"noteable_iid"`
	Body            string `json:"body"`
}

type CommitDetails struct {
	AuthoredDate time.Time `json:"authored_date"`
	Message      string    `json:"message"`
}
type Branch struct {
	Name    string        `json:"name"`
	Default bool          `json:"default"`
	Commit  CommitDetails `json:"commit"`
}

var MergeRequestAlreadyExists = errors.New("merge request already exists")

func (client *ApiClient) OpenedMergeRequests() ([]MergeRequestDetails, error) {
	return client.ListMergeRequests("opened")
}

func (client *ApiClient) MergedMergeRequests() ([]MergeRequestDetails, error) {
	return client.ListMergeRequests("merged")
}

func (client *ApiClient) ListMergeRequests(state string) ([]MergeRequestDetails, error) {
	var mergeRequests []MergeRequestDetails
	_, err := client.resty.R().
		SetResult(&mergeRequests).
		SetQueryParam("author_username", client.userName).
		SetQueryParam("state", state).
		SetQueryParam("per_page", "100").
		SetPathParam(projectIdParam, client.projectName).
		Get(MergeRequestsEndpoint)
	if err != nil {
		return nil, err
	}
	return mergeRequests, nil
}

func (client *ApiClient) ListMergeRequestNotes(mergeRequestIid int) ([]MergeRequestNote, error) {
	var notes []MergeRequestNote
	_, err := client.resty.R().
		SetResult(&notes).
		SetPathParam(projectIdParam, client.projectName).
		SetPathParam(mergeRequestIdParam, strconv.Itoa(mergeRequestIid)).
		Get(MergeRequestsEventsEndpoint)

	if err != nil {
		return nil, err
	}

	return notes, nil
}

func (client *ApiClient) CreateMergeRequestNote(mergeRequestIid int, noteBody string) error {
	var note MergeRequestNote
	_, err := client.resty.R().
		SetResult(&note).
		SetPathParam(projectIdParam, client.projectName).
		SetPathParam(mergeRequestIdParam, strconv.Itoa(mergeRequestIid)).
		SetQueryParam("body", noteBody).
		Post(MergeRequestsEventsEndpoint)

	if err != nil {
		return err
	}

	return nil
}

func (client *ApiClient) listBranches(namePatterns []string) ([]Branch, error) {
	var result []Branch

	for _, pattern := range namePatterns {
		var branches []Branch
		_, err := client.resty.R().
			SetResult(&branches).
			SetPathParam(projectIdParam, client.projectName).
			SetQueryParam("search", "^"+pattern).
			SetQueryParam("per_page", "100").
			Get(BranchesEndpoint)
		if err != nil {
			return nil, err
		}

		result = append(result, branches...)
	}

	return result, nil
}

func (client *ApiClient) FetchBranchesWithPattern(patterns []string) []Branch {
	branches, err := client.listBranches(patterns)

	if err != nil {
		log.Printf("Error when fetching branches list %v", err)
	}

	sort.SliceStable(branches, func(i, j int) bool {
		return branches[i].Commit.AuthoredDate.Unix() > branches[j].Commit.AuthoredDate.Unix()
	})

	return branches
}

func (client *ApiClient) DeleteBranch(branchName string) error {
	_, err := client.resty.R().
		SetPathParam(projectIdParam, client.projectName).
		SetPathParam(branchIdParam, branchName).
		Delete(DeleteBranchEndpoint)

	if err != nil {
		return err
	}
	return nil
}

func (client *ApiClient) CreateMergeRequest(sourceBranch string, targetBranch string, title string) (*MergeRequestDetails, error) {
	var result MergeRequestDetails
	resp, err := client.resty.R().
		SetPathParam(projectIdParam, client.projectName).
		SetQueryParam(sourceBranchParam, sourceBranch).
		SetQueryParam(targetBranchParam, targetBranch).
		SetQueryParam(removeSourceBranchParam, "true").
		SetQueryParam(titleParam, title).
		SetResult(&result).
		Post(MergeRequestsEndpoint)

	resp.Status()
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() == 409 {
		return nil, MergeRequestAlreadyExists
	}

	return &result, nil
}

func (client *ApiClient) MergeMergeRequest(mergeRequestIid int, currentSha string) (*MergeRequestDetails, error) {
	var mergeRequest MergeRequestDetails
	_, err := client.resty.R().
		SetPathParam(projectIdParam, client.projectName).
		SetPathParam(mergeRequestIdParam, strconv.Itoa(mergeRequestIid)).
		SetQueryParam(mergeWhenPipelineSucceeds, "false").
		SetQueryParam(shouldRemoveSourceBranch, "true").
		SetQueryParam(sha, currentSha).
		SetResult(&mergeRequest).
		Put(MergeRequestsMergeEndpoint)
	if err != nil {
		return nil, err
	}

	return &mergeRequest, nil
}

func (client *ApiClient) GetMergeRequestDetails(mergeRequestIid int) (*MergeRequestDetails, error) {
	var mergeRequest MergeRequestDetails
	_, err := client.resty.R().
		SetPathParam(projectIdParam, client.projectName).
		SetPathParam(mergeRequestIdParam, strconv.Itoa(mergeRequestIid)).
		SetQueryParam(includeDivergedCommits, "true").
		SetQueryParam(includeRebaseInProgress, "true").
		SetResult(&mergeRequest).
		Get(MergeRequestsDetailsEndpoint)

	if err != nil {
		return nil, err
	}
	return &mergeRequest, nil
}

func (client *ApiClient) RebaseMergeRequest(mergeRequestIid int, shouldSkipCi bool) error {
	var mergeRequest MergeRequestDetails
	_, err := client.resty.R().
		SetPathParam(projectIdParam, client.projectName).
		SetPathParam(mergeRequestIdParam, strconv.Itoa(mergeRequestIid)).
		SetQueryParam(skipCi, strconv.FormatBool(shouldSkipCi)).
		SetResult(&mergeRequest).
		Put(MergeRequestsRebaseEndpoint)
	return err
}

type MergeRequestPipeline struct {
	Id        int       `json:"id"`
	Sha       string    `json:"sha"`
	Ref       string    `json:"ref"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

func (client *ApiClient) GetMergeRequestPipelines(mergeRequestIid int) ([]MergeRequestPipeline, error) {
	var pipelines []MergeRequestPipeline
	_, err := client.resty.R().
		SetResult(&pipelines).
		SetPathParam(projectIdParam, client.projectName).
		SetPathParam(mergeRequestIdParam, strconv.Itoa(mergeRequestIid)).
		SetQueryParam("order_by", "id").
		SetQueryParam("sort", "asc").
		Get(MergeRequestsPipelinesEndpoint)

	if err != nil {
		return nil, err
	}

	sort.SliceStable(pipelines, func(i, j int) bool {
		return pipelines[i].CreatedAt.Unix() > pipelines[j].CreatedAt.Unix()
	})
	return pipelines, nil
}

func New(gitlabUrl string, projectName string, userName string, apiToken string) *ApiClient {
	client := &ApiClient{
		resty:       createClient(gitlabUrl, apiToken),
		projectName: projectName,
		userName:    userName,
		apiToken:    apiToken,
	}
	return client
}
func createClient(gitlabUrl string, apiToken string) *resty.Client {
	client := resty.New()
	client.SetBaseURL(gitlabUrl)
	client.SetHeader(tokenHeader, apiToken)
	client.SetTimeout(time.Second * 10)
	return client
}
