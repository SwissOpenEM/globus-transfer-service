package tasks

type transferTask struct {
	scicatUrl    *string
	globusTaskId string
	datasetPid   string
}

func (t transferTask) execute() {

}
