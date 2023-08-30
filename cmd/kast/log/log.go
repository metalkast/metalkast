package log

import (
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// SetLogger sets a concrete logging implementation for all deferred Loggers.
func SetLogger(l logr.Logger) {
	Log = l
}

// Log is the base logger used by kubebuilder.  It delegates
// to another logr.Logger.  You *must* call SetLogger to
// get any actual logging.
var Log = logr.New(log.NullLogSink{})
