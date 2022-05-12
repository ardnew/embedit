package fifo

import (
	"errors"
	"strconv"
	"strings"
	//"runtime/volatile"
)

// R defines a type with the same interface as one of the RegisterN types from
// package "runtime/volatile" of TinyGo.
// It may be substituted in place of the RegisterN type for testing FIFO logic
// on a regular PC with standard Go, which does not provide "runtime/volatile".
type R uint32

// Set sets the value of the receiver r to the given uint32 v.
func (r *R) Set(v uint32) { *r = R(v) }

// Get returns the value of the receiver r as a uint32.
func (r *R) Get() uint32 { return uint32(*r) }

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

// Data defines the interface for elements of a FIFO, enabling any arbitrary,
// concrete type be used as FIFO elements' type.
type Data interface{}

// Buffer defines the interface for the user-defined type that stores elements
// of a FIFO.
// It provides type-generalization for the FIFO control type State.
type Buffer interface {
	Len() int
	Get(int) (Data, bool)
	Set(int, Data) bool
	String() string
}

// State contains the state and configuration of a statically-sized, circular
// FIFO (queue) data structure.
// The FIFO itself is represented by any user-defined type that implements the
// Buffer interface.
type State struct {
	buff Buffer
	capa R
	head R
	tail R
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

// Init initializes the receiver State's backing data store with the given
// Buffer buff, FIFO logical capacity capa, and OverflowMode mode.
// Refer to State.Reset for additional constraints and semantics.
// Returns the receiver for convenient initialization call chains.
func (s *State) Init(buff Buffer, capa int, mode OverflowMode) *State {
	s.mode = mode
	s.buff = buff
	s.Reset(capa)
	return s
}

// Reset discards all buffered data and sets the FIFO logical capacity.
// If capa is less than 0 or greater than buffer's physical length, uses the
// buffer's physical length.
//
//go:inline
func (s *State) Reset(capa int) {
	if s.buff != nil {
		if phy := s.buff.Len(); 0 > capa || capa > phy {
			capa = phy
		}
	}
	s.capa.Set(uint32(capa))
	s.head.Set(0)
	s.tail.Set(0)
}

// Cap returns the logical capacity of the receiver FIFO.
// The FIFO can hold at most Cap-1 elements.
//
//go:inline
func (s *State) Cap() int {
	return int(s.capa.Get())
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
	rem := (s.Cap() - 1) - s.Len()
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

// Enq enqueues the given element data at the back of the receiver FIFO and
// returns true.
// If the FIFO is full and no element can be enqueued, returns false.
//
// TODO(ardnew): Document both operations based on receiver's OverflowMode.
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

// Read implements the io.Reader interface. It dequeues min(s.Len(), len(data))
// elements from the receiver FIFO into the given slice data.
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

// Write implements the io.Writer interface. It enqueues min(s.Rem(), len(data))
// elements from the given slice data into the receiver FIFO.
// If len(data) equals 0, returns 0 and ErrWriteZero.
// Otherwise, if s.Rem() equals 0, returns 0 and ErrFull.
//
// TODO(ardnew): Document both operations based on receiver's OverflowMode.
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
			from = more - int(s.capa.Get()) + 1
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

// First returns the next element that would be dequeued from the receiver FIFO
// and true.
// If the FIFO is empty and no element would be dequeued, returns nil and false.
func (s *State) First() (Data, bool) { return s.Get(0) }

// Last returns the last element that would be dequeued from the receiver FIFO
// and true.
// If the FIFO is empty and no element would be dequeued, returns nil and false.
func (s *State) Last() (Data, bool) { return s.Get(-1) }

// index returns an index into the receiver FIFO based on sign/magnitude of i.
//
// If i is greater than or equal to zero and less then s.Len(), returns the
// index of the (i+1)'th element that would be dequeued from the receiver FIFO
// and true.
//
// Otherwise, if i is negative and -i is less than or equal to s.Len(), returns
// the index of the (s.Len()-i+1)'th element that would be dequeued from the
// receiver FIFO and true.
//
// Otherwise, returns 0 and false.
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

// Get returns the value of an element in the receiver FIFO, offset by i from the
// front of the queue if i is positive, or from the back of the queue if i is
// negative. For example:
//
//	Get(0)  == Get(-Len())  == First(), and
//	Get(-1) == Get(Len()-1) == Last().
//
// If the offset is beyond queue boundaries, returns 0 and false.
func (s *State) Get(i int) (Data, bool) {
	if n, ok := s.index(i); ok {
		return s.buff.Get(n)
	}
	return nil, false
}

// Set modifies the value of an element in the receiver FIFO.
// Set uses the same logic as Get to select an element in the FIFO.
func (s *State) Set(i int, data Data) bool {
	if n, ok := s.index(i); ok {
		return s.buff.Set(n, data)
	}
	return false
}

// Remove removes and returns the value of an element from the receiver FIFO,
// moving all trailing elements forward in queue. Reduces FIFO length by 1.
// Remove uses the same logic as Get to select an element in the FIFO.
func (s *State) Remove(i int) (Data, bool) {
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
// Insert uses the same logic as Get to select an insertion index in the FIFO.
// func (s *Fifo) Insert(i int, data FifoData) bool {

// 	head := s.head.Get()
// 	tail := s.tail.Get()
// 	if tail-head == s.capa.Get() {
// 		// queue is full, decide if we are removing the first or last element to
// 		// make room for the insertion
// 		switch s.mode {
// 		case FifoFullDiscardLast: // drop last

// 		case FifoFullDiscardFirst: // drop first

// 			for n := 0; n < i; n++ {
// 				if t, ok := s.Get(n + 1); ok {
// 					_ = s.Set(n, t)
// 				}
// 			}
// 			s.Set(i, data)
// 		}
// 	}

// 	// copy each element at index n to index n+1, for n>=i.
// 	if data, ok := s.Get(i); ok {
// 		for n := i; n < s.Len()-1; n++ {
// 			if t, ok := s.Get(n + 1); ok {
// 				_ = s.Set(n, t)
// 			}
// 		}
// 		return data, true
// 	}
// 	return nil, false
// }

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
