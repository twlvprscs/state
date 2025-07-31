package switchboard

import (
	"fmt"
	"math"
)

const (
	// capacity is the number of uint64 words in a register
	capacity = 64
	// wordSize is the number of bits in each word (uint64)
	wordSize = 64
	// maxReg is the maximum number of states that can be tracked (capacity * wordSize)
	maxReg = capacity * wordSize
)

// offset calculates the word index and bit offset within that word for a given state index.
// It returns the word index and the bit offset within that word.
// Panics if the index exceeds maxReg (4096).
func offset(idx uint) (uint, uint) {
	if idx > maxReg {
		panic(fmt.Sprintf("state: overflow - a single state S can hold no more than %d indices", maxReg))
	}

	return idx / wordSize, idx % wordSize
}

// shift returns a uint64 with a single bit set at the specified offset.
// This is used to create bit masks for manipulating individual bits in a register.
func shift(offs uint) uint64 {
	return uint64(1 << offs)
}

// register is an array of uint64 values used to efficiently store the state (open/closed)
// of up to 4096 different conditions using bit manipulation.
type register [capacity]uint64

// registerWithAllClosed creates a new register with all bits set to 1 (closed state).
// By default, registers are initialized with all bits set to 0 (open state).
func registerWithAllClosed() (r register) {
	for i := 0; i < capacity; i++ {
		r[i] = math.MaxUint64
	}
	return
}

// Commented out functions for potential future use
//
// // registerUnion returns a register that has bits set where both left and right registers have bits set.
// // This is equivalent to a logical AND operation on the registers.
// func registerUnion(left, right register) (out register) {
// 	for i := 0; i < capacity; i++ {
// 		out[i] = left[i] & right[i]
// 	}
// 	return
// }
//
// // registerDiff returns a register that has bits set where the left and right registers differ.
// // This is equivalent to a logical XOR operation on the registers.
// func registerDiff(left, right register) (out register) {
// 	for i := 0; i < capacity; i++ {
// 		out[i] = left[i] ^ right[i]
// 	}
// 	return
// }

// registerClose sets the specified indices to the closed state (bit value 1).
// It returns the modified register and a slice of indices that changed state.
// If an index was already closed, it won't be included in the returned slice.
func registerClose(r register, indices ...uint) (register, []uint) {
	out := make([]uint, 0, len(indices))
	for i := 0; i < len(indices); i++ {
		idx, offs := offset(indices[i])
		shifted := shift(offs)
		if r[idx]&shifted != shifted {
			out = append(out, indices[i])
		}
		r[idx] |= shifted
	}

	return r, out[:]
}

// registerClosed checks if the specified index is in the closed state (bit value 1).
// Returns true if the index is closed, false otherwise.
func registerClosed(r register, index uint) bool {
	idx, offs := offset(index)
	shifted := shift(offs)
	return r[idx]&shifted == shifted
}

// registerAllClosed checks if all the specified indices are in the closed state.
// Returns true only if all indices are closed, false otherwise.
func registerAllClosed(r register, indices ...uint) bool {
	for i := 0; i < len(indices); i++ {
		if !registerClosed(r, indices[i]) {
			return false
		}
	}

	return true
}

// registerAnyClosed checks if any of the specified indices are in the closed state.
// Returns true if at least one index is closed, false otherwise.
func registerAnyClosed(r register, indices ...uint) bool {
	for i := 0; i < len(indices); i++ {
		if registerClosed(r, indices[i]) {
			return true
		}
	}

	return false
}

// registerOpen sets the specified indices to the open state (bit value 0).
// It returns the modified register and a slice of indices that changed state.
// If an index was already open, it won't be included in the returned slice.
func registerOpen(r register, indices ...uint) (register, []uint) {
	out := make([]uint, 0, len(indices))
	for i := 0; i < len(indices); i++ {
		idx, offs := offset(indices[i])
		shifted := shift(offs)
		if r[idx]&shifted == shifted {
			out = append(out, indices[i])
		}
		r[idx] &^= shifted
	}

	return r, out[:]
}

// registerOpened checks if the specified index is in the open state (bit value 0).
// Returns true if the index is open, false otherwise.
func registerOpened(r register, index uint) bool {
	idx, offs := offset(index)
	shifted := shift(offs)
	return r[idx]&shifted != shifted
}

// registerAllOpened checks if all the specified indices are in the open state.
// Returns true only if all indices are open, false otherwise.
func registerAllOpened(r register, indices ...uint) bool {
	for i := 0; i < len(indices); i++ {
		if !registerOpened(r, indices[i]) {
			return false
		}
	}

	return true
}

// registerAnyOpened checks if any of the specified indices are in the open state.
// Returns true if at least one index is open, false otherwise.
func registerAnyOpened(r register, indices ...uint) bool {
	for i := 0; i < len(indices); i++ {
		if registerOpened(r, indices[i]) {
			return true
		}
	}

	return false
}

// registerToggle switches the state of the specified indices.
// If an index is open, it will be closed, and if it's closed, it will be opened.
// It returns the modified register, a slice of indices that were closed, and a slice of indices that were opened.
func registerToggle(r register, indices ...uint) (register, []uint, []uint) {
	var opened, closed []uint
	for i := 0; i < len(indices); i++ {
		idx, offs := offset(indices[i])
		shifted := shift(offs)

		if r[idx]&shifted != shifted {
			closed = append(closed, indices[i])
			r[idx] |= shifted
		} else {
			opened = append(opened, indices[i])
			r[idx] &^= shifted
		}
	}

	return r, closed[:], opened[:]
}
