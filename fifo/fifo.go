package fifo

import (
	"errors"
	"strconv"
	"strings"

	"github.com/ardnew/embedit/volatile"
)

// OverflowMode enumerates the options for handling data enqueued to a
// FIFO that is filled to capacity.
type OverflowMode uint8

// Constant values for enumerated type OverflowMode.
const (
	DiscardLast  OverflowMode = iota // Drop incoming data
	DiscardFirst                     // Drop outgoing data
)

// String returns a string description of the receiver OverflowMode m.
func (m OverflowMode) String() string {
	switch m {
	case DiscardLast:
		return "OverflowMode(DiscardLast)"
	case DiscardFirst:
		return "OverflowMode(DiscardFirst)"
	}
	s := strings.ToUpper(strconv.FormatUint(uint64(m), 16))
	return "<OverflowMode(0x" + s + "):invalid>"
}

// Data defines the interface for elements of a FIFO.
type Data interface{}

// Buffer defines the interface for a list of Data elements.
// Users must provide a Buffer when creating a State struct, which provides the
// methods for managing FIFO operations.
type Buffer interface {
	// Len returns the receiver's size.
	Len() int
	// Get returns a Data element at the given index and true.
	// Returns nil and false if the given index is outside receiver's bounds.
	Get(int) (Data, bool)
	// Set sets the Data element at the given index and returns true.
	// Returns false if the given index is outside receiver's bounds.
	Set(int, Data) bool
	// String returns a descriptive string of the receiver.
	String() string
}

// State contains the state and configuration of a statically-sized, circular
// FIFO (queue) data structure.
// The FIFO itself is represented by any user-defined type that implements the
// Buffer interface.
type State struct {
	buff Buffer
	capa volatile.Register32
	head volatile.Register32
	tail volatile.Register32
	mode OverflowMode
}

// Error values returned by methods of type State.
var (
	ErrReceiver    = errors.New("nil receiver")
	ErrReadZero    = errors.New("copy into zero-length buffer")
	ErrWriteZero   = errors.New("copy from zero-length buffer")
	ErrNil         = errors.New("nil buffer")
	ErrEmpty       = errors.New("buffer empty") // Read underrun
	ErrFull        = errors.New("buffer full")  // Write overrun
	ErrDiscardMode = errors.New("unknown discard mode")
)

// New allocates and initializes a new State with the given paramters and
// returns a pointer to the fully-initialized struct.
func New(buff Buffer, capa int, mode OverflowMode) *State {
	return new(State).Init(buff, capa, mode)
}

// Init initializes the receiver's backing data store, logical capacity, and
// behavior on overflow.
// Refer to State.Reset for additional constraints and semantics.
// Returns the receiver for convenient initialization call chains.
func (s *State) Init(buff Buffer, capa int, mode OverflowMode) *State {
	s.mode = mode
	s.buff = buff
	s.Reset(capa)
	return s
}

// Reset discards all buffered data and sets the FIFO logical capacity.
// If capa is less than 0 or greater than Buffer's length, uses the Buffer's
// length.
//
//go:inline
func (s *State) Reset(capa int) {
	if s.buff != nil {
		if len := s.buff.Len(); capa < 0 || capa > len {
			capa = len
		}
	}
	s.capa.Set(uint32(capa))
	s.head.Set(0)
	s.tail.Set(0)
}

// Cap returns the logical capacity of the receiver.
// The FIFO can hold at most Cap elements.
//
//go:inline
func (s *State) Cap() int {
	if s.buff == nil {
		return 0
	}
	cap := int(s.capa.Get())
	if len := s.buff.Len(); cap > len {
		return len
	}
	return cap
}

// Len returns the number of elements enqueued in the receiver FIFO.
//
//go:inline
func (s *State) Len() int {
	return int(s.tail.Get() - s.head.Get())
}

// Rem returns the number of elements not enqueued in the receiver FIFO.
//
//go:inline
func (s *State) Rem() int {
	rem := s.Cap() - s.Len()
	if rem <= 0 {
		return 0
	}
	return rem
}

// Deq dequeues and returns the element at the front of the receiver FIFO and
// true.
// If the FIFO is empty and no element was dequeued, returns nil and false.
func (s *State) Deq() (data Data, ok bool) {
	if s == nil {
		return // Invalid receiver
	}
	if s.buff == nil {
		return // Uninitialized buffer
	}
	if s.Len() == 0 {
		return // Empty queue
	}
	if data, ok = s.buff.Get(int(s.head.Get() % s.capa.Get())); ok {
		s.head.Set(s.head.Get() + 1)
	}
	return
}

// Enq enqueues the given element at the back of the receiver FIFO and returns
// true.
// If the FIFO is full and no element can be enqueued, returns false.
func (s *State) Enq(data Data) (ok bool) {
	if s == nil {
		return // Invalid receiver
	}
	if s.buff == nil {
		return // Uninitialized buffer
	}
	if s.Rem() == 0 {
		// Full queue:
		switch s.mode {
		case DiscardLast: // Drop incoming data
			return
		case DiscardFirst: // Drop outgoing data
			s.head.Set(s.head.Get() + 1)
		}
	}
	if ok = s.buff.Set(int(s.tail.Get()%s.capa.Get()), data); ok {
		s.tail.Set(s.tail.Get() + 1)
	}
	return
}

// Read dequeues min(s.Len(), len(data)) elements from the receiver FIFO into
// the given Data slice and returns the number of elements read.
// If len(data) equals 0, returns 0 and ErrReadZero.
// Otherwise, if s.Len() equals 0, returns 0 and ErrEmpty.
func (s *State) Read(data []Data) (n int, err error) {
	if s == nil {
		return 0, ErrReceiver
	}
	if s.buff == nil {
		return 0, ErrNil
	}
	// FIFO empty, cannot read any data.
	if s.Len() == 0 {
		return 0, ErrEmpty
	}
	less := len(data)
	// Nowhere to write data, output buffer size is 0.
	if less == 0 {
		return 0, ErrReadZero
	}
	// Only get from used space.
	if less > s.Len() {
		less = s.Len()
	}
	head := s.head.Get()
	for i := 0; i < less; i++ {
		data[i], _ = s.buff.Get(int(head % s.capa.Get()))
		head++
	}
	s.head.Set(head)
	return less, nil
}

// Write enqueues min(s.Rem(), len(data)) elements from the given Data slice
// into the receiver FIFO.
// If len(data) equals 0, returns 0 and ErrWriteZero.
// Otherwise, if s.Rem() equals 0, returns 0 and ErrFull.
func (s *State) Write(data []Data) (int, error) {
	if s == nil {
		return 0, ErrReceiver
	}
	if s.buff == nil {
		return 0, ErrNil
	}
	more := len(data)
	// Nothing to copy from is an error regardless of mode.
	if more == 0 {
		return 0, ErrWriteZero
	}
	// Do not attempt writing if FIFO capacity is zero.
	if s.Cap() == 0 {
		return 0, ErrFull
	}
	switch s.mode {
	case DiscardLast: // Drop incoming data
		// FIFO full, cannot add any data.
		if s.Rem() == 0 {
			return 0, ErrFull
		}
		// Only put to unused space.
		if more > s.Rem() {
			more = s.Rem()
		}
		// Copy a potentially-limited number of elements from data, depending on the
		// current length of FIFO.
		tail := s.tail.Get()
		for i := 0; i < more; i++ {
			if s.buff.Set(int(tail%s.capa.Get()), data[i]) {
				tail++
			}
		}
		s.tail.Set(tail)
		return more, nil

	case DiscardFirst: // Drop outgoing data
		// Trying to write more data than the FIFO will hold will simply overwrite
		// some of the given data, so there is no point writing that data.
		span, from := more, 0
		if more >= int(s.capa.Get()) {
			// Reset the indices.
			s.head.Set(0)
			s.tail.Set(0)
			// Begin copying only the data that will be kept.
			from = more - int(s.capa.Get())
			// We can fill the entire FIFO.
			more = s.Rem()
		}
		// Make space for incoming data by discarding only as many FIFO elements as
		// is necessary to store incoming data.
		if more > s.Rem() {
			s.head.Set(s.head.Get() + uint32(more-s.Rem()))
		}
		// Copy a potentially limited number of elements from data, depending on the
		// current length of FIFO.
		tail := s.tail.Get()
		for i := 0; i < more; i++ {
			if s.buff.Set(int(tail%s.capa.Get()), data[from+i]) {
				tail++
			}
		}
		s.tail.Set(tail)
		return span, nil
	}
	return 0, ErrDiscardMode
}

// First returns the next element that would be dequeued from the receiver FIFO.
// If no element would be dequeued, returns nil.
func (s *State) First() Data { data, _ := s.Get(0); return data }

// Last returns the last element that would be dequeued from the receiver FIFO.
// If no element would be dequeued, returns nil.
func (s *State) Last() Data { data, _ := s.Get(-1); return data }

// index returns a Buffer index offset from the head of the queue if i is
// positive, or offset from the tail of the queue if i is negative, and true.
func (s *State) index(i int) (int, bool) {
	if n := s.Len(); i < 0 {
		if -i <= n {
			return (int(s.tail.Get()) + i) % int(s.capa.Get()), true
		}
	} else {
		if i < n {
			return (int(s.head.Get()) + i) % int(s.capa.Get()), true
		}
	}
	return 0, false
}

// Get returns a FIFO element at index offset from the head of the queue if i is
// positive, or offset from the tail of the queue if i is negative, and true.
// Returns nil and false if i is offset beyond queue boundaries.
//
// For example, Get(0)=Get(-Len())=First(), and Get(-1)=Get(Len()-1)=Last().
func (s *State) Get(i int) (Data, bool) {
	if s == nil || s.buff == nil {
		return nil, false
	}
	n, ok := s.index(i)
	if !ok {
		return nil, false
	}
	return s.buff.Get(n)
}

// Set modifies the value of an element in the receiver FIFO.
// Set uses the same logic as Get to select an index in the FIFO.
func (s *State) Set(i int, data Data) bool {
	if s == nil || s.buff == nil {
		return false
	}
	n, ok := s.index(i)
	if !ok {
		return false
	}
	return s.buff.Set(n, data)
}

// Remove removes and returns the value of an element from the receiver FIFO,
// moving all trailing elements forward in queue. Reduces FIFO length by 1.
// Remove uses the same logic as Get to select an index in the FIFO.
func (s *State) Remove(i int) (Data, bool) {
	if s == nil || s.buff == nil {
		return nil, false
	}
	head := s.head.Get()
	tail := s.tail.Get()
	if head == tail {
		return nil, false // Empty queue
	}
	// Copy each element at index n+1 to index n, for n>=i.
	if data, ok := s.Get(i); ok {
		for n := i; n < int(tail-head)-1; n++ {
			if t, ok := s.Get(n + 1); ok {
				_ = s.Set(n, t)
			}
		}
		s.tail.Set(tail - 1)
		return data, true
	}
	return nil, false
}

// Insert increases FIFO length by 1, moving all elements trailing the insertion
// index backward in queue, and copies the given data into the queue at that
// index.
// Insert uses the same logic as Get to select an index in the FIFO.
func (s *State) Insert(i int, data Data) bool {
	if s == nil || s.buff == nil || i >= s.Len() {
		return false
	}
	head := s.head.Get()
	tail := s.tail.Get()
	var move int
	if s.Rem() == 0 {
		// Queue is full, decide if we are removing the first or last element to
		// make room for the insertion
		switch s.mode {
		case DiscardLast: // Drop incoming data

		case DiscardFirst: // Drop outgoing data
			s.head.Set(head + 1)
			s.tail.Set(tail + 1)
		}
		move = 1
	} else {
		s.tail.Set(tail + 1)
		move = 0
	}
	// Copy each element at index n to index n+1, for n>=i.
	result := true
	for n := int(tail-head) - move; n > i; n-- {
		if t, ok := s.Get(n - 1); ok {
			result = result && s.Set(n, t)
		}
	}
	// Restore head/tail if setting the inserted Data fails.
	if !result || !s.Set(i, data) {
		s.head.Set(head)
		s.tail.Set(tail)
		return false
	}
	return true
}

func (s *State) String() string {
	if s == nil {
		return "<nil>"
	}
	var sb strings.Builder
	sb.WriteRune('{')
	sb.WriteString("mode:" + s.mode.String() + ", ")
	sb.WriteString("capa:" + strconv.FormatUint(uint64(s.capa.Get()), 10) + ", ")
	sb.WriteString("head:" + strconv.FormatUint(uint64(s.head.Get()), 10))
	if s.capa.Get() != 0 && s.head.Get() >= s.capa.Get() {
		sb.WriteString("[" + strconv.FormatUint(uint64(s.head.Get()%s.capa.Get()), 10) + "]")
	}
	sb.WriteString(", ")
	sb.WriteString("tail:" + strconv.FormatUint(uint64(s.tail.Get()), 10))
	if s.capa.Get() != 0 && s.tail.Get() >= s.capa.Get() {
		sb.WriteString("[" + strconv.FormatUint(uint64(s.tail.Get()%s.capa.Get()), 10) + "]")
	}
	sb.WriteString(", ")
	sb.WriteString("size:" + strconv.FormatUint(uint64(s.Len()), 10) + ", ")
	if s.buff == nil {
		sb.WriteString("buff:<nil>")
	} else {
		sb.WriteString("buff:" + s.buff.String())
	}
	sb.WriteRune('}')
	return sb.String()
}
