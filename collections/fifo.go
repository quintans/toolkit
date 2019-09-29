package collections

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sync"

	tk "github.com/quintans/toolkit"
)

const (
	intByteSize = 4
)

var ErrShortRead = errors.New("short read")
var ErrNilData = errors.New("nil data")

// FileFifo stores stores all data to disk.
type FileFifo struct {
	dir     string
	fileCap int64

	headFileSize int64
	headIdx      int64 // head position
	headFile     *os.File
	headFileIdx  int64

	tailIdx     int64 // tail position
	tailFile    *os.File
	tailFileIdx int64

	peekedData []byte
}

// NewFileFifo creates a FIFO supported by files.
// The supporting files will have a max size. Whenever that size is exceeded, a new file will be created.
// When all elements of a file are consumed (Pop) that file will be deleted.
//
// FileFifo is not safe for concurrent access.
func NewFileFifo(dir string, fileCap int64) (*FileFifo, error) {
	this := new(FileFifo)
	this.dir = dir
	this.fileCap = 1024 * 1024 * fileCap // MB to b

	err := this.Clear()
	if err != nil {
		return nil, err
	}

	return this, nil
}

func (this *FileFifo) Clear() error {
	this.headFileSize = 0
	this.headIdx = 0
	this.headFileIdx = 0
	if this.headFile != nil {
		err := this.headFile.Close()
		if err != nil {
			return err
		}
		this.headFile = nil
	}

	this.tailIdx = 0
	this.tailFileIdx = 0
	if this.tailFile != nil {
		err := this.tailFile.Close()
		if err != nil {
			return err
		}
		this.tailFile = nil
	}

	// (re)create dir
	logger.Debugf("removing dir %s", this.dir)
	err := os.RemoveAll(this.dir)
	if err != nil {
		return err
	}

	logger.Debugf("creating dir %s", this.dir)
	err = os.MkdirAll(this.dir, 0777)
	if err != nil {
		return err
	}

	// open file
	err = this.nextHeadFile()
	if err != nil {
		return err
	}
	err = this.nextTailFile()
	if err != nil {
		return err
	}

	return nil
}

func (this *FileFifo) nextHeadFile() error {
	var err error
	if this.headFile != nil {
		err = this.headFile.Close()
		if err != nil {
			return err
		}
	}
	this.headFileIdx++
	fp := filepath.Join(this.dir, fmt.Sprintf("%016X", this.headFileIdx))
	logger.Debugf("creating file %s", fp)
	this.headFile, err = os.Create(fp)
	if err != nil {
		return err
	}
	this.headFileSize = 0
	return nil
}

func (this *FileFifo) nextTailFile() error {
	var err error
	if this.tailFile != nil {
		err = this.tailFile.Close()
		if err != nil {
			return err
		}
		logger.Debugf("removing file %s", this.tailFile.Name())
		err = os.Remove(this.tailFile.Name())
		if err != nil {
			return err
		}
	}
	this.tailFileIdx++
	fp := filepath.Join(this.dir, fmt.Sprintf("%016X", this.tailFileIdx))
	logger.Debugf("opening file %s", fp)
	this.tailFile, err = os.Open(fp)
	if err != nil {
		return err
	}
	return nil
}

func (this *FileFifo) Push(data []byte) error {
	if data == nil {
		return ErrNilData
	}

	if this.headFileSize > this.fileCap {
		this.nextHeadFile()
	}

	// write data size
	var buf32 = make([]byte, intByteSize)
	size := len(data)
	binary.BigEndian.PutUint32(buf32, uint32(size))
	n, err := this.headFile.Write(buf32)
	if err != nil {
		return err
	} else if n < intByteSize {
		return io.ErrShortWrite
	}

	// write data
	if size > 0 {
		n, err = this.headFile.Write(data)
		if err != nil {
			return err
		} else if n < size {
			return io.ErrShortWrite
		}
	}

	this.headFileSize += int64(intByteSize + size)
	this.headIdx++
	return nil
}

func (this *FileFifo) Pop() ([]byte, error) {
	data, err := this.Peek()
	this.peekedData = nil
	if data != nil {
		this.tailIdx++
	}
	return data, err
}

func (this *FileFifo) Peek() ([]byte, error) {
	if this.peekedData != nil {
		return this.peekedData, nil
	} else if this.Size() > 0 {
		// read data size
		buf := make([]byte, intByteSize)
		n, err := this.tailFile.Read(buf)
		if err == io.EOF {
			this.nextTailFile()
			return this.Peek()
		} else if err != nil {
			return nil, err
		} else if n < intByteSize {
			return nil, ErrShortRead
		}
		size := int(binary.BigEndian.Uint32(buf))

		// read data
		buf = make([]byte, size)
		if size > 0 {
			n, err = this.tailFile.Read(buf)
			if err != nil {
				return nil, err
			} else if n < size {
				return nil, ErrShortRead
			}
		}

		this.peekedData = buf
		return buf, nil

	} else {
		return nil, nil
	}
}

func (this *FileFifo) Size() int64 {
	return this.headIdx - this.tailIdx
}

type BigFifo struct {
	fileFifo *FileFifo
	sync.RWMutex

	buffer   chan interface{}
	header   chan interface{}
	quit     chan struct{}
	head     interface{}
	fileHead interface{}

	dir      string
	codec    tk.Codec
	dataType reflect.Type
}

// NewBigFifo creates a FIFO that after a certain number of elements will use disk files to store the elements.
//
// threshold: buffer size. number after which will store to disk.
// dir: directory where the files will be created.
// codec: codec to convert between []byte and interface{}
// zero: zero data type
func NewBigFifo(threshold int, dir string, fileCap int64, codec tk.Codec, zero interface{}) (*BigFifo, error) {
	// validate
	if threshold < 2 {
		return nil, errors.New("threshold must be greater than than 1")
	}
	if len(dir) == 0 {
		return nil, errors.New("dir is empty")
	}
	if codec == nil {
		return nil, errors.New("codec is nil")
	}
	if zero == nil {
		return nil, errors.New("zero is nil")
	}

	var err error
	this := new(BigFifo)
	this.fileFifo, err = NewFileFifo(dir, fileCap)
	if err != nil {
		return nil, err
	}
	this.quit = make(chan struct{})
	this.header = make(chan interface{})
	this.buffer = make(chan interface{}, threshold-1)
	this.codec = codec

	t := reflect.TypeOf(zero)
	// if pointer user non pointer type
	if t.Kind() == reflect.Ptr {
		this.dataType = t.Elem()
	} else {
		this.dataType = t
	}

	// a way to listen for the comsuption of the buffer
	go func() {
		for data := range this.buffer {
			this.Lock()
			this.head = data
			this.Unlock()

			select {
			case this.header <- data:
			case <-this.quit:
			}
			// wake up the reading from file to the buffer
			this.Lock()
		out:
			for this.fileHead != nil {
				select {
				case this.buffer <- this.fileHead:
					this.fileHead, _ = this.popFromFile()
				default:
					break out
				}
			}
			this.Unlock()
		}

	}()

	return this, nil
}

func (this *BigFifo) popFromFile() (interface{}, error) {
	// if there is data stored in file, get to memory
	data, err := this.fileFifo.Pop()
	if err != nil {
		return nil, err
	}

	if data != nil {
		// copy
		v := reflect.New(this.dataType)
		// decode
		err = this.codec.Decode(data, v.Interface())
		if err != nil {
			return nil, err
		}
		return v.Elem().Interface(), nil
	}

	return nil, nil
}

func (this *BigFifo) Size() int64 {
	this.RLock()
	defer this.RUnlock()

	var c int64
	if this.fileHead != nil {
		c = 1
	}

	return int64(len(this.buffer)) + c + this.fileFifo.Size()
}

func (this *BigFifo) Clear() error {
	this.Lock()
	defer this.Unlock()

	return this.fileFifo.Clear()
}

func (this *BigFifo) Close() error {
	var err = this.Close()
	close(this.buffer)
	close(this.quit)
	return err
}

func (this *BigFifo) Push(value interface{}) error {

	var err error
	select {
	case this.buffer <- value:
		// still has space int the buffer
		if this.head == nil {
			this.Lock()
			// double check
			if this.head == nil {
				this.head = value
			}
			this.Unlock()
		}

	default:
		this.Lock()
		// we only care for the first time that the file value is set
		if this.fileHead == nil {
			this.fileHead = value
		} else {
			// use disk, since the threshold was reached.
			data, err := this.codec.Encode(value)
			if err != nil {
				return err
			}
			err = this.fileFifo.Push(data)
		}

		this.Unlock()
	}

	return err
}

// Popper returns the channel element.
func (this *BigFifo) Popper() <-chan interface{} {
	return this.header
}

// Peek returns the next element. When we consume it it might be different
func (this *BigFifo) Peek() interface{} {
	this.RLock()
	defer this.RUnlock()
	return this.head
}

type item struct {
	next  *item
	value interface{}
}

// the Idea is to have a FIFO with a windowing (circular) feature.
// If the max size is reached, the oldest element will be removed.
type Fifo struct {
	mu   sync.RWMutex
	lock bool

	head     *item
	tail     *item
	size     int
	capacity int
}

func NewFifo(capacity int) *Fifo {
	return newLockFifo(capacity, false)
}

// NewLockFifo creates a Queue to be accessed concurrently
func NewLockFifo(capacity int) *Fifo {
	return newLockFifo(capacity, true)
}

func newLockFifo(capacity int, lock bool) *Fifo {
	this := new(Fifo)
	this.capacity = capacity
	this.lock = lock
	return this
}

func (this *Fifo) Size() int {
	if this.lock {
		this.mu.RLock()
		defer this.mu.RUnlock()
	}

	return this.size
}

// Clear resets the queue.
func (this *Fifo) Clear() {
	if this.lock {
		this.mu.Lock()
		defer this.mu.Unlock()
	}

	this.head = nil
	this.tail = nil
	this.size = 0
}

// Push adds an element to the head of the fifo.
// If the capacity was exceeded returns the element that had to be pushed out, otherwise returns nil.
func (this *Fifo) Push(value interface{}) interface{} {
	if this.lock {
		this.mu.Lock()
		defer this.mu.Unlock()
	}

	var old interface{}
	// if capacity == 0 it will add until memory is exausted
	if this.capacity > 0 && this.size == this.capacity {
		old = this.pop()
	}
	// adds new element
	e := &item{value: value}
	if this.head != nil {
		this.head.next = e
	}
	this.head = e

	if this.tail == nil {
		this.tail = e
	}

	this.size++

	return old
}

func (this *Fifo) pop() interface{} {
	var value interface{}
	if this.tail != nil {
		value = this.tail.value
		this.tail = this.tail.next
		this.size--
	}
	return value
}

// Pop returns the tail element removing it.
func (this *Fifo) Pop() interface{} {
	if this.lock {
		this.mu.Lock()
		defer this.mu.Unlock()
	}

	return this.pop()
}

// Peek returns the tail element without removing it.
func (this *Fifo) Peek() interface{} {
	if this.lock {
		this.mu.RLock()
		defer this.mu.RUnlock()
	}

	if this.tail != nil {
		return this.tail.value
	}
	return nil
}
