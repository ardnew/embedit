package fifo

import (
	"errors"
	"strconv"
	"strings"
	//"runtime/volatile"
)

// R defines a type with the same interface as one of the RegisterN types from
// package "runtime/volatile" of TinyGo.
// It may be substituted in place of the RegisterN type for testing Fifo logic
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
		return "discard-last"
	case DiscardFirst:
		return "discard-first"
	}
	return ""
}

// Data defines the interface for elements of a FIFO, enabling any arbitrary,
// concrete type be used as a FIFO's elements.
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

// State contains the state and configuration of a generalized queue data
// structure. The queue itself is represented by any user-defined type that
// implements the Buffer interface.
type State struct {
	mode OverflowMode
	capa R // volatile.Register32
	buff Buffer
	head R // volatile.Register32 // Oldest element in queue is at index head
	tail R // volatile.Register32 // New elements are enqueued at index tail
}

// Error values returned by methods of type State.
var (
	ErrFifoReadZero    = errors.New("copy into zero-length buffer")
	ErrFifoWriteZero   = errors.New("copy from zero-length buffer")
	ErrFifoEmpty       = errors.New("buffer empty") // Read underrun
	ErrFifoFull        = errors.New("buffer full")  // Write overrun
	ErrFifoDiscardMode = errors.New("unknown discard mode")
)

// New allocates and initializes a new State with the given paramters and
// returns a pointer to this fully-initialized State.
func New(buff Buffer, capa int, mode OverflowMode) *State {
	fifo := State{}
	fifo.Init(buff, capa, mode)
	return &fifo
}

// Init initializes the receiver queue's backing data store with the given
// buffer buff, logical capacity capa, and discard rule mode.
// If capa is greater than the buffer's physical length, uses the buffer's
// physical length.
func (q *State) Init(buff Buffer, capa int, mode OverflowMode) {
	q.mode = mode
	q.buff = buff
	q.Reset(capa)
}

// Reset discards all buffered data and sets the FIFO logical capacity.
// If capa is less than 0 or greater than buffer's physical length, uses the
// buffer's physical length.
//go:inline
func (q *State) Reset(capa int) {
	if q.buff != nil {
		if phy := q.buff.Len(); 0 > capa || capa > phy {
			capa = phy
		}
	}
	q.capa.Set(uint32(capa))
	q.head.Set(0)
	q.tail.Set(0)
}

// Cap returns the logical capacity of the receiver FIFO.
//go:inline
func (q *State) Cap() int {
	return int(q.capa.Get())
}

// Len returns the number of elements enqueued in the receiver FIFO.
//go:inline
func (q *State) Len() int {
	return int(q.tail.Get() - q.head.Get())
}

// Rem returns the number of elements not enqueued in the receiver FIFO.
//go:inline
func (q *State) Rem() int {
	return q.Cap() - q.Len()
}

// Deq dequeues and returns the element at the front of the receiver FIFO and
// true.
// If the FIFO is empty and no element was dequeued, returns nil and false.
func (q *State) Deq() (data Data, ok bool) {
	if q == nil {
		return // invalid receiver
	}
	head := q.head.Get()
	if head == q.tail.Get() {
		return // empty queue
	}
	if q.buff == nil {
		return // uninitialized buffer
	}
	if data, ok = q.buff.Get(int(head % q.capa.Get())); ok {
		q.head.Set(head + 1)
	}
	return
}

// Enq enqueues the given element data at the back of the receiver FIFO and
// returns true.
// If the FIFO is full and no element can be enqueued, returns false.
//
// TODO(ardnew): Document both operations based on receiver's FifoFullMode.
func (q *State) Enq(data Data) (ok bool) {
	if q == nil {
		return // invalid receiver
	}
	tail := q.tail.Get()
	head := q.head.Get()
	if q.Len() == int(q.capa.Get())-1 {
		// full queue:
		switch q.mode {
		case DiscardLast: // drop incoming data
			return
		case DiscardFirst: // drop outgoing data
			q.head.Set(head + 1)
		}
	}
	if q.buff == nil {
		return // uninitialized buffer
	}
	if ok = q.buff.Set(int(tail%q.capa.Get()), data); ok {
		q.tail.Set(tail + 1)
	}
	return
}

// Read implements the io.Reader interface. It dequeues min(q.Len(), len(data))
// elements from the receiver FIFO into the given slice data.
// If len(data) equals 0, returns 0 and ErrReadBuffer.
// Otherwise, if q.Len() equals 0, returns 0 and ErrFifoEmpty.
func (q *State) Read(data []Data) (int, error) {
	less := uint32(len(data))
	if less == 0 {
		return 0, ErrFifoReadZero
	} // nothing to copy into

	head := q.head.Get()
	used := q.tail.Get() - head

	if used == 0 {
		return 0, ErrFifoEmpty
	} // empty queue

	if less > used {
		less = used
	} // only get from used space

	for i := uint32(0); i < less; i++ {
		data[i], _ = q.buff.Get(int(head % q.capa.Get()))
		head++
	}
	q.head.Set(head)

	return int(less), nil
}

// Write implements the io.Writer interface. It enqueues min(q.Rem(), len(data))
// elements from the given slice data into the receiver FIFO.
// If len(data) equals 0, returns 0 and ErrWriteBuffer.
// Otherwise, if q.Rem() equals 0, returns 0 and ErrFifoFull.
//
// TODO(ardnew): Document both operations based on receiver's FifoFullMode.
func (q *State) Write(data []Data) (int, error) {
	more := uint32(len(data))

	// Nothing to copy from is an error regardless of mode.
	if more == 0 {
		return 0, ErrFifoWriteZero
	}

	switch q.mode {
	case DiscardLast:
		// drop incoming data

		tail := q.tail.Get()
		used := tail - q.head.Get()

		// Full queue, cannot add any data.
		if used == q.capa.Get() {
			return 0, ErrFifoFull
		}

		// Only put to unused space.
		if used+more > q.capa.Get() {
			more = q.capa.Get() - used
		}

		// Copy a potentially-limited number of elements from data, depending on the
		// current length of FIFO.
		for i := uint32(0); i < more; i++ {
			if q.buff.Set(int(tail%q.capa.Get()), data[i]) {
				tail++
			}
		}
		q.tail.Set(tail)

		return int(more), nil

	case DiscardFirst:
		// drop outgoing data

		// Trying to write more data than the FIFO will hold will simply overwrite
		// some of the given data, so there is no point writing that data.
		from := uint32(0)
		if more >= q.capa.Get() {
			// Begin copying only the data that will be kept.
			from = more - q.capa.Get()
			// We can fill the entire FIFO.
			more = q.capa.Get()
			// Reset the indices
			q.head.Set(0)
			q.tail.Set(0)
		}

		tail := q.tail.Get()
		used := tail - q.head.Get()

		// Make space for incoming data by discarding only as many FIFO elements as
		// is necessary to store incoming data.
		if used+more > q.capa.Get() {
			q.head.Set(tail + more - q.capa.Get())
		}

		// Copy a potentially-limited number of elements from data, depending on the
		// current length of FIFO.
		for i := uint32(0); i < more; i++ {
			if q.buff.Set(int(tail%q.capa.Get()), data[from+i]) {
				tail++
			}
		}
		q.tail.Set(tail)

		return int(more), nil
	}

	return 0, ErrFifoDiscardMode
}

// First returns the next element that would be dequeued from the receiver FIFO
// and true.
// If the FIFO is empty and no element would be dequeued, returns nil and false.
func (q *State) First() (Data, bool) { return q.Get(0) }

// Last returns the last element that would be dequeued from the receiver FIFO
// and true.
// If the FIFO is empty and no element would be dequeued, returns nil and false.
func (q *State) Last() (Data, bool) { return q.Get(-1) }

// index returns an index into the receiver FIFO based on sign and magnitude of i:
// 1. If i is greater than or equal to zero and less then q.Len(), returns the
//    index of the (i+1)'th element that would be dequeued from the receiver FIFO
//    and true.
// 2. Otherwise, if i is negative and -i is less than or equal to q.Len(), returns
//    the index of the (q.Len()-i+1)'th element that would be dequeued from the
//    receiver FIFO and true.
// 3. Otherwise, returns 0 and false.
func (q *State) index(i int) (int, bool) {
	if n := q.Len(); i < 0 {
		if -i <= n {
			return (int(q.tail.Get()) + i) % int(q.capa.Get()), true
		}
	} else {
		if i < n {
			return (int(q.head.Get()) + i) % int(q.capa.Get()), true
		}
	}
	return 0, false
}

// Get returns the value of an element in the receiver FIFO, offset by i from the
// front of the queue if i is positive, or from the back of the queue if i is
// negative. For example:
//  	Get(0)  == Get(-Len())  == First(), and
// 		Get(-1) == Get(Len()-1) == Last().
// If the offset is beyond queue boundaries, returns 0 and false.
func (q *State) Get(i int) (Data, bool) {
	if n, ok := q.index(i); ok {
		return q.buff.Get(n)
	}
	return nil, false
}

// Set modifies the value of an element in the receiver FIFO.
// Set uses the same logic as Get to select an element in the FIFO.
func (q *State) Set(i int, data Data) bool {
	if n, ok := q.index(i); ok {
		return q.buff.Set(n, data)
	}
	return false
}

// Remove removes and returns the value of an element from the receiver FIFO,
// moving all trailing elements forward in queue. Reduces FIFO length by 1.
// Remove uses the same logic as Get to select an element in the FIFO.
func (q *State) Remove(i int) (Data, bool) {
	head := q.head.Get()
	tail := q.tail.Get()
	if head == tail {
		return nil, false
	} // empty queue

	// copy each element at index n+1 to index n, for n>=i.
	if data, ok := q.Get(i); ok {
		for n := i; n < int(tail-head)-1; n++ {
			if t, ok := q.Get(n + 1); ok {
				_ = q.Set(n, t)
			}
		}
		q.tail.Set(tail - 1)
		return data, true
	}
	return nil, false
}

// Insert increases FIFO length by 1, moving all elements trailing the insertion
// index backward in queue, and copies the given data into the queue at that
// index.
// Insert uses the same logic as Get to select an insertion index in the FIFO.
// func (q *Fifo) Insert(i int, data FifoData) bool {

// 	head := q.head.Get()
// 	tail := q.tail.Get()
// 	if tail-head == q.capa.Get() {
// 		// queue is full, decide if we are removing the first or last element to
// 		// make room for the insertion
// 		switch q.mode {
// 		case FifoFullDiscardLast: // drop last

// 		case FifoFullDiscardFirst: // drop first

// 			for n := 0; n < i; n++ {
// 				if t, ok := q.Get(n + 1); ok {
// 					_ = q.Set(n, t)
// 				}
// 			}
// 			q.Set(i, data)
// 		}
// 	}

// 	// copy each element at index n to index n+1, for n>=i.
// 	if data, ok := q.Get(i); ok {
// 		for n := i; n < q.Len()-1; n++ {
// 			if t, ok := q.Get(n + 1); ok {
// 				_ = q.Set(n, t)
// 			}
// 		}
// 		return data, true
// 	}
// 	return nil, false
// }

func (q *State) String() string {
	if q == nil {
		return "<nil>"
	}
	var sb strings.Builder
	sb.WriteRune('{')
	sb.WriteString("mode:" + q.mode.String() + ", ")
	sb.WriteString("capa:" + strconv.FormatUint(uint64(q.capa.Get()), 10) + ", ")
	sb.WriteString("head:" + strconv.FormatUint(uint64(q.head.Get()), 10) + "[" + strconv.FormatUint(uint64(q.head.Get()), 10) + "], ")
	sb.WriteString("tail:" + strconv.FormatUint(uint64(q.tail.Get()), 10) + "[" + strconv.FormatUint(uint64(q.tail.Get()), 10) + "], ")
	sb.WriteString("size:" + strconv.FormatUint(uint64(q.Len()), 10) + "], ")
	// if q.buff != nil {
	sb.WriteString("buff:" + q.buff.String())
	// } else {
	// sb.WriteString("buff:<nil>")
	// }
	sb.WriteRune('}')
	return sb.String()
}
