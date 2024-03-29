package warmer

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
)

var (
	// Warm indicates if the Lambda is warm
	Warm = false

	// LastAccess indicates the last Lambda access
	LastAccess = time.Time{}

	funcName    = os.Getenv("AWS_LAMBDA_FUNCTION_NAME")
	funcVersion = os.Getenv("AWS_LAMBDA_FUNCTION_VERSION")
)

const (
	defaultConcurrency         = 1
	defaultInvocation          = 1
	defaultDelayInMilliSeconds = 125
)

// Log defines log message
type Log struct {
	Action                string    `json:"action"`
	Function              string    `json:"function"`
	CorrelationID         string    `json:"correlationId"`
	Count                 int       `json:"count"`
	Concurrency           int       `json:"concurrency"`
	Warm                  bool      `json:"warm"`
	LastAccessed          time.Time `json:"lastAccessed"`
	LastAccessedInSeconds float64   `json:"lastAccessedInSeconds"`
}

// Event defines Lambda warmer event
type Event struct {
	Warmer            bool   `json:"warmer"`
	Concurrency       int    `json:"concurrency"`
	WarmerInvocation  int    `json:"warmerinvocation"`
	WarmerConcurrency int    `json:"warmerConcurrency"`
	CorrelationID     string `json:"correlationId"`
}

// Config  defines Lambda warmer configurations
type Config struct {
	CorrelationID string
}

// Handler handles AWS and returns if event is a warmer event
func Handler(ctx context.Context, event map[string]interface{}, cfg ...Config) bool {
	var (
		payload       Event
		config        Config
		b, _          = json.Marshal(event)
		_             = json.Unmarshal(b, &payload)
		concurrency   = payload.Concurrency
		invokeCount   = payload.WarmerInvocation
		invokeTotal   = payload.WarmerConcurrency
		delay         = defaultDelayInMilliSeconds
		correlationID string
	)

	if cfg != nil {
		config = cfg[0]
		correlationID = config.CorrelationID
	}

	if !payload.Warmer {
		Warm = true
		LastAccess = time.Now()
		return false
	}

	if concurrency < 1 {
		concurrency = defaultConcurrency
	}

	if invokeCount < 1 {
		invokeCount = defaultInvocation
	}

	if invokeTotal < 1 {
		invokeTotal = concurrency
	}

	logMessage := Log{
		Action:        "warmer",
		Function:      funcName + ":" + funcVersion,
		CorrelationID: correlationID,
		Count:         invokeCount,
		Concurrency:   invokeTotal,
		Warm:          Warm,
		LastAccessed:  LastAccess,
	}
	if !LastAccess.IsZero() {
		logMessage.LastAccessedInSeconds = time.Now().Sub(LastAccess).Seconds()
	}
	b, _ = json.Marshal(logMessage)
	log.Println(string(b))

	Warm = true
	LastAccess = time.Now()

	if concurrency > 1 {
		var (
			waitgroup    sync.WaitGroup
			lambdaClient = lambda.New(session.New(), &aws.Config{MaxRetries: aws.Int(1)})
		)

		for i := 2; i <= concurrency; i++ {
			waitgroup.Add(1)
			go invokeLambda(&waitgroup, lambdaClient, i, concurrency, correlationID)
		}

		waitgroup.Wait()
	} else if invokeCount > 1 {
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}

	return true
}

func invokeLambda(waitgroup *sync.WaitGroup, lambdaClient *lambda.Lambda, invocation, concurrency int, correlationID string) {
	b, _ := json.Marshal(Event{
		Warmer:            true,
		WarmerInvocation:  invocation,
		WarmerConcurrency: concurrency,
		CorrelationID:     correlationID,
	})
	if _, err := lambdaClient.Invoke(&lambda.InvokeInput{
		FunctionName:   aws.String(funcName + ":" + funcVersion),
		InvocationType: aws.String("RequestResponse"),
		LogType:        aws.String("None"),
		Payload:        b,
	}); err != nil {
		log.Println(err)
	}

	waitgroup.Done()
}
