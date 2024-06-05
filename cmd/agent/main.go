package main

import (
	"arithmetic-expression-calculator/internal/config"
	"arithmetic-expression-calculator/internal/entities"
	"arithmetic-expression-calculator/internal/logger"
	"arithmetic-expression-calculator/internal/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"
)

type TaskResult struct {
	Id     string  `json:"id"`
	Result float64 `json:"result"`
}

func (t *TaskResult) String() string {
	return fmt.Sprintf("Id: %s, Result: %g", t.Id, t.Result)
}

func GetTimeOperation(taskOperation string, conf *config.Config) time.Duration {
	var t time.Duration
	switch taskOperation {
	case "*":
		t = time.Duration(conf.TimeMultiplicationsMs) * time.Millisecond
	case "/":
		t = time.Duration(conf.TimeDivisionsMs) * time.Millisecond
	case "+":
		t = time.Duration(conf.TimeAdditionMs) * time.Millisecond
	case "-":
		t = time.Duration(conf.TimeSubtractionMs) * time.Millisecond
	case "^":
		t = time.Duration(conf.TimeExponentiationMs) * time.Millisecond
	}
	return t
}

func Calculate(cl chan bool, tasks chan entities.Task, results chan TaskResult, conf *config.Config) {
	for {
		select {
		case <-cl:
			logger.Info("close calculate goroutines")
			return
		case task := <-tasks:
			result := TaskResult{}
			result.Id = task.Id
			var t time.Duration

			// Get time
			if task.OperationTime == 0 {
				t = GetTimeOperation(task.Operation, conf)
			} else {
				t = time.Duration(task.OperationTime) * time.Millisecond
			}
			task.OperationTime = int(t)

			logger.Info("new task:", task.String())

			// Calculate
			switch task.Operation {
			case "*":
				result.Result = task.Arg1 * task.Arg2
			case "/":
				result.Result = task.Arg1 / task.Arg2
			case "+":
				result.Result = task.Arg1 + task.Arg2
			case "-":
				result.Result = task.Arg1 - task.Arg2
			case "^":
				result.Result = math.Pow(task.Arg1, task.Arg2)
			}

			time.Sleep(t)
			results <- result
		}
	}
}

func Run() {
	logger.Info("start main agent")
	logger.Info("reading config...")
	conf, err := config.LoadFromEnv()
	if err != nil {
		logger.Error("failed to read config")
		os.Exit(1)
	}

	countGoroutines := conf.CountGoroutines
	logger.Info("count goroutines: ", countGoroutines)

	cl := make(chan bool)
	tasks := make(chan entities.Task)
	results := make(chan TaskResult)

	for i := 0; i < countGoroutines; i++ {
		go Calculate(cl, tasks, results, conf)
	}

	client := &http.Client{}
	serverPort := conf.ServerPort
	countRepeatConnection := 10
	timeToWaitConnectionSeconds := 5

	for {
		select {
		// GET /internal/task
		case <-time.After(1 * time.Second):

			resp, err := client.Get("http://localhost:" + strconv.Itoa(serverPort) + "/internal/task")
			if err != nil {
				countRepeatConnection--
				logger.Info("connection error: ", err)
				if countRepeatConnection == 0 {
					return
				}
				logger.Info("sleep", timeToWaitConnectionSeconds, "seconds before repeat connection")
				logger.Info("you have", countRepeatConnection, "attempt to repeat connection")
				time.Sleep(time.Duration(timeToWaitConnectionSeconds) * time.Second)
				continue
			}

			if resp.StatusCode == 200 {
				task := new(entities.Task)
				utils.DecodeRespondBody(resp.Body, task)
				defer resp.Body.Close()
				if err != nil {
					logger.Error("failed to decode body", err)
				} else {
					go func() {
						tasks <- *task
					}()
				}
			} else {
				logger.Info("StatusCode: ", resp.StatusCode)
			}
		// POST /internal/task
		case r := <-results:
			logger.Info("new result: ", r.String())
			body, err := json.Marshal(r)
			if err != nil {
				logger.Error("failed to marshall payload:", r.String())
			} else {
				resp, err := client.Post("http://localhost:"+strconv.Itoa(serverPort)+"/internal/task", "application/json", bytes.NewBuffer(body))
				if err != nil {
					logger.Error(err)
				}
				resp.Body.Close()
				logger.Info("StatusCode: ", resp.StatusCode)
			}
		}
	}
}

func main() {
	Run()
}
