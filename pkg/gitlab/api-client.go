package gitlab

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"strconv"
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
const includeDivergedCommits = "include_diverged_commits_count"
const includeRebaseInProgress = "include_rebase_in_progress"
const MergeRequestsEndpoint = "/api/v4/projects/{" + projectIdParam + "}/merge_requests"
const MergeRequestsMergeEndpoint = "/api/v4/projects/{" + projectIdParam + "}/merge_requests/{" + mergeRequestIdParam + "}/merge"
const MergeRequestsDetailsEndpoint = "/api/v4/projects/{" + projectIdParam + "}/merge_requests/{" + mergeRequestIdParam + "}"
const MergeRequestsEventsEndpoint = "/api/v4/projects/{" + projectIdParam + "}/merge_requests/{" + mergeRequestIdParam + "}/notes"
const BranchesEndpoint = "/api/v4/projects/{" + projectIdParam + "}/repository/branches"
const DeleteBranchEndpoint = "/api/v4/projects/{" + projectIdParam + "}/repository/branches/{" + branchIdParam + "}"

type ApiClient struct {
	resty        *resty.Client
	projectName  string
	userName     string
	BranchPrefix string
	apiToken     string
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
}

type MergeRequestNote struct {
	MergeRequestIid string `json:"noteable_iid"`
	Body            string `json:"body"`
}

type CommitDetails struct {
	AuthoredDate string `json:"authored_date"`
	Message      string `json:"message"`
}
type Branch struct {
	Name   string        `json:"name"`
	Commit CommitDetails `json:"commit"`
}

func (client *ApiClient) ListMergeRequests() []MergeRequestDetails {
	var mergeRequests []MergeRequestDetails
	_, err := client.resty.R().
		SetResult(&mergeRequests).
		SetQueryParam("author_username", client.userName).
		SetQueryParam("state", "opened").
		SetPathParam(projectIdParam, client.projectName).
		Get(MergeRequestsEndpoint)

	if err != nil {
		fmt.Println("Error when executing query." + err.Error())
	}
	return mergeRequests
}

func (client *ApiClient) ListMergeRequestNotes(mergeRequestIid int) []MergeRequestNote {
	var notes []MergeRequestNote
	_, err := client.resty.R().
		SetResult(&notes).
		SetPathParam(projectIdParam, client.projectName).
		SetPathParam(mergeRequestIdParam, strconv.Itoa(mergeRequestIid)).
		Get(MergeRequestsEventsEndpoint)

	if err != nil {
		fmt.Println("Error when executing query." + err.Error())
	}

	return notes
}

func (client *ApiClient) ListBranches() []Branch {
	var branches []Branch
	_, err := client.resty.R().
		SetResult(&branches).
		SetPathParam(projectIdParam, client.projectName).
		Get(BranchesEndpoint)

	if err != nil {
		fmt.Println("Error when executing query." + err.Error())
	}

	return branches
}

func (client *ApiClient) DeleteBranch(branchName string) {
	_, err := client.resty.R().
		SetPathParam(projectIdParam, client.projectName).
		SetPathParam(branchIdParam, branchName).
		Delete(DeleteBranchEndpoint)

	if err != nil {
		fmt.Println("Error when executing query." + err.Error())
	}
}

func (client *ApiClient) CreateMergeRequest(sourceBranch string, targetBranch string, title string) int {
	var result MergeRequestDetails
	_, err := client.resty.R().
		SetPathParam(projectIdParam, client.projectName).
		SetQueryParam(sourceBranchParam, sourceBranch).
		SetQueryParam(targetBranchParam, targetBranch).
		SetQueryParam(removeSourceBranchParam, "true").
		SetQueryParam(titleParam, title).
		SetResult(&result).
		Post(MergeRequestsEndpoint)

	if err != nil {
		fmt.Println("Error when executing query." + err.Error())
	}

	return result.Iid
}

func (client *ApiClient) mergeMergeRequest(mergeRequestIid int) MergeRequestDetails {
	var mergeRequest MergeRequestDetails
	_, err := client.resty.R().
		SetPathParam(projectIdParam, client.projectName).
		SetPathParam(mergeRequestIdParam, strconv.Itoa(mergeRequestIid)).
		SetQueryParam(mergeWhenPipelineSucceeds, "true").
		SetQueryParam(shouldRemoveSourceBranch, "true").
		SetResult(&mergeRequest).
		Put(MergeRequestsMergeEndpoint)
	if err != nil {
		fmt.Println("Error when executing query." + err.Error())
	}

	return mergeRequest
}

func (client *ApiClient) getMergeRequestDetails(mergeRequestIid int) MergeRequestDetails {
	var mergeRequest MergeRequestDetails
	_, err := client.resty.R().
		SetPathParam(projectIdParam, client.projectName).
		SetPathParam(mergeRequestIdParam, strconv.Itoa(mergeRequestIid)).
		SetQueryParam(includeDivergedCommits, "true").
		SetQueryParam(includeRebaseInProgress, "true").
		SetResult(&mergeRequest).
		Get(MergeRequestsDetailsEndpoint)

	if err != nil {
		fmt.Println("Error when executing query." + err.Error())
	}
	return mergeRequest
}

func New(gitlabUrl string, projectName string, branchPrefix string, userName string, apiToken string) *ApiClient {
	client := &ApiClient{
		resty:        createClient(gitlabUrl, apiToken),
		projectName:  projectName,
		BranchPrefix: branchPrefix,
		userName:     userName,
		apiToken:     apiToken,
	}
	return client
}
func createClient(gitlabUrl string, apiToken string) *resty.Client {
	client := resty.New()
	client.SetBaseURL(gitlabUrl)
	client.SetHeader(tokenHeader, apiToken)
	return client
}
