package handle

import (
	op "arithmetic-expression-calculator/internal/entities"
	"arithmetic-expression-calculator/internal/logger"
	"arithmetic-expression-calculator/internal/utils"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type AddExpressionResponse struct {
	Id int `json:"id"`
}

type Expression struct {
	Id               int        `json:"id"`
	Status           string     `json:"status"`
	Result           float64    `json:"result"`
	ExpressionChunks []op.Chunk `json:"-"`
}

type Repo struct {
	RepoE []Expression `json:"expressions"`
	Tasks []op.Task    `json:"-"`
	mx    *sync.Mutex  `json:"-"`
}

func NewRepo() *Repo {
	return &Repo{
		mx: &sync.Mutex{},
	}
}

func (repo *Repo) GetAllExpressions() []Expression {
	repo.mx.Lock()
	defer repo.mx.Unlock()
	return repo.RepoE
}

func (repo *Repo) SaveExpression(expression Expression) int {
	repo.mx.Lock()
	defer repo.mx.Unlock()
	expression.Id = len(repo.RepoE) + 1
	repo.RepoE = append(repo.RepoE, expression)
	return expression.Id
}

func (repo *Repo) GetExpression(id int) (Expression, error) {

	for _, expression := range repo.RepoE {
		if expression.Id == id {
			return expression, nil
		}
	}
	return Expression{}, errors.New("wrong id")
}

func (repo *Repo) SaveTask(task op.Task) {
	repo.mx.Lock()
	defer repo.mx.Unlock()
	repo.Tasks = append(repo.Tasks, task)
}

func (repo *Repo) DeleteTask(taskId string) bool {
	repo.mx.Lock()
	defer repo.mx.Unlock()
	for i := 0; i < len(repo.Tasks); i++ {
		if repo.Tasks[i].Id == taskId {
			if len(repo.Tasks) == 1 {
				repo.Tasks = []op.Task{}
			} else if i == len(repo.Tasks)-1 {
				repo.Tasks = repo.Tasks[:i]
			} else {
				repo.Tasks = append(repo.Tasks[:i], repo.Tasks[:i+1]...)
			}
			return true
		}
	}
	return false
}

func (repo *Repo) GetTaskAtWork() (op.Task, bool) {
	repo.mx.Lock()
	defer repo.mx.Unlock()
	if len(repo.Tasks) > 0 {
		for i := 0; i < len(repo.Tasks); i++ {
			if !repo.Tasks[i].AtWork {
				repo.Tasks[i].AtWork = true
				return repo.Tasks[i], true
			}
		}
	}
	return op.Task{}, false
}

func (repo *Repo) UpdateExpressionChunks(taskId int, indexReplaceChunk int, calcResultResult float64) bool {
	for i := 0; i < len(repo.RepoE); i++ {
		if repo.RepoE[i].Id == taskId {
			logger.Info("Postfix before replace: ", repo.RepoE[i].ExpressionChunks)

			repo.RepoE[i].ExpressionChunks = append(repo.RepoE[i].ExpressionChunks[:indexReplaceChunk], repo.RepoE[i].ExpressionChunks[indexReplaceChunk+2:]...)

			repo.RepoE[i].ExpressionChunks[indexReplaceChunk].Val = fmt.Sprintf("%g", calcResultResult)
			repo.RepoE[i].ExpressionChunks[indexReplaceChunk].OpFlag = op.Num
			logger.Info("Postfix after replace: ", repo.RepoE[i].ExpressionChunks)

			if len(repo.RepoE[i].ExpressionChunks) == 1 {
				repo.RepoE[i].Status = "Finished"
				err := errors.New("")
				repo.RepoE[i].Result, err = strconv.ParseFloat(repo.RepoE[i].ExpressionChunks[0].Val, 64)
				if err != nil {
					logger.Error(err)
				}
			}
			// Repeat - add new tasks
			if len(repo.RepoE[i].ExpressionChunks) > 1 {
				// Find tasks
				countNewTasks := 0
				for ii := 2; ii < len(repo.RepoE[i].ExpressionChunks); ii++ {
					if repo.RepoE[i].ExpressionChunks[ii].OpFlag == op.Operation && repo.RepoE[i].ExpressionChunks[ii-1].OpFlag == op.Num && repo.RepoE[i].ExpressionChunks[ii-2].OpFlag == op.Num {
						countNewTasks++
						arg1, _ := strconv.ParseFloat(repo.RepoE[i].ExpressionChunks[ii-2].Val, 64)
						arg2, _ := strconv.ParseFloat(repo.RepoE[i].ExpressionChunks[ii-1].Val, 64)
						repo.Tasks = append(repo.Tasks, op.Task{
							Id:            strconv.Itoa(taskId) + "-" + strconv.Itoa(ii-2),
							Arg1:          arg1,
							Arg2:          arg2,
							Operation:     repo.RepoE[i].ExpressionChunks[ii].Val,
							OperationTime: 0,
						})
					}
				}
				logger.Info("Added", countNewTasks, "new tasks")
			}

			return true
		}
	}
	return false
}

func (repo *Repo) AddExpressionHandleFunc(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondWith400(w, "http method must be POST")
		return
	}

	expression := new(op.Expression)
	err := utils.DecodeBody(w, r, expression)
	if err != nil {
		logger.Error("failed to decode body", err)
		if err = utils.RespondWith400(w, "failed to decode body"); err != nil {
			logger.Error(err)
		}
		return
	}
	defer r.Body.Close()

	// CheckExpression ---------------------------------------------------------
	chunks, e := utils.CheckExpression(expression.Val)
	if e != nil {
		utils.RespondWith422(w)
		return
	}

	// InfixToPostfix ---------------------------------------------------------
	chunksPostfix := utils.InfixToPostfix(chunks)
	logger.Info("Postfix:", chunksPostfix)

	// Add expression to repo -------------------------------------------------
	expressionId := repo.SaveExpression(Expression{
		Status:           "Calculate",
		Result:           0,
		ExpressionChunks: chunksPostfix,
	})

	// Find tasks in expression -----------------------------------------------
	countNewTasks := 0
	for i := 2; i < len(chunksPostfix); i++ {
		if chunksPostfix[i].OpFlag == op.Operation && chunksPostfix[i-1].OpFlag == op.Num && chunksPostfix[i-2].OpFlag == op.Num {
			countNewTasks++
			arg1, _ := strconv.ParseFloat(chunksPostfix[i-2].Val, 64)
			arg2, _ := strconv.ParseFloat(chunksPostfix[i-1].Val, 64)
			repo.SaveTask(op.Task{
				Id:            strconv.Itoa(expressionId) + "-" + strconv.Itoa(i-2),
				Arg1:          arg1,
				Arg2:          arg2,
				Operation:     chunksPostfix[i].Val,
				OperationTime: 0,
			})
		}
	}
	logger.Info("Added", countNewTasks, "new tasks")

	// Response --------------------------------------------------------------
	payload := AddExpressionResponse{Id: expressionId}

	respondErr := utils.SuccessRespondWith201(w, payload)
	if respondErr != nil {
		logger.Error(respondErr)
	}

}

func (repo *Repo) GetExpressionsHandleFunc(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.String(), "/expressions/:") {
		s, _ := strings.CutPrefix(r.URL.String(), "/expressions/:")

		id, err := strconv.Atoi(s)
		if err != nil {
			utils.RespondWith404(w)
			return
		}

		expression, err := repo.GetExpression(id)
		if err != nil {
			utils.RespondWith404(w)
			return
		}

		payload := map[string]Expression{"expression": expression}

		respondErr := utils.SuccessRespondWith200(w, payload)
		if respondErr != nil {
			logger.Error(respondErr)
		}
		return
	}

	payload := repo
	respondErr := utils.SuccessRespondWith200(w, payload)
	if respondErr != nil {
		logger.Error(respondErr)
	}

}

func (repo *Repo) TaskHandleFunc(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		logger.Info("Task POST")

		calcResult := new(op.CalcResult)
		err := utils.DecodeBody(w, r, calcResult)
		if err != nil {
			logger.Error("failed to decode body", err)
			if err = utils.RespondWith400(w, "failed to decode body"); err != nil {
				logger.Error(err)
			}
			return
		}
		defer r.Body.Close()

		// Update postfix ---------------------------------------------------
		idArr := strings.Split(calcResult.Id, "-")
		taskId, _ := strconv.Atoi(idArr[0])
		indexReplaceChunk, _ := strconv.Atoi(idArr[1])

		// Find and delete task
		flagCorrectId := repo.DeleteTask(calcResult.Id)

		if !flagCorrectId {
			if err = utils.RespondWith404(w); err != nil {
				logger.Error(err)
			}
			return
		}

		// Update postfix
		ok := repo.UpdateExpressionChunks(taskId, indexReplaceChunk, calcResult.Result)

		if ok {
			utils.Respond200(w)
			return
		} else {
			utils.RespondWith500(w)
			return
		}

	}

	if r.Method == "" || r.Method == http.MethodGet {
		logger.Info("Task GET")
		payload, ok := repo.GetTaskAtWork()
		if ok {
			respondErr := utils.SuccessRespondWith200(w, payload)
			if respondErr != nil {
				logger.Error(respondErr)
			}
			return
		}

		utils.RespondWith404(w)
		return
	}
}
