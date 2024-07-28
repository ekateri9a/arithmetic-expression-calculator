package handle

import (
	"arithmetic-expression-calculator/internal/config"
	op "arithmetic-expression-calculator/internal/entities"
	"arithmetic-expression-calculator/internal/logger"
	"arithmetic-expression-calculator/internal/models"
	"arithmetic-expression-calculator/internal/utils"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
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
	DB       *sql.DB
	RepoE    []Expression  `json:"expressions"` // todo
	Tasks    []op.Task     `json:"-"`
	mx       *sync.Mutex   `json:"-"`
	conf     config.Config `json:"-"`
	UserName string        `json:"-"`
}

func NewRepo(db *sql.DB) *Repo {
	ctx := context.TODO()
	expressions, _ := models.SelectExpressionsCalculate(ctx, db)
	expressionsRepo := make([]Expression, 0, len(expressions))
	repo := Repo{
		DB:    db,
		mx:    &sync.Mutex{},
		RepoE: expressionsRepo,
	}

	for _, expression := range expressions {
		chunks, _ := utils.CheckExpression(expression.Expression) //todo
		// InfixToPostfix ---------------------------------------------------------
		chunksPostfix := utils.InfixToPostfix(chunks)
		logger.Info("Postfix:", chunksPostfix)

		repo.RepoE = append(repo.RepoE, Expression{
			Id:               int(expression.ID),
			Status:           expression.Status,
			Result:           0,
			ExpressionChunks: chunksPostfix,
		})

		// Find tasks in expression -----------------------------------------------
		repo.FindTask(int(expression.ID), chunksPostfix) // todo int
	}

	return &repo
}

func (repo *Repo) GetAllExpressions() []Expression {
	repo.mx.Lock()
	defer repo.mx.Unlock()
	return repo.RepoE
}

func (repo *Repo) SaveExpression(expression Expression) {
	repo.mx.Lock()
	defer repo.mx.Unlock()
	repo.RepoE = append(repo.RepoE, expression)
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
				repo.Tasks = append(repo.Tasks[:i], repo.Tasks[i+1:]...)
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
				logger.Info("GetTaskAtWork i", repo.Tasks[i].Id, ",", repo.Tasks)
				return repo.Tasks[i], true
			}
		}
	}
	logger.Info("GetTaskAtWork false ", repo.Tasks)
	return op.Task{}, false
}

func (repo *Repo) UpdateExpressionChunks(taskId int, indexReplaceChunk int, calcResultResult float64) bool {
	repo.mx.Lock()
	//defer repo.mx.Unlock()
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

				ctx := context.TODO()
				models.UpdateExpressionFinish(ctx, repo.DB, int64(repo.RepoE[i].Id), repo.RepoE[i].Result)
			}
			// Repeat - add new tasks
			if len(repo.RepoE[i].ExpressionChunks) > 1 {
				// Find tasks
				chunksPostfix := repo.RepoE[i].ExpressionChunks
				repo.mx.Unlock()
				repo.FindTask(taskId, chunksPostfix)

				//countNewTasks := 0
				//for ii := 2; ii < len(repo.RepoE[i].ExpressionChunks); ii++ {
				//	if repo.RepoE[i].ExpressionChunks[ii].OpFlag == op.Operation && repo.RepoE[i].ExpressionChunks[ii-1].OpFlag == op.Num && repo.RepoE[i].ExpressionChunks[ii-2].OpFlag == op.Num {
				//		countNewTasks++
				//		arg1, _ := strconv.ParseFloat(repo.RepoE[i].ExpressionChunks[ii-2].Val, 64)
				//		arg2, _ := strconv.ParseFloat(repo.RepoE[i].ExpressionChunks[ii-1].Val, 64)
				//		repo.Tasks = append(repo.Tasks, op.Task{
				//			Id:            strconv.Itoa(taskId) + "-" + strconv.Itoa(ii-2),
				//			Arg1:          arg1,
				//			Arg2:          arg2,
				//			Operation:     repo.RepoE[i].ExpressionChunks[ii].Val,
				//			OperationTime: 0,
				//		})
				//	}
				//}
				//logger.Info("Added", countNewTasks, "new tasks")
			} else {
				repo.mx.Unlock()
			}
			return true
		}
	}
	repo.mx.Unlock()
	return false
}

func (repo *Repo) FindTask(expressionId int, chunksPostfix []op.Chunk) {
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
	logger.Info("Added", countNewTasks, "new tasks for ", expressionId)
	logger.Info("All tasks", repo.Tasks)
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

	// Add expression to db -------------------------------------------------
	ctx := context.TODO()
	logger.Info("Username ", repo.UserName)
	user, _ := models.SelectUserByLogin(ctx, repo.DB, repo.UserName)
	if err != nil {
		logger.Error("user not found")
	}

	expressionId, err := models.InsertExpression(ctx, repo.DB, &models.Expression{
		Expression: expression.Val,
		UserID:     user.ID,
	})

	if err != nil {
		logger.Error("can not add expression")
	}

	// Add expression to repo -------------------------------------------------
	repo.SaveExpression(Expression{
		Id:               int(expressionId), // todo int
		Status:           "Calculate",
		Result:           0,
		ExpressionChunks: chunksPostfix,
	})

	// Find tasks in expression -----------------------------------------------
	repo.FindTask(int(expressionId), chunksPostfix) // todo int

	// Response --------------------------------------------------------------
	payload := AddExpressionResponse{Id: int(expressionId)} // todo int

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

		// db
		ctx := context.TODO()
		logger.Info("Username expression", repo.UserName)
		user, err := models.SelectUserByLogin(ctx, repo.DB, repo.UserName)
		if err != nil {
			logger.Fatal(err)
			utils.RespondWith500(w) // todo
			return
		}
		expression, err := models.SelectExpressionByUserID(ctx, repo.DB, user.ID, int64(id))
		if err != nil {
			utils.RespondWith404(w)
			return
		}

		//expression, err := repo.GetExpression(id)
		//if err != nil {
		//	utils.RespondWith404(w)
		//	return
		//}

		payload := map[string]models.Expression{"expression": expression}

		respondErr := utils.SuccessRespondWith200(w, payload)
		if respondErr != nil {
			logger.Error(respondErr)
		}
		return
	}

	// db
	ctx := context.TODO()
	logger.Info("Username expressions", repo.UserName)
	user, err := models.SelectUserByLogin(ctx, repo.DB, repo.UserName)
	if err != nil {
		logger.Info(err)
		utils.RespondWith500(w) // todo
		return
	}
	expressions, err := models.SelectExpressionsByUserID(ctx, repo.DB, user.ID)
	if err != nil {
		utils.RespondWith404(w)
		return
	}
	//------

	payload := map[string][]models.Expression{"expressions": expressions}
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

func (repo *Repo) AddRegistration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondWith400(w, "http method must be POST")
		return
	}

	user := new(op.User)
	err := utils.DecodeBody(w, r, user)
	if err != nil {
		logger.Error("failed to decode body", err)
		if err = utils.RespondWith400(w, "failed to decode body"); err != nil {
			logger.Error(err)
		}
		return
	}
	defer r.Body.Close()

	// todo check password and login not empty

	// todo add in db
	ctx := context.TODO()
	id, err := models.InsertUser(ctx, repo.DB, &models.User{Login: user.Login, Password: user.Password})
	if err != nil {
		utils.RespondWith500(w) // todo
		return
	}
	logger.Info("New user with id:", id, " - ", user)

	utils.Respond200(w)
	return
}

func (repo *Repo) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.RespondWith400(w, "http method must be POST")
		return
	}

	user := new(op.User)
	err := utils.DecodeBody(w, r, user)
	if err != nil {
		logger.Error("failed to decode body", err)
		if err = utils.RespondWith400(w, "failed to decode body"); err != nil {
			logger.Error(err)
		}
		return
	}
	defer r.Body.Close()

	ctx := context.TODO()
	dbUser, err := models.SelectUserByLogin(ctx, repo.DB, user.Login)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "wrong login")
		return
	}
	if dbUser.Password != user.Password {
		utils.RespondWithError(w, http.StatusUnauthorized, "wrong password")
		return
	}

	// New JWT
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"name": user.Login, // todo
		"nbf":  now.Unix(),
		"exp":  now.Add(24 * time.Hour).Unix(),
		"iat":  now.Unix(),
	})

	tokenString, err := token.SignedString([]byte(repo.conf.SecretKey))
	if err != nil {
		logger.Fatal("JWT token fail")
		utils.RespondWith500(w)
	}
	logger.Info("Token: ", tokenString)

	// add JWT
	payload := map[string]string{"token": tokenString}
	respondErr := utils.SuccessRespondWith200(w, payload)
	if respondErr != nil {
		logger.Error(respondErr)
	}
	return
}

func (repo *Repo) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("AccessToken")

		logger.Info("Token: ", tokenString)

		tokenFromString, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				panic(fmt.Errorf("unexpected signing method: %v", token.Header["alg"]))
			}

			return []byte(repo.conf.SecretKey), nil
		})

		if err != nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "bad token, repeat auth")
			return
		}

		if claims, ok := tokenFromString.Claims.(jwt.MapClaims); ok {
			fmt.Println("user name: ", claims["name"])
			repo.UserName = claims["name"].(string)
		} else {
			utils.RespondWithError(w, http.StatusUnauthorized, "bad token, repeat auth")
			return
		}
		next.ServeHTTP(w, r)
	})
}
