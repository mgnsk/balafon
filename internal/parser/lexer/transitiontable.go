// Code generated by gocc; DO NOT EDIT.

package lexer

/*
Let s be the current state
Let r be the current input rune
transitionTable[s](r) returns the next state.
*/
type TransitionTable [NumStates]func(rune) int

var TransTab = TransitionTable{
	// S0
	func(r rune) int {
		switch {
		case r == 9: // ['\t','\t']
			return 1
		case r == 10: // ['\n','\n']
			return 2
		case r == 13: // ['\r','\r']
			return 1
		case r == 32: // [' ',' ']
			return 1
		case r == 35: // ['#','#']
			return 3
		case r == 36: // ['$','$']
			return 4
		case r == 41: // [')',')']
			return 5
		case r == 42: // ['*','*']
			return 6
		case r == 45: // ['-','-']
			return 7
		case r == 46: // ['.','.']
			return 8
		case r == 47: // ['/','/']
			return 9
		case r == 48: // ['0','0']
			return 10
		case 49 <= r && r <= 57: // ['1','9']
			return 11
		case r == 58: // [':',':']
			return 12
		case r == 59: // [';',';']
			return 13
		case r == 62: // ['>','>']
			return 14
		case 65 <= r && r <= 90: // ['A','Z']
			return 15
		case r == 91: // ['[','[']
			return 16
		case r == 93: // [']',']']
			return 17
		case r == 94: // ['^','^']
			return 18
		case 97 <= r && r <= 122: // ['a','z']
			return 15
		}
		return NoState
	},
	// S1
	func(r rune) int {
		switch {
		}
		return NoState
	},
	// S2
	func(r rune) int {
		switch {
		}
		return NoState
	},
	// S3
	func(r rune) int {
		switch {
		}
		return NoState
	},
	// S4
	func(r rune) int {
		switch {
		}
		return NoState
	},
	// S5
	func(r rune) int {
		switch {
		}
		return NoState
	},
	// S6
	func(r rune) int {
		switch {
		}
		return NoState
	},
	// S7
	func(r rune) int {
		switch {
		}
		return NoState
	},
	// S8
	func(r rune) int {
		switch {
		}
		return NoState
	},
	// S9
	func(r rune) int {
		switch {
		case r == 42: // ['*','*']
			return 19
		case r == 47: // ['/','/']
			return 20
		case r == 51: // ['3','3']
			return 21
		case r == 53: // ['5','5']
			return 21
		}
		return NoState
	},
	// S10
	func(r rune) int {
		switch {
		}
		return NoState
	},
	// S11
	func(r rune) int {
		switch {
		case 48 <= r && r <= 57: // ['0','9']
			return 11
		}
		return NoState
	},
	// S12
	func(r rune) int {
		switch {
		case r == 97: // ['a','a']
			return 22
		case r == 98: // ['b','b']
			return 23
		case r == 99: // ['c','c']
			return 24
		case r == 101: // ['e','e']
			return 25
		case r == 112: // ['p','p']
			return 26
		case r == 115: // ['s','s']
			return 27
		case r == 116: // ['t','t']
			return 28
		case r == 118: // ['v','v']
			return 29
		}
		return NoState
	},
	// S13
	func(r rune) int {
		switch {
		case r == 10: // ['\n','\n']
			return 2
		}
		return NoState
	},
	// S14
	func(r rune) int {
		switch {
		}
		return NoState
	},
	// S15
	func(r rune) int {
		switch {
		}
		return NoState
	},
	// S16
	func(r rune) int {
		switch {
		}
		return NoState
	},
	// S17
	func(r rune) int {
		switch {
		}
		return NoState
	},
	// S18
	func(r rune) int {
		switch {
		}
		return NoState
	},
	// S19
	func(r rune) int {
		switch {
		case r == 42: // ['*','*']
			return 30
		default:
			return 19
		}
	},
	// S20
	func(r rune) int {
		switch {
		case r == 10: // ['\n','\n']
			return 31
		default:
			return 20
		}
	},
	// S21
	func(r rune) int {
		switch {
		}
		return NoState
	},
	// S22
	func(r rune) int {
		switch {
		case r == 115: // ['s','s']
			return 32
		}
		return NoState
	},
	// S23
	func(r rune) int {
		switch {
		case r == 97: // ['a','a']
			return 33
		}
		return NoState
	},
	// S24
	func(r rune) int {
		switch {
		case r == 104: // ['h','h']
			return 34
		case r == 111: // ['o','o']
			return 35
		}
		return NoState
	},
	// S25
	func(r rune) int {
		switch {
		case r == 110: // ['n','n']
			return 36
		}
		return NoState
	},
	// S26
	func(r rune) int {
		switch {
		case r == 108: // ['l','l']
			return 37
		case r == 114: // ['r','r']
			return 38
		}
		return NoState
	},
	// S27
	func(r rune) int {
		switch {
		case r == 116: // ['t','t']
			return 39
		}
		return NoState
	},
	// S28
	func(r rune) int {
		switch {
		case r == 101: // ['e','e']
			return 40
		case r == 105: // ['i','i']
			return 41
		}
		return NoState
	},
	// S29
	func(r rune) int {
		switch {
		case r == 101: // ['e','e']
			return 42
		}
		return NoState
	},
	// S30
	func(r rune) int {
		switch {
		case r == 42: // ['*','*']
			return 30
		case r == 47: // ['/','/']
			return 43
		default:
			return 19
		}
	},
	// S31
	func(r rune) int {
		switch {
		}
		return NoState
	},
	// S32
	func(r rune) int {
		switch {
		case r == 115: // ['s','s']
			return 44
		}
		return NoState
	},
	// S33
	func(r rune) int {
		switch {
		case r == 114: // ['r','r']
			return 45
		}
		return NoState
	},
	// S34
	func(r rune) int {
		switch {
		case r == 97: // ['a','a']
			return 46
		}
		return NoState
	},
	// S35
	func(r rune) int {
		switch {
		case r == 110: // ['n','n']
			return 47
		}
		return NoState
	},
	// S36
	func(r rune) int {
		switch {
		case r == 100: // ['d','d']
			return 48
		}
		return NoState
	},
	// S37
	func(r rune) int {
		switch {
		case r == 97: // ['a','a']
			return 49
		}
		return NoState
	},
	// S38
	func(r rune) int {
		switch {
		case r == 111: // ['o','o']
			return 50
		}
		return NoState
	},
	// S39
	func(r rune) int {
		switch {
		case r == 97: // ['a','a']
			return 51
		case r == 111: // ['o','o']
			return 52
		}
		return NoState
	},
	// S40
	func(r rune) int {
		switch {
		case r == 109: // ['m','m']
			return 53
		}
		return NoState
	},
	// S41
	func(r rune) int {
		switch {
		case r == 109: // ['m','m']
			return 54
		}
		return NoState
	},
	// S42
	func(r rune) int {
		switch {
		case r == 108: // ['l','l']
			return 55
		}
		return NoState
	},
	// S43
	func(r rune) int {
		switch {
		}
		return NoState
	},
	// S44
	func(r rune) int {
		switch {
		case r == 105: // ['i','i']
			return 56
		}
		return NoState
	},
	// S45
	func(r rune) int {
		switch {
		case r == 32: // [' ',' ']
			return 57
		}
		return NoState
	},
	// S46
	func(r rune) int {
		switch {
		case r == 110: // ['n','n']
			return 58
		}
		return NoState
	},
	// S47
	func(r rune) int {
		switch {
		case r == 116: // ['t','t']
			return 59
		}
		return NoState
	},
	// S48
	func(r rune) int {
		switch {
		}
		return NoState
	},
	// S49
	func(r rune) int {
		switch {
		case r == 121: // ['y','y']
			return 60
		}
		return NoState
	},
	// S50
	func(r rune) int {
		switch {
		case r == 103: // ['g','g']
			return 61
		}
		return NoState
	},
	// S51
	func(r rune) int {
		switch {
		case r == 114: // ['r','r']
			return 62
		}
		return NoState
	},
	// S52
	func(r rune) int {
		switch {
		case r == 112: // ['p','p']
			return 63
		}
		return NoState
	},
	// S53
	func(r rune) int {
		switch {
		case r == 112: // ['p','p']
			return 64
		}
		return NoState
	},
	// S54
	func(r rune) int {
		switch {
		case r == 101: // ['e','e']
			return 65
		}
		return NoState
	},
	// S55
	func(r rune) int {
		switch {
		case r == 111: // ['o','o']
			return 66
		}
		return NoState
	},
	// S56
	func(r rune) int {
		switch {
		case r == 103: // ['g','g']
			return 67
		}
		return NoState
	},
	// S57
	func(r rune) int {
		switch {
		case r == 48: // ['0','0']
			return 68
		case 49 <= r && r <= 57: // ['1','9']
			return 69
		case 65 <= r && r <= 90: // ['A','Z']
			return 70
		case 97 <= r && r <= 122: // ['a','z']
			return 70
		}
		return NoState
	},
	// S58
	func(r rune) int {
		switch {
		case r == 110: // ['n','n']
			return 71
		}
		return NoState
	},
	// S59
	func(r rune) int {
		switch {
		case r == 114: // ['r','r']
			return 72
		}
		return NoState
	},
	// S60
	func(r rune) int {
		switch {
		case r == 32: // [' ',' ']
			return 73
		}
		return NoState
	},
	// S61
	func(r rune) int {
		switch {
		case r == 114: // ['r','r']
			return 74
		}
		return NoState
	},
	// S62
	func(r rune) int {
		switch {
		case r == 116: // ['t','t']
			return 75
		}
		return NoState
	},
	// S63
	func(r rune) int {
		switch {
		}
		return NoState
	},
	// S64
	func(r rune) int {
		switch {
		case r == 111: // ['o','o']
			return 76
		}
		return NoState
	},
	// S65
	func(r rune) int {
		switch {
		case r == 115: // ['s','s']
			return 77
		}
		return NoState
	},
	// S66
	func(r rune) int {
		switch {
		case r == 99: // ['c','c']
			return 78
		}
		return NoState
	},
	// S67
	func(r rune) int {
		switch {
		case r == 110: // ['n','n']
			return 79
		}
		return NoState
	},
	// S68
	func(r rune) int {
		switch {
		case r == 48: // ['0','0']
			return 68
		case 49 <= r && r <= 57: // ['1','9']
			return 80
		case 65 <= r && r <= 90: // ['A','Z']
			return 70
		case 97 <= r && r <= 122: // ['a','z']
			return 70
		}
		return NoState
	},
	// S69
	func(r rune) int {
		switch {
		case 48 <= r && r <= 57: // ['0','9']
			return 69
		case 65 <= r && r <= 90: // ['A','Z']
			return 70
		case 97 <= r && r <= 122: // ['a','z']
			return 70
		}
		return NoState
	},
	// S70
	func(r rune) int {
		switch {
		case r == 48: // ['0','0']
			return 68
		case 49 <= r && r <= 57: // ['1','9']
			return 80
		case 65 <= r && r <= 90: // ['A','Z']
			return 70
		case 97 <= r && r <= 122: // ['a','z']
			return 70
		}
		return NoState
	},
	// S71
	func(r rune) int {
		switch {
		case r == 101: // ['e','e']
			return 81
		}
		return NoState
	},
	// S72
	func(r rune) int {
		switch {
		case r == 111: // ['o','o']
			return 82
		}
		return NoState
	},
	// S73
	func(r rune) int {
		switch {
		case r == 48: // ['0','0']
			return 83
		case 49 <= r && r <= 57: // ['1','9']
			return 84
		case 65 <= r && r <= 90: // ['A','Z']
			return 85
		case 97 <= r && r <= 122: // ['a','z']
			return 85
		}
		return NoState
	},
	// S74
	func(r rune) int {
		switch {
		case r == 97: // ['a','a']
			return 86
		}
		return NoState
	},
	// S75
	func(r rune) int {
		switch {
		}
		return NoState
	},
	// S76
	func(r rune) int {
		switch {
		}
		return NoState
	},
	// S77
	func(r rune) int {
		switch {
		case r == 105: // ['i','i']
			return 87
		}
		return NoState
	},
	// S78
	func(r rune) int {
		switch {
		case r == 105: // ['i','i']
			return 88
		}
		return NoState
	},
	// S79
	func(r rune) int {
		switch {
		}
		return NoState
	},
	// S80
	func(r rune) int {
		switch {
		case 48 <= r && r <= 57: // ['0','9']
			return 80
		case 65 <= r && r <= 90: // ['A','Z']
			return 70
		case 97 <= r && r <= 122: // ['a','z']
			return 70
		}
		return NoState
	},
	// S81
	func(r rune) int {
		switch {
		case r == 108: // ['l','l']
			return 89
		}
		return NoState
	},
	// S82
	func(r rune) int {
		switch {
		case r == 108: // ['l','l']
			return 90
		}
		return NoState
	},
	// S83
	func(r rune) int {
		switch {
		case r == 48: // ['0','0']
			return 83
		case 49 <= r && r <= 57: // ['1','9']
			return 91
		case 65 <= r && r <= 90: // ['A','Z']
			return 85
		case 97 <= r && r <= 122: // ['a','z']
			return 85
		}
		return NoState
	},
	// S84
	func(r rune) int {
		switch {
		case 48 <= r && r <= 57: // ['0','9']
			return 84
		case 65 <= r && r <= 90: // ['A','Z']
			return 85
		case 97 <= r && r <= 122: // ['a','z']
			return 85
		}
		return NoState
	},
	// S85
	func(r rune) int {
		switch {
		case r == 48: // ['0','0']
			return 83
		case 49 <= r && r <= 57: // ['1','9']
			return 91
		case 65 <= r && r <= 90: // ['A','Z']
			return 85
		case 97 <= r && r <= 122: // ['a','z']
			return 85
		}
		return NoState
	},
	// S86
	func(r rune) int {
		switch {
		case r == 109: // ['m','m']
			return 92
		}
		return NoState
	},
	// S87
	func(r rune) int {
		switch {
		case r == 103: // ['g','g']
			return 93
		}
		return NoState
	},
	// S88
	func(r rune) int {
		switch {
		case r == 116: // ['t','t']
			return 94
		}
		return NoState
	},
	// S89
	func(r rune) int {
		switch {
		}
		return NoState
	},
	// S90
	func(r rune) int {
		switch {
		}
		return NoState
	},
	// S91
	func(r rune) int {
		switch {
		case 48 <= r && r <= 57: // ['0','9']
			return 91
		case 65 <= r && r <= 90: // ['A','Z']
			return 85
		case 97 <= r && r <= 122: // ['a','z']
			return 85
		}
		return NoState
	},
	// S92
	func(r rune) int {
		switch {
		}
		return NoState
	},
	// S93
	func(r rune) int {
		switch {
		}
		return NoState
	},
	// S94
	func(r rune) int {
		switch {
		case r == 121: // ['y','y']
			return 95
		}
		return NoState
	},
	// S95
	func(r rune) int {
		switch {
		}
		return NoState
	},
}
