package main

// DataLogIterator allows to it iterate over DataLogMem lines
type DataLogIterator struct {
	dlm     *DataLogMem
	current int
}

// DataLogIteratorNew creates an iterator (see DataLogIterator above)
// A call to Next() or Prev() must be done before reading the first Value()
// (iterator will start form start or end accordingly)
func DataLogIteratorNew(dlm *DataLogMem) *DataLogIterator {
	var iter DataLogIterator
	iter.dlm = dlm
	iter.current = -1
	return &iter
}

// Next iterates forward (false if not possible)
func (iter *DataLogIterator) Next() bool {
	if iter.current+1 >= len(iter.dlm.Lines) {
		return false
	}
	iter.current++
	return true
}

// Prev iterates backward (false if not possible)
func (iter *DataLogIterator) Prev() bool {
	// first call is a Prev: let's start from the end
	if iter.current == -1 {
		iter.current = len(iter.dlm.Lines)
	}

	if iter.current-1 < 0 {
		return false
	}
	iter.current--
	return true
}

// Value return the current value of the iterator (if possible)
func (iter *DataLogIterator) Value() *DataLogLine {
	if iter.current >= 0 && iter.current < len(iter.dlm.Lines) {
		return &iter.dlm.Lines[iter.current]
	}
	return nil
}
