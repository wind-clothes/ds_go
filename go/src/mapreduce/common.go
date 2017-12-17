package mapreduce

import (
	"fmt"
	"strconv"
)

// Debugging enabled?
const (
	debugEnabled          = true
	mapPhase     jobPhase = "Map"
	reducePhase  jobPhase = "Reduce"
)

// jobPhase indicates whether a task is scheduled as a map or reduce task.
// jobPhase指示是否将task转化为map or reduce task.。
type jobPhase string

// KeyValue is a type used to hold the key/value pairs passed to the map and
// reduce functions.
type KeyValue struct {
	Key   string
	Value string
}

// DPrintf will only print if the debugEnabled const has been set to true
// interface{} 理解有点像Java中的Object对象，
func debug(format string, a ...interface{}) (n int, err error) {
	if debugEnabled {
		n, err = fmt.Printf(format, a)
	}
	return
}

// reduceName constructs the name of the intermediate file which map task
// <mapTask> produces for reduce task <reduceTask>.
func reduceName(jobName string, mapTask int, reduceTask int) string {
	return "mrtmp." + jobName + "-" + strconv.Itoa(mapTask) + "-" + strconv.Itoa(reduceTask)
}

// mergeName constructs the name of the output file of reduce task <reduceTask>
func mergeName(jobName string, reduceTask int) string {
	return "mrtmp." + jobName + "-res-" + strconv.Itoa(reduceTask)
}

/*func main() {
	str := reduceName("jobname", 1, 1)
	fmt.Printf("aaaaaaaaaa:%v", str)
}*/
