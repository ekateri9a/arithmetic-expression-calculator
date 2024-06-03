package entities

import "fmt"

type Task struct {
	Id            string  `json:"id"`
	Arg1          float64 `json:"arg1"`
	Arg2          float64 `json:"arg2"`
	Operation     string  `json:"operation"`
	OperationTime int     `json:"operation_time"`
	AtWork        bool    `json:"-"`
}

func (t *Task) String() string {
	return fmt.Sprintf("Id: %s, Arg1: %g, Arg2: %g, Operation: %s, OperationTime: %d", t.Id, t.Arg1, t.Arg2, t.Operation, t.OperationTime)
}
