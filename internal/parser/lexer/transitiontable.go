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
			return 2
		case r == 62: // ['>','>']
			return 13
		case 65 <= r && r <= 90: // ['A','Z']
			return 14
		case r == 91: // ['[','[']
			return 15
		case r == 93: // [']',']']
			return 16
		case r == 94: // ['^','^']
			return 17
		case r == 96: // ['`','`']
			return 18
		case 97 <= r && r <= 122: // ['a','z']
			return 14
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
		case r == 51: // ['3','3']
			return 20
		case r == 53: // ['5','5']
			return 20
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
			return 21
		case r == 98: // ['b','b']
			return 22
		case r == 99: // ['c','c']
			return 23
		case r == 101: // ['e','e']
			return 24
		case r == 107: // ['k','k']
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
		}
		return NoState
	},
	// S21
	func(r rune) int {
		switch {
		case r == 115: // ['s','s']
			return 31
		}
		return NoState
	},
	// S22
	func(r rune) int {
		switch {
		case r == 97: // ['a','a']
			return 32
		}
		return NoState
	},
	// S23
	func(r rune) int {
		switch {
		case r == 104: // ['h','h']
			return 33
		case r == 111: // ['o','o']
			return 34
		}
		return NoState
	},
	// S24
	func(r rune) int {
		switch {
		case r == 110: // ['n','n']
			return 35
		}
		return NoState
	},
	// S25
	func(r rune) int {
		switch {
		case r == 101: // ['e','e']
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
		case r == 111: // ['o','o']
			return 43
		}
		return NoState
	},
	// S30
	func(r rune) int {
		switch {
		case r == 42: // ['*','*']
			return 30
		case r == 47: // ['/','/']
			return 44
		default:
			return 19
		}
	},
	// S31
	func(r rune) int {
		switch {
		case r == 115: // ['s','s']
			return 45
		}
		return NoState
	},
	// S32
	func(r rune) int {
		switch {
		case r == 114: // ['r','r']
			return 46
		}
		return NoState
	},
	// S33
	func(r rune) int {
		switch {
		case r == 97: // ['a','a']
			return 47
		}
		return NoState
	},
	// S34
	func(r rune) int {
		switch {
		case r == 110: // ['n','n']
			return 48
		}
		return NoState
	},
	// S35
	func(r rune) int {
		switch {
		case r == 100: // ['d','d']
			return 49
		}
		return NoState
	},
	// S36
	func(r rune) int {
		switch {
		case r == 121: // ['y','y']
			return 50
		}
		return NoState
	},
	// S37
	func(r rune) int {
		switch {
		case r == 97: // ['a','a']
			return 51
		}
		return NoState
	},
	// S38
	func(r rune) int {
		switch {
		case r == 111: // ['o','o']
			return 52
		}
		return NoState
	},
	// S39
	func(r rune) int {
		switch {
		case r == 97: // ['a','a']
			return 53
		case r == 111: // ['o','o']
			return 54
		}
		return NoState
	},
	// S40
	func(r rune) int {
		switch {
		case r == 109: // ['m','m']
			return 55
		}
		return NoState
	},
	// S41
	func(r rune) int {
		switch {
		case r == 109: // ['m','m']
			return 56
		}
		return NoState
	},
	// S42
	func(r rune) int {
		switch {
		case r == 108: // ['l','l']
			return 57
		}
		return NoState
	},
	// S43
	func(r rune) int {
		switch {
		case r == 105: // ['i','i']
			return 58
		}
		return NoState
	},
	// S44
	func(r rune) int {
		switch {
		}
		return NoState
	},
	// S45
	func(r rune) int {
		switch {
		case r == 105: // ['i','i']
			return 59
		}
		return NoState
	},
	// S46
	func(r rune) int {
		switch {
		case r == 9: // ['\t','\t']
			return 60
		case r == 32: // [' ',' ']
			return 60
		}
		return NoState
	},
	// S47
	func(r rune) int {
		switch {
		case r == 110: // ['n','n']
			return 61
		}
		return NoState
	},
	// S48
	func(r rune) int {
		switch {
		case r == 116: // ['t','t']
			return 62
		}
		return NoState
	},
	// S49
	func(r rune) int {
		switch {
		}
		return NoState
	},
	// S50
	func(r rune) int {
		switch {
		case r == 9: // ['\t','\t']
			return 63
		case r == 32: // [' ',' ']
			return 63
		}
		return NoState
	},
	// S51
	func(r rune) int {
		switch {
		case r == 121: // ['y','y']
			return 64
		}
		return NoState
	},
	// S52
	func(r rune) int {
		switch {
		case r == 103: // ['g','g']
			return 65
		}
		return NoState
	},
	// S53
	func(r rune) int {
		switch {
		case r == 114: // ['r','r']
			return 66
		}
		return NoState
	},
	// S54
	func(r rune) int {
		switch {
		case r == 112: // ['p','p']
			return 67
		}
		return NoState
	},
	// S55
	func(r rune) int {
		switch {
		case r == 112: // ['p','p']
			return 68
		}
		return NoState
	},
	// S56
	func(r rune) int {
		switch {
		case r == 101: // ['e','e']
			return 69
		}
		return NoState
	},
	// S57
	func(r rune) int {
		switch {
		case r == 111: // ['o','o']
			return 70
		}
		return NoState
	},
	// S58
	func(r rune) int {
		switch {
		case r == 99: // ['c','c']
			return 71
		}
		return NoState
	},
	// S59
	func(r rune) int {
		switch {
		case r == 103: // ['g','g']
			return 72
		}
		return NoState
	},
	// S60
	func(r rune) int {
		switch {
		case r == 9: // ['\t','\t']
			return 60
		case r == 32: // [' ',' ']
			return 60
		case r == 48: // ['0','0']
			return 73
		case 49 <= r && r <= 57: // ['1','9']
			return 74
		case 65 <= r && r <= 90: // ['A','Z']
			return 75
		case 97 <= r && r <= 122: // ['a','z']
			return 75
		}
		return NoState
	},
	// S61
	func(r rune) int {
		switch {
		case r == 110: // ['n','n']
			return 76
		}
		return NoState
	},
	// S62
	func(r rune) int {
		switch {
		case r == 114: // ['r','r']
			return 77
		}
		return NoState
	},
	// S63
	func(r rune) int {
		switch {
		case r == 9: // ['\t','\t']
			return 63
		case r == 32: // [' ',' ']
			return 63
		case r == 65: // ['A','A']
			return 78
		case r == 66: // ['B','B']
			return 79
		case r == 67: // ['C','C']
			return 80
		case r == 68: // ['D','D']
			return 81
		case r == 69: // ['E','E']
			return 82
		case r == 70: // ['F','F']
			return 83
		case r == 71: // ['G','G']
			return 84
		}
		return NoState
	},
	// S64
	func(r rune) int {
		switch {
		case r == 9: // ['\t','\t']
			return 85
		case r == 32: // [' ',' ']
			return 85
		}
		return NoState
	},
	// S65
	func(r rune) int {
		switch {
		case r == 114: // ['r','r']
			return 86
		}
		return NoState
	},
	// S66
	func(r rune) int {
		switch {
		case r == 116: // ['t','t']
			return 87
		}
		return NoState
	},
	// S67
	func(r rune) int {
		switch {
		}
		return NoState
	},
	// S68
	func(r rune) int {
		switch {
		case r == 111: // ['o','o']
			return 88
		}
		return NoState
	},
	// S69
	func(r rune) int {
		switch {
		}
		return NoState
	},
	// S70
	func(r rune) int {
		switch {
		case r == 99: // ['c','c']
			return 89
		}
		return NoState
	},
	// S71
	func(r rune) int {
		switch {
		case r == 101: // ['e','e']
			return 90
		}
		return NoState
	},
	// S72
	func(r rune) int {
		switch {
		case r == 110: // ['n','n']
			return 91
		}
		return NoState
	},
	// S73
	func(r rune) int {
		switch {
		case r == 48: // ['0','0']
			return 73
		case 49 <= r && r <= 57: // ['1','9']
			return 92
		case 65 <= r && r <= 90: // ['A','Z']
			return 75
		case 97 <= r && r <= 122: // ['a','z']
			return 75
		}
		return NoState
	},
	// S74
	func(r rune) int {
		switch {
		case 48 <= r && r <= 57: // ['0','9']
			return 74
		case 65 <= r && r <= 90: // ['A','Z']
			return 75
		case 97 <= r && r <= 122: // ['a','z']
			return 75
		}
		return NoState
	},
	// S75
	func(r rune) int {
		switch {
		case r == 48: // ['0','0']
			return 73
		case 49 <= r && r <= 57: // ['1','9']
			return 92
		case 65 <= r && r <= 90: // ['A','Z']
			return 75
		case 97 <= r && r <= 122: // ['a','z']
			return 75
		}
		return NoState
	},
	// S76
	func(r rune) int {
		switch {
		case r == 101: // ['e','e']
			return 93
		}
		return NoState
	},
	// S77
	func(r rune) int {
		switch {
		case r == 111: // ['o','o']
			return 94
		}
		return NoState
	},
	// S78
	func(r rune) int {
		switch {
		case r == 98: // ['b','b']
			return 95
		case r == 109: // ['m','m']
			return 96
		}
		return NoState
	},
	// S79
	func(r rune) int {
		switch {
		case r == 98: // ['b','b']
			return 97
		case r == 109: // ['m','m']
			return 96
		}
		return NoState
	},
	// S80
	func(r rune) int {
		switch {
		case r == 35: // ['#','#']
			return 98
		case r == 109: // ['m','m']
			return 99
		}
		return NoState
	},
	// S81
	func(r rune) int {
		switch {
		case r == 35: // ['#','#']
			return 100
		case r == 98: // ['b','b']
			return 95
		case r == 109: // ['m','m']
			return 99
		}
		return NoState
	},
	// S82
	func(r rune) int {
		switch {
		case r == 98: // ['b','b']
			return 101
		case r == 109: // ['m','m']
			return 96
		}
		return NoState
	},
	// S83
	func(r rune) int {
		switch {
		case r == 35: // ['#','#']
			return 102
		case r == 109: // ['m','m']
			return 99
		}
		return NoState
	},
	// S84
	func(r rune) int {
		switch {
		case r == 35: // ['#','#']
			return 103
		case r == 98: // ['b','b']
			return 95
		case r == 109: // ['m','m']
			return 99
		}
		return NoState
	},
	// S85
	func(r rune) int {
		switch {
		case r == 9: // ['\t','\t']
			return 85
		case r == 32: // [' ',' ']
			return 85
		case r == 48: // ['0','0']
			return 104
		case 49 <= r && r <= 57: // ['1','9']
			return 105
		case 65 <= r && r <= 90: // ['A','Z']
			return 106
		case 97 <= r && r <= 122: // ['a','z']
			return 106
		}
		return NoState
	},
	// S86
	func(r rune) int {
		switch {
		case r == 97: // ['a','a']
			return 107
		}
		return NoState
	},
	// S87
	func(r rune) int {
		switch {
		}
		return NoState
	},
	// S88
	func(r rune) int {
		switch {
		}
		return NoState
	},
	// S89
	func(r rune) int {
		switch {
		case r == 105: // ['i','i']
			return 108
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
		}
		return NoState
	},
	// S92
	func(r rune) int {
		switch {
		case 48 <= r && r <= 57: // ['0','9']
			return 92
		case 65 <= r && r <= 90: // ['A','Z']
			return 75
		case 97 <= r && r <= 122: // ['a','z']
			return 75
		}
		return NoState
	},
	// S93
	func(r rune) int {
		switch {
		case r == 108: // ['l','l']
			return 109
		}
		return NoState
	},
	// S94
	func(r rune) int {
		switch {
		case r == 108: // ['l','l']
			return 110
		}
		return NoState
	},
	// S95
	func(r rune) int {
		switch {
		}
		return NoState
	},
	// S96
	func(r rune) int {
		switch {
		}
		return NoState
	},
	// S97
	func(r rune) int {
		switch {
		case r == 109: // ['m','m']
			return 99
		}
		return NoState
	},
	// S98
	func(r rune) int {
		switch {
		case r == 109: // ['m','m']
			return 96
		}
		return NoState
	},
	// S99
	func(r rune) int {
		switch {
		}
		return NoState
	},
	// S100
	func(r rune) int {
		switch {
		case r == 109: // ['m','m']
			return 96
		}
		return NoState
	},
	// S101
	func(r rune) int {
		switch {
		case r == 109: // ['m','m']
			return 99
		}
		return NoState
	},
	// S102
	func(r rune) int {
		switch {
		case r == 109: // ['m','m']
			return 96
		}
		return NoState
	},
	// S103
	func(r rune) int {
		switch {
		case r == 109: // ['m','m']
			return 96
		}
		return NoState
	},
	// S104
	func(r rune) int {
		switch {
		case r == 48: // ['0','0']
			return 104
		case 49 <= r && r <= 57: // ['1','9']
			return 111
		case 65 <= r && r <= 90: // ['A','Z']
			return 106
		case 97 <= r && r <= 122: // ['a','z']
			return 106
		}
		return NoState
	},
	// S105
	func(r rune) int {
		switch {
		case 48 <= r && r <= 57: // ['0','9']
			return 105
		case 65 <= r && r <= 90: // ['A','Z']
			return 106
		case 97 <= r && r <= 122: // ['a','z']
			return 106
		}
		return NoState
	},
	// S106
	func(r rune) int {
		switch {
		case r == 48: // ['0','0']
			return 104
		case 49 <= r && r <= 57: // ['1','9']
			return 111
		case 65 <= r && r <= 90: // ['A','Z']
			return 106
		case 97 <= r && r <= 122: // ['a','z']
			return 106
		}
		return NoState
	},
	// S107
	func(r rune) int {
		switch {
		case r == 109: // ['m','m']
			return 112
		}
		return NoState
	},
	// S108
	func(r rune) int {
		switch {
		case r == 116: // ['t','t']
			return 113
		}
		return NoState
	},
	// S109
	func(r rune) int {
		switch {
		}
		return NoState
	},
	// S110
	func(r rune) int {
		switch {
		}
		return NoState
	},
	// S111
	func(r rune) int {
		switch {
		case 48 <= r && r <= 57: // ['0','9']
			return 111
		case 65 <= r && r <= 90: // ['A','Z']
			return 106
		case 97 <= r && r <= 122: // ['a','z']
			return 106
		}
		return NoState
	},
	// S112
	func(r rune) int {
		switch {
		}
		return NoState
	},
	// S113
	func(r rune) int {
		switch {
		case r == 121: // ['y','y']
			return 114
		}
		return NoState
	},
	// S114
	func(r rune) int {
		switch {
		}
		return NoState
	},
}
