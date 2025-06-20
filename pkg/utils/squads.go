package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/gagliardetto/solana-go"
)

// TransactionMessage represents a transaction's message structure.
type TransactionMessage struct {
	NumSigners            uint8
	NumWritableSigners    uint8
	NumWritableNonSigners uint8
	AccountKeys           SmallVec[solana.PublicKey]
	Instructions          SmallVec[CompiledInstruction]
	AddressTableLookups   SmallVec[MessageAddressTableLookup]
}

// Serialize serializes the TransactionMessage into a byte slice.
func (tm *TransactionMessage) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)

	// Serialize num_signers, num_writable_signers, num_writable_non_signers
	if err := buf.WriteByte(tm.NumSigners); err != nil {
		return nil, err
	}
	if err := buf.WriteByte(tm.NumWritableSigners); err != nil {
		return nil, err
	}
	if err := buf.WriteByte(tm.NumWritableNonSigners); err != nil {
		return nil, err
	}

	// Serialize account_keys
	accountKeysBytes, err := tm.AccountKeys.Serialize(func(pk solana.PublicKey) ([]byte, error) {
		return pk[:], nil
	})
	if err != nil {
		return nil, err
	}
	if _, err := buf.Write(accountKeysBytes); err != nil {
		return nil, err
	}

	// Serialize instructions
	instructionsBytes, err := tm.Instructions.Serialize(func(instr CompiledInstruction) ([]byte, error) {
		return instr.Serialize()
	})
	if err != nil {
		return nil, err
	}
	if _, err := buf.Write(instructionsBytes); err != nil {
		return nil, err
	}

	// Serialize address_table_lookups
	lookupsBytes, err := tm.AddressTableLookups.Serialize(func(lookup MessageAddressTableLookup) ([]byte, error) {
		return lookup.Serialize()
	})
	if err != nil {
		return nil, err
	}
	if _, err := buf.Write(lookupsBytes); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// SmallVec is a concise vector serialization wrapper.
// L is typically uint8 or uint16 to represent the length.
type SmallVec[T any] struct {
	Items []T
}

// Len returns the number of elements in the SmallVec.
func (sv *SmallVec[T]) Len() int {
	return len(sv.Items)
}

// IsEmpty checks whether the SmallVec is empty.
func (sv *SmallVec[T]) IsEmpty() bool {
	return len(sv.Items) == 0
}

// ToSlice converts SmallVec to a native slice.
func (sv *SmallVec[T]) ToSlice() []T {
	return sv.Items
}

// NewSmallVecFromSlice creates a SmallVec from a regular slice.
func NewSmallVecFromSlice[T any](items []T) SmallVec[T] {
	return SmallVec[T]{Items: items}
}

// Serialize serializes a SmallVec with length encoded as u8.
func (sv *SmallVec[T]) Serialize(serializeElem func(T) ([]byte, error)) ([]byte, error) {
	if sv.Len() > 255 {
		return nil, fmt.Errorf("SmallVec: length exceeds u8 limit")
	}

	buf := new(bytes.Buffer)
	if err := buf.WriteByte(uint8(sv.Len())); err != nil {
		return nil, err
	}

	for _, elem := range sv.Items {
		data, err := serializeElem(elem)
		if err != nil {
			return nil, err
		}
		if _, err := buf.Write(data); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

// SmallVec16 is like SmallVec, but encodes the length as uint16 (2 bytes).
type SmallVec16[T any] struct {
	Items []T
}

// Len returns the number of elements.
func (sv *SmallVec16[T]) Len() int {
	return len(sv.Items)
}

// IsEmpty checks whether the SmallVec16 is empty.
func (sv *SmallVec16[T]) IsEmpty() bool {
	return len(sv.Items) == 0
}

// ToSlice converts SmallVec16 to a native slice.
func (sv *SmallVec16[T]) ToSlice() []T {
	return sv.Items
}

// NewSmallVec16FromSlice creates a SmallVec16 from a regular slice.
func NewSmallVec16FromSlice[T any](items []T) SmallVec16[T] {
	return SmallVec16[T]{Items: items}
}

// Serialize serializes a SmallVec16 with length encoded as uint16.
func (sv *SmallVec16[T]) Serialize(serializeElem func(T) ([]byte, error)) ([]byte, error) {
	if sv.Len() > 65535 {
		return nil, fmt.Errorf("SmallVec16: length exceeds u16 limit")
	}

	buf := new(bytes.Buffer)

	lenBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(lenBytes, uint16(sv.Len()))
	if _, err := buf.Write(lenBytes); err != nil {
		return nil, err
	}

	for _, elem := range sv.Items {
		data, err := serializeElem(elem)
		if err != nil {
			return nil, err
		}
		if _, err := buf.Write(data); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

// CompiledInstruction represents a compiled instruction inside a transaction.
type CompiledInstruction struct {
	ProgramIDIndex uint8
	AccountIndexes SmallVec[uint8]
	Data           SmallVec16[uint8]
}

// Serialize serializes a CompiledInstruction.
func (ci *CompiledInstruction) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)

	if err := buf.WriteByte(ci.ProgramIDIndex); err != nil {
		return nil, err
	}

	accountIndexesBytes, err := ci.AccountIndexes.Serialize(func(v uint8) ([]byte, error) {
		return []byte{v}, nil
	})
	if err != nil {
		return nil, err
	}
	if _, err := buf.Write(accountIndexesBytes); err != nil {
		return nil, err
	}

	dataBytes, err := ci.Data.Serialize(func(v uint8) ([]byte, error) {
		return []byte{v}, nil
	})
	if err != nil {
		return nil, err
	}
	if _, err := buf.Write(dataBytes); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// MessageAddressTableLookup describes address table lookups for loading extra accounts.
type MessageAddressTableLookup struct {
	AccountKey      solana.PublicKey
	WritableIndexes SmallVec[uint8]
	ReadonlyIndexes SmallVec[uint8]
}

// Serialize serializes a MessageAddressTableLookup.
func (lookup *MessageAddressTableLookup) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)

	if _, err := buf.Write(lookup.AccountKey[:]); err != nil {
		return nil, err
	}

	writableBytes, err := lookup.WritableIndexes.Serialize(func(v uint8) ([]byte, error) {
		return []byte{v}, nil
	})
	if err != nil {
		return nil, err
	}
	if _, err := buf.Write(writableBytes); err != nil {
		return nil, err
	}

	readonlyBytes, err := lookup.ReadonlyIndexes.Serialize(func(v uint8) ([]byte, error) {
		return []byte{v}, nil
	})
	if err != nil {
		return nil, err
	}
	if _, err := buf.Write(readonlyBytes); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
