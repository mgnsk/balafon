// J.S. Bach - Musikalisches Opfer - 6. Canon A 2 Per Tonos

// C3
:assign C 48
:assign D 50
:assign E 52
:assign F 53
:assign G 55
:assign A 57
:assign B 59

// C4 (middle C)
:assign c 60
:assign d 62
:assign e 64
:assign f 65
:assign g 67
:assign a 69
:assign b 71

:velocity 100

:bar bar1
	:timesig 4 4
	c.            d8 [e$ e f f#]8
	[-CE$G]16 c2          [B$A]8
:end

// 16th rests instead of ties (unimplemented).
:bar bar2
	:timesig 4 4
	g2                  a$      [-fd$c]16
	[-GB$d]16  g2               [f e]8
	B$        [-EDE]16 [FCFG]16  A$
:end

:tempo 73

:play bar1
:play bar2
