// Code generated by mkerrors.pl. DO NOT EDIT.

package errors

//go:generate perl mkerrors.pl
// InvalidReceiver
// InvalidArgument
// OutOfRange
// WriteOverflow
// ReadOverflow
// PasteIndicator

type (
	InvalidReceiver struct{}
	InvalidArgument struct{}
	OutOfRange      struct{}
	WriteOverflow   struct{}
	ReadOverflow    struct{}
	PasteIndicator  struct{}
)

var (
	ErrInvalidReceiver InvalidReceiver
	ErrInvalidArgument InvalidArgument
	ErrOutOfRange      OutOfRange
	ErrWriteOverflow   WriteOverflow
	ErrReadOverflow    ReadOverflow
	ErrPasteIndicator  PasteIndicator
)

func (e *InvalidReceiver) Error() string {
	return "invalid receiver"
}

func (e *InvalidArgument) Error() string {
	return "invalid argument"
}

func (e *OutOfRange) Error() string {
	return "out of range"
}

func (e *WriteOverflow) Error() string {
	return "write overflow"
}

func (e *ReadOverflow) Error() string {
	return "read overflow"
}

func (e *PasteIndicator) Error() string {
	return "paste indicator"
}
