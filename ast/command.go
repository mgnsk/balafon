package ast

import (
	"bytes"
	"io"
	"math"

	"github.com/mgnsk/balafon/constants"
	"github.com/mgnsk/balafon/internal/parser/token"
)

// CmdAssign is a note assignment command.
type CmdAssign struct {
	Note rune
	Key  int
}

func (c CmdAssign) WriteTo(w io.Writer) (int64, error) {
	ew := newErrWriter(w)
	var n int

	n += ew.WriteString(":assign ")
	n += ew.WriteRune(c.Note)
	n += ew.WriteString(" ")
	n += ew.WriteInt(int(c.Key))

	return int64(n), ew.Flush()
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
		Key:  int(v),
	}, nil
}

// CmdTempo is a tempo command.
type CmdTempo uint16

// Value returns the tempo value.
func (c CmdTempo) Value() float64 {
	return float64(c)
}

func (c CmdTempo) WriteTo(w io.Writer) (int64, error) {
	ew := newErrWriter(w)
	var n int

	n += ew.WriteString(":tempo ")
	n += ew.WriteInt(int(c))

	return int64(n), ew.Flush()
}

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
type CmdTimeSig [2]uint8

func (c CmdTimeSig) WriteTo(w io.Writer) (int64, error) {
	ew := newErrWriter(w)
	var n int

	n += ew.WriteString(":timesig ")
	n += ew.WriteInt(int(c[0]))
	n += ew.WriteString(" ")
	n += ew.WriteInt(int(c[1]))

	return int64(n), ew.Flush()
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
		uint8(b),
		uint8(v),
	}, nil
}

// CmdChannel is a channel change command.
type CmdChannel uint8

// Value returns the channel's value.
func (c CmdChannel) Value() uint8 {
	return uint8(c)
}

func (c CmdChannel) WriteTo(w io.Writer) (int64, error) {
	ew := newErrWriter(w)
	var n int

	n += ew.WriteString(":channel ")
	n += ew.WriteInt(int(c))

	return int64(n), ew.Flush()
}

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
type CmdVelocity int

// Value returns the velocity value.
func (c CmdVelocity) Value() int {
	return int(c)
}

func (c CmdVelocity) WriteTo(w io.Writer) (int64, error) {
	ew := newErrWriter(w)
	var n int

	n += ew.WriteString(":velocity ")
	n += ew.WriteInt(int(c))

	return int64(n), ew.Flush()
}

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

// Value returns the program value.
func (c CmdProgram) Value() uint8 {
	return uint8(c)
}

func (c CmdProgram) WriteTo(w io.Writer) (int64, error) {
	ew := newErrWriter(w)
	var n int

	n += ew.WriteString(":program ")
	n += ew.WriteInt(int(c))

	return int64(n), ew.Flush()
}

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

func (c CmdControl) WriteTo(w io.Writer) (int64, error) {
	ew := newErrWriter(w)
	var n int

	n += ew.WriteString(":control ")
	n += ew.WriteInt(int(c.Control))
	n += ew.WriteString(" ")
	n += ew.WriteInt(int(c.Parameter))

	return int64(n), ew.Flush()
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

// CmdPlay is a bar play command.
type CmdPlay struct {
	Name  string
	Token *token.Token
}

func NewCmdPlay(t *token.Token) (CmdPlay, error) {
	return CmdPlay{
		Name:  string(bytes.TrimPrefix(t.Lit, []byte(":play "))),
		Token: t,
	}, nil
}

func (c CmdPlay) WriteTo(w io.Writer) (int64, error) {
	ew := newErrWriter(w)
	var n int

	n += ew.WriteString(":play ")
	n += ew.WriteString(c.Name)

	return int64(n), ew.Flush()
}

// CmdStart is a start commad.
type CmdStart struct{}

func (c CmdStart) WriteTo(w io.Writer) (int64, error) {
	ew := newErrWriter(w)
	var n int

	n += ew.WriteString(":start")

	return int64(n), ew.Flush()
}

// CmdStop is a stop command.
type CmdStop struct{}

func (c CmdStop) WriteTo(w io.Writer) (int64, error) {
	ew := newErrWriter(w)
	var n int

	n += ew.WriteString(":stop")

	return int64(n), ew.Flush()
}
