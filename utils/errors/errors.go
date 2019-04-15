package errors

type ErrorReason string

const (
	ErrorReasonUbiquityServiceIPEmpty    ErrorReason = "Failed getting ubiquity serviceIP, it is empty"
	ErrorReasonENVNotSet                 ErrorReason = "ENV not set"
	ErrorReasonENVNamespaceNotSet        ErrorReason = "ENV NAMESPACE is not set"
	ErrorReasonENVStorageClassNotSet     ErrorReason = "ENV STORAGE_CLASS is not set"
	ErrorReasonENVUbiquityDbPvNameNotSet ErrorReason = "ENV UBIQUITY_DB_PV_NAME is not set"
	ErrorReasonENVUbiquityDbSCNotSet     ErrorReason = "ENV UBIQUITY_DB_STORAGECLASS is not set"
	ErrorReasonUnknown                   ErrorReason = "unknown"
)

type ubError struct {
	reason ErrorReason
}

func NewError(reason ErrorReason) *ubError {
	return &ubError{reason: reason}
}

type ReasonInterface interface {
	Reason() ErrorReason
}

var _ error = &ubError{}

// ubError implements the Error interface.
func (e *ubError) Error() string {
	return string(e.reason)
}

func (e *ubError) Reason() ErrorReason {
	return e.reason
}

// ReasonForError returns the reason for a particular error.
func ReasonForError(err error) ErrorReason {
	switch t := err.(type) {
	case ReasonInterface:
		return t.Reason()
	}
	return ErrorReasonUnknown
}

func IsUbiquityServiceIPEmpty(err error) bool {
	return ReasonForError(err) == ErrorReasonUbiquityServiceIPEmpty
}

func IsENVNamespaceNotSet(err error) bool {
	return ReasonForError(err) == ErrorReasonENVNamespaceNotSet
}

func IsENVStorageClassNotSet(err error) bool {
	return ReasonForError(err) == ErrorReasonENVStorageClassNotSet
}

func IsENVUbiquityDbPvNameNotSet(err error) bool {
	return ReasonForError(err) == ErrorReasonENVUbiquityDbPvNameNotSet
}

func IsENVUbiquityDbSCNotSet(err error) bool {
	return ReasonForError(err) == ErrorReasonENVUbiquityDbSCNotSet
}

var UbiquityServiceIPEmpty = NewError(ErrorReasonUbiquityServiceIPEmpty)
var ENVNamespaceNotSet = NewError(ErrorReasonENVNamespaceNotSet)
var ENVStorageClassNotSet = NewError(ErrorReasonENVStorageClassNotSet)
var ENVUbiquityDbPvNameNotSet = NewError(ErrorReasonENVUbiquityDbPvNameNotSet)
var ENVUbiquityDbSCNotSet = NewError(ErrorReasonENVUbiquityDbSCNotSet)
