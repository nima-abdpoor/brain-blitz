package richerror

type Op string
type Kind int

const (
	KindInvalid Kind = iota + 1
	KindForbidden
	KindNotFound
	KindUnexpected
)

type RichError struct {
	operation    Op
	kind         Kind
	wrappedError error
	message      string
	meta         map[string]interface{}
}

func New2(op Op) RichError {
	return RichError{operation: op}
}

func (re RichError) WithOperation(op Op) RichError {
	re.operation = op
	return re
}

func (re RichError) WithMessage(message string) RichError {
	re.message = message
	return re
}

func (re RichError) WithError(err error) RichError {
	re.wrappedError = err
	return re
}

func (re RichError) WithKind(kind Kind) RichError {
	re.kind = kind
	return re
}

func (re RichError) WithMeta(meta map[string]interface{}) RichError {
	re.meta = meta
	return re
}

func (re RichError) Error() string {
	return re.message
}

func (re RichError) Kind() Kind {
	if re.kind != 0 {
		return re.kind
	}

	err, ok := re.wrappedError.(RichError)
	if !ok {
		return 0
	}
	return err.Kind()
}

func (re RichError) Message() string {
	if re.message != "" {
		return re.message
	}

	err, ok := re.wrappedError.(RichError)
	if !ok {
		return re.wrappedError.Error()
	}
	return err.Message()
}
