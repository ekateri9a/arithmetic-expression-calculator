package entities

type OpFlag int64

const (
	Num       OpFlag = 0
	Operation OpFlag = 1
)

type Chunk struct {
	OpFlag OpFlag
	Val    string
}
