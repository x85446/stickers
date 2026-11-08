package table

// ErrorBadType reports a value whose type is not one a column can hold.
type ErrorBadType struct {
	msg string
}

// Error returns the message.
func (e ErrorBadType) Error() string {
	return e.msg
}

// ErrorRowLen reports a row whose length differs from the number of columns.
type ErrorRowLen struct {
	msg string
}

// Error returns the message.
func (e ErrorRowLen) Error() string {
	return e.msg
}

// ErrorBadCellType reports a value whose type differs from its column's type.
type ErrorBadCellType struct {
	msg string
}

// Error returns the message.
func (e ErrorBadCellType) Error() string {
	return e.msg
}
