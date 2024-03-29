/* A simplified Bonham half time shuffle */

/* Percussion channel. */
:channel 10

/* Kick drum. */
:assign k 36
/* Acoustic snare drum. */
:assign s 38
/* Hi-Hat closed. */
:assign x 42
/* Hi-Hat open. */
:assign o 46
/* Hi-Hat foot. */
:assign X 44
/* Crash cymbal. */
:assign c 49
/* Low tom. */
:assign q 45
/* Floor tom 2. */
:assign g 41

:tempo 132
:time 4 4
:velocity 100

/*
Start the first bar with a crash cymbal and let it ring.
*/
:bar bonham1
	[[c*-o]   [x>-x]    [x^-x] [x>-x]]8/3
	-         [-s)-]8/3 s^     [-s)-]8/3
	[k-k]>8/3 [--k]8/3  -      [--k]8/3
	-         X         -2
:end

:bar bonham2
	[[x^-o]   [x>-x]    [x^-x] [x>-x]]8/3
	-         [-s)-]8/3 s^     [-s)-]8/3
	[k-k]>8/3 [--k]8/3  -      [--k]8/3
	-         X         -2
:end

:bar fill
	[[x^-s] [sss] [ssq] [qgg]]8/3
	[[k-k]> [--k]]8/3   -2
	-       X     X     X
:end

/*
Count in.
*/
xxxo

/* Play 8 :bars of the Bonham groove. */
:play bonham1
:play bonham2
:play bonham2
:play fill
:play bonham1
:play bonham2
:play bonham2
:play fill
