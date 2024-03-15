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

func New(op Op) RichError {
	return RichError{operation: op}
}

func (re RichError) withOperation(op Op) RichError {
	re.operation = op
	return re
}

func (re RichError) withMessage(message string) RichError {
	re.message = message
	return re
}

func (re RichError) withKind(kind Kind) RichError {
	re.kind = kind
	return re
}

func (re RichError) withMeta(meta map[string]interface{}) RichError {
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
