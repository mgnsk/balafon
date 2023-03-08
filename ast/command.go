package ast

import (
	"io"
	"math"

	"github.com/mgnsk/balafon/constants"
	"github.com/mgnsk/balafon/internal/parser/token"
)

// CmdAssign is a note assignment command.
type CmdAssign struct {
	Note rune
	Key  uint8
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

// func (c CmdAssign) String() string {
// 	return fmt.Sprintf("assign %c %d", c.Note, c.Key)
// }

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

func (c CmdTempo) WriteTo(w io.Writer) (int64, error) {
	ew := newErrWriter(w)
	var n int

	n += ew.WriteString(":tempo ")
	n += ew.WriteInt(int(c))

	return int64(n), ew.Flush()
}

// func (c CmdTempo) String() string {
// 	return fmt.Sprintf("tempo %d", c)
// }

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

// func (c CmdTimeSig) String() string {
// 	return fmt.Sprintf("timesig %d %d", c.Num, c.Denom)
// }

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

func (c CmdChannel) WriteTo(w io.Writer) (int64, error) {
	ew := newErrWriter(w)
	var n int

	n += ew.WriteString(":channel ")
	n += ew.WriteInt(int(c))

	return int64(n), ew.Flush()
}

// func (c CmdChannel) String() string {
// 	return fmt.Sprintf("channel %d", c)
// }

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

func (c CmdVelocity) WriteTo(w io.Writer) (int64, error) {
	ew := newErrWriter(w)
	var n int

	n += ew.WriteString(":velocity ")
	n += ew.WriteInt(int(c))

	return int64(n), ew.Flush()
}

// func (c CmdVelocity) String() string {
// 	return fmt.Sprintf("velocity %d", c)
// }

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

func (c CmdProgram) WriteTo(w io.Writer) (int64, error) {
	ew := newErrWriter(w)
	var n int

	n += ew.WriteString(":program ")
	n += ew.WriteInt(int(c))

	return int64(n), ew.Flush()
}

// func (c CmdProgram) String() string {
// 	return fmt.Sprintf("program %d", c)
// }

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

// func (c CmdControl) String() string {
// 	return fmt.Sprintf("control %d %d", c.Control, c.Parameter)
// }

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

func NewCmdName(t *token.Token) (CmdPlay, error) {
	return CmdPlay{
		Name:  t.StringValue(),
		Token: t,
	}, nil
}

func (c CmdPlay) WriteTo(w io.Writer) (int64, error) {
	ew := newErrWriter(w)
	var n int

	n += ew.WriteString(":play \"")
	n += ew.WriteString(c.Name)
	n += ew.WriteString("\"")

	return int64(n), ew.Flush()
}

// func (c CmdPlay) String() string {
// 	return fmt.Sprintf(`play "%s"`, string(c))
// }

// CmdStart is a start commad.
type CmdStart struct{}

func (c CmdStart) WriteTo(w io.Writer) (int64, error) {
	ew := newErrWriter(w)
	var n int

	n += ew.WriteString(":start")

	return int64(n), ew.Flush()
}

// func (c CmdStart) String() string {
// 	return "start"
// }

// CmdStop is a stop command.
type CmdStop struct{}

func (c CmdStop) WriteTo(w io.Writer) (int64, error) {
	ew := newErrWriter(w)
	var n int

	n += ew.WriteString(":stop")

	return int64(n), ew.Flush()
}

// func (c CmdStop) String() string {
// 	return "stop"
// }
