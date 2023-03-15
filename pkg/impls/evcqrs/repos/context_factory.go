package repos

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"techunicorn.com/udc-core/gettingStarted/pkg/domain/base/cntxt"
	domcntxt "techunicorn.com/udc-core/gettingStarted/pkg/domain/base/cntxt"
	"techunicorn.com/udc-core/gettingStarted/pkg/domain/base/logger"
	implcntxt "techunicorn.com/udc-core/gettingStarted/pkg/impls/evcqrs/cntxt"
	"techunicorn.com/udc-core/gettingStarted/pkg/impls/evcqrs/common"

	"github.com/BetaLixT/go-resiliency/retrier"
	"go.uber.org/zap"
)

// =============================================================================
// Context factory and trace parsing logic
// =============================================================================

var _ cntxt.IFactory = (*ContextFactory)(nil)

// ContextFactory to create new contexts
type ContextFactory struct {
	lgrf logger.IFactory
}

// NewContextFactory constructor for context factory
func NewContextFactory(
	lgrf logger.IFactory,
) *ContextFactory {
	return &ContextFactory{
		lgrf: lgrf,
	}
}

// Create creates a new context with timeout, transactions and trace info
func (f *ContextFactory) Create(
	traceparent string,
) domcntxt.IContext {
	ver, tid, pid, rid, flg, err := parseTraceParent(traceparent)
	if err != nil {
		lgr := f.lgrf.Create(context.Background())
		lgr.Error("failed to generate trace info", zap.Error(err))
	}
	c := &internalContext{
		lgrf: f.lgrf,

		cancelmtx: &sync.Mutex{},
		err:       nil,
		done:      make(chan struct{}, 1),

		rtr: *retrier.New(retrier.ExponentialBackoff(
			5,
			500*time.Millisecond,
		),
			retrier.DefaultClassifier{},
		),
		compensatoryActions: []implcntxt.Action{},
		commitActions:       []implcntxt.Action{},
		events:              []dispatchableEvent{},
		txObjs:              map[string]interface{}{},
		isCommited:          false,
		isRolledback:        false,
		txmtx:               &sync.Mutex{},
		ver:                 ver,
		tid:                 tid,
		pid:                 pid,
		rid:                 rid,
		flg:                 flg,
	}

	// TODO: tracing values

	return c
}

// parseTraceParent parses and or generates trace information
func parseTraceParent(
	traceprnt string,
) (ver, tid, pid, rid, flg string, err error) {
	ver, tid, pid, flg, err = decodeTraceparent(traceprnt)
	// If the header could not be decoded, generate a new header
	if err != nil {
		ver, flg = "00", "01"
		if tid, err = generateRadomHexString(16); err != nil {
			return "", "", "", "", "", common.NewHexStringGenerationFailedError(err)
		}
	}

	// Generate a new resource id
	rid, err = generateRadomHexString(8)
	if err != nil {
		return "", "", "", "", "", common.NewHexStringGenerationFailedError(err)
	}
	return
}

func generateRadomHexString(n int) (string, error) {
	buff := make([]byte, n)
	if _, err := rand.Read(buff); err != nil {
		return "", err
	}
	return hex.EncodeToString(buff), nil
}

func decodeTraceparent(traceparent string) (string, string, string, string, error) {
	// Fast fail for common case of empty string
	if traceparent == "" {
		return "", "", "", "", fmt.Errorf("traceparent is empty string")
	}

	hexfmt, err := regexp.Compile("^[0-9A-Fa-f]*$")
	vals := strings.Split(traceparent, "-")

	if len(vals) == 4 {
		ver, tid, pid, flg := vals[0], vals[1], vals[2], vals[3]
		if !hexfmt.MatchString(ver) || len(ver) != 2 {
			err = fmt.Errorf("invalid traceparent version")
		} else if !hexfmt.MatchString(pid) || len(pid) != 16 {
			err = fmt.Errorf("invalid traceparent parent id")
		} else if !hexfmt.MatchString(flg) || len(flg) != 2 {
			err = fmt.Errorf("invalid traceparent flag")
		} else if !hexfmt.MatchString(tid) || len(tid) != 32 {
			err = fmt.Errorf("invalid traceparent trace id")
		} else if tid == "00000000000000000000000000000000" {
			err = fmt.Errorf("traceparent trace id value is zero")
		} else {
			return ver, tid, pid, flg, nil
		}
	} else {
		err = fmt.Errorf("invalid traceparent trace id")
	}

	return "", "", "", "", err
}
