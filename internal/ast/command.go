package ast

import (
	"io"
	"math"

	"github.com/mgnsk/balafon/internal/constants"
	"github.com/mgnsk/balafon/internal/parser/token"
)

// CmdAssign is a note assignment command.
type CmdAssign struct {
	Pos  token.Pos
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
func NewCmdAssign(pos token.Pos, note rune, key int64) (CmdAssign, error) {
	if err := validateRange(key, 0, constants.MaxValue); err != nil {
		return CmdAssign{}, err
	}
	return CmdAssign{
		Pos:  pos,
		Note: note,
		Key:  int(key),
	}, nil
}

// CmdTempo is a tempo command.
type CmdTempo struct {
	BPM uint16
}

// Value returns the tempo value.
func (c CmdTempo) Value() float64 {
	return float64(c.BPM)
}

func (c CmdTempo) WriteTo(w io.Writer) (int64, error) {
	ew := newErrWriter(w)
	var n int

	n += ew.WriteString(":tempo ")
	n += ew.WriteInt(int(c.BPM))

	return int64(n), ew.Flush()
}

// NewCmdTempo creates a tempo command.
func NewCmdTempo(bpm int64) (CmdTempo, error) {
	if err := validateRange(bpm, 1, math.MaxUint16); err != nil {
		return CmdTempo{}, err
	}

	return CmdTempo{
		BPM: uint16(bpm),
	}, nil
}

// CmdTimeSig is a time signature change command.
type CmdTimeSig struct {
	Num   uint8
	Denom uint8
}

func (c CmdTimeSig) WriteTo(w io.Writer) (int64, error) {
	ew := newErrWriter(w)
	var n int

	n += ew.WriteString(":timesig ")
	n += ew.WriteInt(int(c.Num))
	n += ew.WriteString(" ")
	n += ew.WriteInt(int(c.Denom))

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
		Num:   uint8(num),
		Denom: uint8(denom),
	}, nil
}

// CmdChannel is a channel change command.
type CmdChannel struct {
	Channel uint8
}

func (c CmdChannel) WriteTo(w io.Writer) (int64, error) {
	ew := newErrWriter(w)
	var n int

	n += ew.WriteString(":channel ")
	n += ew.WriteInt(int(c.Channel))

	return int64(n), ew.Flush()
}

// NewCmdChannel creates a channel change command.
func NewCmdChannel(value int64) (CmdChannel, error) {
	if err := validateRange(value, 0, constants.MaxChannel); err != nil {
		return CmdChannel{}, err
	}

	return CmdChannel{
		Channel: uint8(value),
	}, nil
}

// CmdVelocity is a velocity change command.
type CmdVelocity struct {
	Velocity int
}

func (c CmdVelocity) WriteTo(w io.Writer) (int64, error) {
	ew := newErrWriter(w)
	var n int

	n += ew.WriteString(":velocity ")
	n += ew.WriteInt(c.Velocity)

	return int64(n), ew.Flush()
}

// NewCmdVelocity creates a velocity change command.
func NewCmdVelocity(value int64) (CmdVelocity, error) {
	if err := validateRange(value, 0, constants.MaxValue); err != nil {
		return CmdVelocity{}, err
	}

	return CmdVelocity{
		Velocity: int(value),
	}, nil
}

// CmdProgram is a program change command.
type CmdProgram struct {
	Program uint8
}

func (c CmdProgram) WriteTo(w io.Writer) (int64, error) {
	ew := newErrWriter(w)
	var n int

	n += ew.WriteString(":program ")
	n += ew.WriteInt(int(c.Program))

	return int64(n), ew.Flush()
}

// NewCmdProgram creates a program change command.
func NewCmdProgram(value int64) (CmdProgram, error) {
	if err := validateRange(value, 0, constants.MaxValue); err != nil {
		return CmdProgram{}, err
	}

	return CmdProgram{
		Program: uint8(value),
	}, nil
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
	Pos     token.Pos
	BarName string
}

func NewCmdPlay(pos token.Pos, barName string) (CmdPlay, error) {
	return CmdPlay{
		Pos:     pos,
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
