// Code generated by gocc; DO NOT EDIT.

package parser

type (
	actionTable [numStates]actionRow
	actionRow   struct {
		canRecover bool
		actions    [numSymbols]action
	}
)

var actionTab = actionTable{
	actionRow{ // S0
		canRecover: false,
		actions: [numSymbols]action{
			nil,       // INVALID
			reduce(5), // $, reduce: Track
			shift(5),  // ident
			nil,       // =
			nil,       // uint64
			nil,       // empty
			nil,       // dot
			nil,       // tuplet
			shift(7),  // bar
			shift(8),  // play
			shift(9),  // end
		},
	},
	actionRow{ // S1
		canRecover: false,
		actions: [numSymbols]action{
			nil,          // INVALID
			accept(true), // $
			nil,          // ident
			nil,          // =
			nil,          // uint64
			nil,          // empty
			nil,          // dot
			nil,          // tuplet
			nil,          // bar
			nil,          // play
			nil,          // end
		},
	},
	actionRow{ // S2
		canRecover: false,
		actions: [numSymbols]action{
			nil,       // INVALID
			reduce(1), // $, reduce: Expr
			nil,       // ident
			nil,       // =
			nil,       // uint64
			nil,       // empty
			nil,       // dot
			nil,       // tuplet
			nil,       // bar
			nil,       // play
			nil,       // end
		},
	},
	actionRow{ // S3
		canRecover: false,
		actions: [numSymbols]action{
			nil,       // INVALID
			reduce(2), // $, reduce: Expr
			nil,       // ident
			nil,       // =
			nil,       // uint64
			nil,       // empty
			nil,       // dot
			nil,       // tuplet
			nil,       // bar
			nil,       // play
			nil,       // end
		},
	},
	actionRow{ // S4
		canRecover: false,
		actions: [numSymbols]action{
			nil,       // INVALID
			reduce(3), // $, reduce: Expr
			nil,       // ident
			nil,       // =
			nil,       // uint64
			nil,       // empty
			nil,       // dot
			nil,       // tuplet
			nil,       // bar
			nil,       // play
			nil,       // end
		},
	},
	actionRow{ // S5
		canRecover: false,
		actions: [numSymbols]action{
			nil,       // INVALID
			reduce(8), // $, reduce: PropertyList
			reduce(8), // ident, reduce: PropertyList
			shift(10), // =
			shift(11), // uint64
			nil,       // empty
			shift(13), // dot
			shift(14), // tuplet
			nil,       // bar
			nil,       // play
			nil,       // end
		},
	},
	actionRow{ // S6
		canRecover: false,
		actions: [numSymbols]action{
			nil,       // INVALID
			reduce(5), // $, reduce: Track
			shift(16), // ident
			nil,       // =
			nil,       // uint64
			nil,       // empty
			nil,       // dot
			nil,       // tuplet
			nil,       // bar
			nil,       // play
			nil,       // end
		},
	},
	actionRow{ // S7
		canRecover: false,
		actions: [numSymbols]action{
			nil,       // INVALID
			nil,       // $
			shift(17), // ident
			nil,       // =
			nil,       // uint64
			nil,       // empty
			nil,       // dot
			nil,       // tuplet
			nil,       // bar
			nil,       // play
			nil,       // end
		},
	},
	actionRow{ // S8
		canRecover: false,
		actions: [numSymbols]action{
			nil,       // INVALID
			nil,       // $
			shift(18), // ident
			nil,       // =
			nil,       // uint64
			nil,       // empty
			nil,       // dot
			nil,       // tuplet
			nil,       // bar
			nil,       // play
			nil,       // end
		},
	},
	actionRow{ // S9
		canRecover: false,
		actions: [numSymbols]action{
			nil,        // INVALID
			reduce(14), // $, reduce: Command
			nil,        // ident
			nil,        // =
			nil,        // uint64
			nil,        // empty
			nil,        // dot
			nil,        // tuplet
			nil,        // bar
			nil,        // play
			nil,        // end
		},
	},
	actionRow{ // S10
		canRecover: false,
		actions: [numSymbols]action{
			nil,       // INVALID
			nil,       // $
			nil,       // ident
			nil,       // =
			shift(19), // uint64
			nil,       // empty
			nil,       // dot
			nil,       // tuplet
			nil,       // bar
			nil,       // play
			nil,       // end
		},
	},
	actionRow{ // S11
		canRecover: false,
		actions: [numSymbols]action{
			nil,       // INVALID
			reduce(8), // $, reduce: PropertyList
			reduce(8), // ident, reduce: PropertyList
			nil,       // =
			shift(11), // uint64
			nil,       // empty
			shift(13), // dot
			shift(14), // tuplet
			nil,       // bar
			nil,       // play
			nil,       // end
		},
	},
	actionRow{ // S12
		canRecover: false,
		actions: [numSymbols]action{
			nil,       // INVALID
			reduce(7), // $, reduce: NoteList
			reduce(7), // ident, reduce: NoteList
			nil,       // =
			nil,       // uint64
			nil,       // empty
			nil,       // dot
			nil,       // tuplet
			nil,       // bar
			nil,       // play
			nil,       // end
		},
	},
	actionRow{ // S13
		canRecover: false,
		actions: [numSymbols]action{
			nil,       // INVALID
			reduce(8), // $, reduce: PropertyList
			reduce(8), // ident, reduce: PropertyList
			nil,       // =
			shift(11), // uint64
			nil,       // empty
			shift(13), // dot
			shift(14), // tuplet
			nil,       // bar
			nil,       // play
			nil,       // end
		},
	},
	actionRow{ // S14
		canRecover: false,
		actions: [numSymbols]action{
			nil,       // INVALID
			reduce(8), // $, reduce: PropertyList
			reduce(8), // ident, reduce: PropertyList
			nil,       // =
			shift(11), // uint64
			nil,       // empty
			shift(13), // dot
			shift(14), // tuplet
			nil,       // bar
			nil,       // play
			nil,       // end
		},
	},
	actionRow{ // S15
		canRecover: false,
		actions: [numSymbols]action{
			nil,       // INVALID
			reduce(6), // $, reduce: Track
			nil,       // ident
			nil,       // =
			nil,       // uint64
			nil,       // empty
			nil,       // dot
			nil,       // tuplet
			nil,       // bar
			nil,       // play
			nil,       // end
		},
	},
	actionRow{ // S16
		canRecover: false,
		actions: [numSymbols]action{
			nil,       // INVALID
			reduce(8), // $, reduce: PropertyList
			reduce(8), // ident, reduce: PropertyList
			nil,       // =
			shift(11), // uint64
			nil,       // empty
			shift(13), // dot
			shift(14), // tuplet
			nil,       // bar
			nil,       // play
			nil,       // end
		},
	},
	actionRow{ // S17
		canRecover: false,
		actions: [numSymbols]action{
			nil,        // INVALID
			reduce(12), // $, reduce: Command
			nil,        // ident
			nil,        // =
			nil,        // uint64
			nil,        // empty
			nil,        // dot
			nil,        // tuplet
			nil,        // bar
			nil,        // play
			nil,        // end
		},
	},
	actionRow{ // S18
		canRecover: false,
		actions: [numSymbols]action{
			nil,        // INVALID
			reduce(13), // $, reduce: Command
			nil,        // ident
			nil,        // =
			nil,        // uint64
			nil,        // empty
			nil,        // dot
			nil,        // tuplet
			nil,        // bar
			nil,        // play
			nil,        // end
		},
	},
	actionRow{ // S19
		canRecover: false,
		actions: [numSymbols]action{
			nil,       // INVALID
			reduce(4), // $, reduce: Assignment
			nil,       // ident
			nil,       // =
			nil,       // uint64
			nil,       // empty
			nil,       // dot
			nil,       // tuplet
			nil,       // bar
			nil,       // play
			nil,       // end
		},
	},
	actionRow{ // S20
		canRecover: false,
		actions: [numSymbols]action{
			nil,       // INVALID
			reduce(9), // $, reduce: PropertyList
			reduce(9), // ident, reduce: PropertyList
			nil,       // =
			nil,       // uint64
			nil,       // empty
			nil,       // dot
			nil,       // tuplet
			nil,       // bar
			nil,       // play
			nil,       // end
		},
	},
	actionRow{ // S21
		canRecover: false,
		actions: [numSymbols]action{
			nil,        // INVALID
			reduce(10), // $, reduce: PropertyList
			reduce(10), // ident, reduce: PropertyList
			nil,        // =
			nil,        // uint64
			nil,        // empty
			nil,        // dot
			nil,        // tuplet
			nil,        // bar
			nil,        // play
			nil,        // end
		},
	},
	actionRow{ // S22
		canRecover: false,
		actions: [numSymbols]action{
			nil,        // INVALID
			reduce(11), // $, reduce: PropertyList
			reduce(11), // ident, reduce: PropertyList
			nil,        // =
			nil,        // uint64
			nil,        // empty
			nil,        // dot
			nil,        // tuplet
			nil,        // bar
			nil,        // play
			nil,        // end
		},
	},
}
