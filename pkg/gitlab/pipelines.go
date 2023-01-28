package gitlab

type void struct{}

var present = void{}
var runningStates = map[string]void{
	"created":              present,
	"waiting_for_resource": present,
	"preparing":            present,
	"pending":              present,
	"running":              present,
}

func IsPipelineRunning(pipelines []MergeRequestPipeline) bool {
	running := false
	for _, pipeline := range pipelines {
		_, exists := runningStates[pipeline.Status]
		if exists {
			running = true
		}
	}
	return running
}
func IsAutomaticMergeAllowed(pipelines []MergeRequestPipeline) bool {
	for _, pipeline := range pipelines {
		if pipeline.Status == "skipped" {
			continue
		} else if pipeline.Status == "success" {
			return true
		} else {
			return false
		}
	}
	return false
}
