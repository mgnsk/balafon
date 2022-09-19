package ast

import (
	"math"

	"github.com/mgnsk/gong/internal/constants"
	"github.com/mgnsk/gong/internal/parser/token"
)

// CmdAssign is a note assignment command.
type CmdAssign struct {
	Note rune
	Key  uint8
}

// NewCmdAssign creates a note assignment command.
func NewCmdAssign(note, key *token.Token) (CmdAssign, error) {
	v, err := key.Int32Value()
	if err != nil {
		return CmdAssign{}, err
	}
	if err := validateRange(v, 0, constants.MaxValue); err != nil {
		return CmdAssign{}, err
	}
	return CmdAssign{
		Note: []rune(note.IDValue())[0],
		Key:  uint8(v),
	}, nil
}

// CmdTempo is a tempo command.
type CmdTempo uint16

// NewCmdTempo creates a tempo command.
func NewCmdTempo(bpm *token.Token) (CmdTempo, error) {
	v, err := bpm.Int32Value()
	if err != nil {
		return 0, err
	}
	if err := validateRange(v, 1, math.MaxUint16); err != nil {
		return 0, err
	}
	return CmdTempo(v), nil
}

// CmdTimeSig is a time signature change command.
type CmdTimeSig struct {
	Num   uint8
	Denom uint8
}

// NewCmdTimeSig creates a time signature change command.
func NewCmdTimeSig(num, denom *token.Token) (CmdTimeSig, error) {
	b, err := num.Int32Value()
	if err != nil {
		return CmdTimeSig{}, err
	}
	v, err := denom.Int32Value()
	if err != nil {
		return CmdTimeSig{}, err
	}
	if err := validateRange(b, 1, constants.MaxBeatsPerBar); err != nil {
		return CmdTimeSig{}, err
	}
	if err := validateNoteValue(int(v)); err != nil {
		return CmdTimeSig{}, err
	}
	return CmdTimeSig{
		Num:   uint8(b),
		Denom: uint8(v),
	}, nil
}

// CmdChannel is a channel change command.
type CmdChannel uint8

// NewCmdChannel creates a channel change command.
func NewCmdChannel(value *token.Token) (CmdChannel, error) {
	v, err := value.Int32Value()
	if err != nil {
		return 0, err
	}
	if err := validateRange(v, 0, constants.MaxChannel); err != nil {
		return 0, err
	}
	return CmdChannel(v), nil
}

// CmdVelocity is a velocity change command.
type CmdVelocity uint8

// NewCmdVelocity creates a velocity change command.
func NewCmdVelocity(value *token.Token) (CmdVelocity, error) {
	v, err := value.Int32Value()
	if err != nil {
		return 0, err
	}
	if err := validateRange(v, 0, constants.MaxValue); err != nil {
		return 0, err
	}
	return CmdVelocity(v), nil
}

// CmdProgram is a program change command.
type CmdProgram uint8

// NewCmdProgram creates a program change command.
func NewCmdProgram(value *token.Token) (CmdProgram, error) {
	v, err := value.Int32Value()
	if err != nil {
		return 0, err
	}
	if err := validateRange(v, 0, constants.MaxValue); err != nil {
		return 0, err
	}
	return CmdProgram(v), nil
}

// CmdControl is a control change command.
type CmdControl struct {
	Control   uint8
	Parameter uint8
}

// NewCmdControl creates a control change command.
func NewCmdControl(control, value *token.Token) (CmdControl, error) {
	c, err := control.Int32Value()
	if err != nil {
		return CmdControl{}, err
	}
	if err := validateRange(c, 0, constants.MaxValue); err != nil {
		return CmdControl{}, err
	}
	v, err := value.Int32Value()
	if err != nil {
		return CmdControl{}, err
	}
	if err := validateRange(v, 0, constants.MaxValue); err != nil {
		return CmdControl{}, err
	}
	return CmdControl{
		Control:   uint8(c),
		Parameter: uint8(v),
	}, nil
}

// CmdBar is a bar begin command.
type CmdBar string

// CmdEnd is a bar end command.
type CmdEnd struct{}

// CmdPlay is a bar play command.
type CmdPlay string

// CmdStart is a start commad.
type CmdStart struct{}

// CmdStop is a stop command.
type CmdStop struct{}
