package ast

import (
	"io"
	"math"

	"github.com/mgnsk/balafon/constants"
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
func NewCmdAssign(note rune, key int64) (CmdAssign, error) {
	if err := validateRange(key, 0, constants.MaxValue); err != nil {
		return CmdAssign{}, err
	}
	return CmdAssign{
		Note: note,
		Key:  int(key),
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
func NewCmdTempo(bpm int64) (CmdTempo, error) {
	if err := validateRange(bpm, 1, math.MaxUint16); err != nil {
		return 0, err
	}
	return CmdTempo(bpm), nil
}

// CmdTimeSig is a time signature change command.
type CmdTimeSig [2]uint8

// Num returns the timesig's numerator.
func (c CmdTimeSig) Num() uint8 {
	return c[0]
}

// Denom returns the timesig's denominator.
func (c CmdTimeSig) Denom() uint8 {
	return c[1]
}

func (c CmdTimeSig) WriteTo(w io.Writer) (int64, error) {
	ew := newErrWriter(w)
	var n int

	n += ew.WriteString(":timesig ")
	n += ew.WriteInt(int(c.Num()))
	n += ew.WriteString(" ")
	n += ew.WriteInt(int(c.Denom()))

	return int64(n), ew.Flush()
}

// NewCmdTimeSig creates a time signature change command.
func NewCmdTimeSig(num, denom int64) (CmdTimeSig, error) {
	if err := validateRange(num, 1, constants.MaxBeatsPerBar); err != nil {
		return CmdTimeSig{}, err
	}
	if err := validateNoteValue(int(denom)); err != nil {
		return CmdTimeSig{}, err
	}
	return CmdTimeSig{
		uint8(num),
		uint8(denom),
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
func NewCmdChannel(value int64) (CmdChannel, error) {
	if err := validateRange(value, 0, constants.MaxChannel); err != nil {
		return 0, err
	}
	return CmdChannel(value), nil
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
func NewCmdVelocity(value int64) (CmdVelocity, error) {
	if err := validateRange(value, 0, constants.MaxValue); err != nil {
		return 0, err
	}
	return CmdVelocity(value), nil
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
func NewCmdProgram(value int64) (CmdProgram, error) {
	if err := validateRange(value, 0, constants.MaxValue); err != nil {
		return 0, err
	}
	return CmdProgram(value), nil
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
func NewCmdControl(control, value int64) (CmdControl, error) {
	if err := validateRange(control, 0, constants.MaxValue); err != nil {
		return CmdControl{}, err
	}
	if err := validateRange(value, 0, constants.MaxValue); err != nil {
		return CmdControl{}, err
	}
	return CmdControl{
		Control:   uint8(control),
		Parameter: uint8(value),
	}, nil
}

// CmdPlay is a bar play command.
type CmdPlay struct {
	BarName string
}

func NewCmdPlay(barName string) (CmdPlay, error) {
	return CmdPlay{
		BarName: barName,
	}, nil
}

func (c CmdPlay) WriteTo(w io.Writer) (int64, error) {
	ew := newErrWriter(w)
	var n int

	n += ew.WriteString(":play ")
	n += ew.WriteString(c.BarName)

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
