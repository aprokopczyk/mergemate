package gitlab

import (
	"fmt"
	"github.com/go-resty/resty/v2"
)

const projectIdParam = "projectId"
const branchIdParam = "branchId"
const sourceBranchParam = "source_branch"
const targetBranchParam = "target_branch"
const titleParam = "title"
const tokenHeader = "PRIVATE-TOKEN"
const MergeRequestsEndpoint = "/api/v4/projects/{" + projectIdParam + "}/merge_requests"
const BranchesEndpoint = "/api/v4/projects/{" + projectIdParam + "}/repository/branches"
const DeleteBranchEndpoint = "/api/v4/projects/{" + projectIdParam + "}/repository/branches/{" + branchIdParam + "}"

type ApiClient struct {
	resty        *resty.Client
	projectName  string
	userName     string
	branchPrefix string
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
}

type CommitDetails struct {
	AuthoredDate string `json:"authored_date"`
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
		SetPathParam(projectIdParam, client.projectName).
		Get(MergeRequestsEndpoint)

	if err != nil {
		fmt.Println("Error when executing query." + err.Error())
	}
	return mergeRequests
}

func (client *ApiClient) ListBranches() []Branch {
	var branches []Branch
	_, err := client.resty.R().
		SetResult(&branches).
		SetQueryParam("search", "^"+client.branchPrefix).
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

func (client *ApiClient) CreateMergeRequest(sourceBranch string, targetBranch string, title string) {
	_, err := client.resty.R().
		SetPathParam(projectIdParam, client.projectName).
		SetQueryParam(sourceBranchParam, sourceBranch).
		SetQueryParam(targetBranchParam, targetBranch).
		SetQueryParam(titleParam, title).
		Post(MergeRequestsEndpoint)

	if err != nil {
		fmt.Println("Error when executing query." + err.Error())
	}

}

func New(gitlabUrl string, projectName string, branchPrefix string, userName string, apiToken string) *ApiClient {
	client := &ApiClient{
		resty:        createClient(gitlabUrl, apiToken),
		projectName:  projectName,
		branchPrefix: branchPrefix,
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
