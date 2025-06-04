package main

import (
	"encoding/binary"
	"io"

	"github.com/alecthomas/errors"

	sm "github.com/block/ftl/internal/statemachine"
)

type IntStateMachine struct {
	sum int64
}

type IntEvent int64

func (i *IntEvent) UnmarshalBinary(data []byte) error { //nolint:unparam
	*i = IntEvent(binary.BigEndian.Uint64(data))
	return nil
}

func (i IntEvent) MarshalBinary() ([]byte, error) { //nolint:unparam
	return binary.BigEndian.AppendUint64([]byte{}, uint64(i)), nil
}

var _ sm.Snapshotting[int64, int64, IntEvent] = &IntStateMachine{}

func (s IntStateMachine) Lookup(key int64) (int64, error) {
	return s.sum, nil
}

func (s *IntStateMachine) Publish(msg IntEvent) error {
	s.sum += int64(msg)
	return nil
}

func (s IntStateMachine) Close() error {
	return nil
}

func (s IntStateMachine) Recover(reader io.Reader) error {
	err := binary.Read(reader, binary.BigEndian, &s.sum)
	if err != nil {
		return errors.Wrap(err, "failed to recover from snapshot")
	}
	return nil
}

func (s IntStateMachine) Save(writer io.Writer) error {
	err := binary.Write(writer, binary.BigEndian, s.sum)
	if err != nil {
		return errors.Wrap(err, "failed to save snapshot")
	}
	return nil
}
