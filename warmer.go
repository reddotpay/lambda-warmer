package warmer

import (
	"context"
	"encoding/json"
	"log"
	"os"
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
	defaultDelayInMilliSeconds = 75
)

// Log defines log message
type Log struct {
	Action        string
	Function      string
	CorrelationID string
	Count         int
	Concurrency   int
	Warm          bool
	LastAccessed  time.Time
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
}

// Handler handles AWS
func Handler(ctx context.Context, event map[string]interface{}, cfg ...Config) error {
	var (
		payload Event
		b, _    = json.Marshal(event)
		_       = json.Unmarshal(b, &payload)
	)

	Warm = true
	LastAccess = time.Now()

	if !payload.Warmer {
		return New(ErrCodeNotWarmerEvent)
	}

	var (
		concurrency   = payload.Concurrency
		invokeCount   = payload.WarmerInvocation
		invokeTotal   = concurrency
		correlationID = ""
		delay         = defaultDelayInMilliSeconds
	)

	if concurrency < 1 {
		concurrency = defaultConcurrency
	}

	if invokeCount < 1 {
		invokeCount = defaultInvocation
	}

	log.Println(Log{
		Action:        "warmer",
		Function:      funcName + ":" + funcVersion,
		CorrelationID: correlationID,
		Count:         invokeCount,
		Concurrency:   invokeTotal,
		Warm:          Warm,
		LastAccessed:  LastAccess,
	})

	if concurrency > 1 {
		lambdaClient := lambda.New(session.New())

		for i := 2; i <= concurrency; i++ {
			invocationType := "Event"
			if i == concurrency {
				invocationType = "RequestResponse"
			}

			b, _ := json.Marshal(Event{
				Warmer:            true,
				WarmerInvocation:  i,
				WarmerConcurrency: concurrency,
				CorrelationID:     correlationID,
			})
			lambdaClient.InvokeWithContext(ctx, &lambda.InvokeInput{
				FunctionName:   aws.String(funcName + ":" + funcVersion),
				InvocationType: aws.String(invocationType),
				LogType:        aws.String("None"),
				Payload:        b,
			})
		}
	} else if invokeCount > 1 {
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}

	return nil
}
