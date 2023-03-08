# Balafon

> [Balafon](https://en.wikipedia.org/wiki/Balafon) - From Wikipedia, the free encyclopedia
>
> The balafon is a gourd-resonated xylophone, a type of struck idiophone. It is closely associated with the neighbouring Mandé, Senoufo and Gur peoples of West Africa, particularly the Guinean branch of the Mandinka ethnic group, but is now found across West Africa from Guinea to Mali. Its common name, balafon, is likely a European coinage combining its Mandinka name ߓߟߊ bala with the word ߝߐ߲ fôn 'to speak' or the Greek root phono.

## Introduction

balafon is a small multitrack MIDI control language. It consists of a shell with live mode, player and a linter.

## Install

To install the `balafon` command from source, `go` and `rtmidi` are required.
Not tested on platforms other than Linux.

```sh
go install github.com/mgnsk/balafon/cmd/balafon@latest # Requires rtmidi development package.
```

## Running

- The default command lists the available MIDI ports. The default port is the 0 port.

```sh
$ balafon
0: Midi Through:Midi Through Port-0 14:0
1: Hydrogen:Hydrogen Midi-In 135:0
2: VMPK Input:in 128:0
```

- Play a file through a specific port. The port name must contain the passed in flag value:

```sh
balafon play --port "VMPK" examples/bach
```

To use piped input, pass `-` as the argument:

```sh
cat examples/bach | balafon play --port "VMPK" -
```

- Port can also be specified by its number:

```sh
balafon play --port 2 examples/bonham
```

- Enter a shell on the default port:

```sh
balafon shell
```

A shell is a text shell for the balafon language. It is capable of multiline input when entering bars.

- Enter a shell on a specific port:

```sh
balafon shell --port "Hydrogen"
```

- Load a file and enter a shell:

```sh
balafon load --port "Hydrogen" examples/bonham
```

- Enter live mode by entering the `live` command (TODO):

```sh
balafon live --port "Hydrogen" examples/live_drumset
```

Live mode is an unbuffered input mode in the shell. Whenever an assigned key is pressed,
a note on message is sent to the port.

- Lint a file:

```sh
balafon lint examples/bonham
```

- Help.

```sh
$ balafon --help
balafon is a MIDI control language and interpreter.

Usage:
   [flags]
   [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  lint        Lint a file
  live        Load a file and continue in a live shell
  load        Load a file and continue in a balafon shell
  play        Play a file
  shell       Run a balafon shell

Flags:
  -h, --help   help for this command

Use " [command] --help" for more information about a command.
```

## Syntax

The language consists of commands and note lists.
The commands tempo, program, control, start and stop are global commands.
The commands timesig, velocity and channel, when used in a bar, are scoped to that bar only.

- ### Comments

  ```
  // This is a line comment.
  ```

- ### Commands

Commands begin with a `:`.

```
// Assign a note.
:assign c 60

// Start message.
:start

// Stop message.
:stop

// Set time signature.
:timesig 4 4

// Set tempo.
:tempo 120

// Set channel.
:channel 10

// Set velocity.
:velocity 127

// Program change message on the current channel.
:program 0

// Control change message on the current channel.
:control 1 127
```

- ### Note assignment

  Assign a MIDI note number to a note letter.

  ```
  // Kick drum (on the drum channel).
  :assign k 36
  // Middle C (on other channels).
  :assign c 60
  ```

- ### Notes

  Notes are written as a letter symbol (must be assigned first) plus properties.
  The available properties are

  - sharp (`#`)
  - flat (`$`)
  - accentuated (`^`)
  - ghost (`)`)
  - numeric note value (`1`, `2`, `4`, `8` and so on)
  - dot (`.`)
  - tuplet (`/3`) (The number in the tuplet specifies the divison, for example a quintuplet `/5`)
  - let ring (`*`)

- ### Note values

  ```
  // Whole note.
  x1
  // Half note.
  x2
  // Quarter note (same as x4).
  x
  // 8th note.
  x8
  // 16th note.
  x16
  // 32th note.
  x32
  // And so on...
  ```

- ### Rests

  ```
  // A quarter rest.
  -
  // An 8th rest.
  -8
  ```

- ### Dotted notes and tuplets

  ```
  // Dotted quarter note.
  x.
  // Double-dotted note.
  x..
  // Triple-dotted note.
  x...
  // Dotted 8th note.
  x8.
  // Quarter triplet note.
  x/3
  // Dotted 8th quintuplet note.
  x8./5
  ```

- ### Flat and sharp notes

  ```
  // A note.
  c
  // A sharp note (MIDI note number + 1).
  c#
  // A flat note (MIDI note number - 1).
  c$
  ```

- ### Note grouping

  Notes can be arbitrarily grouped and properties applied to multiple notes at once.

  ```
  // Ti-Tiri.
  x8 x16 x16
  // Can be written as:
  x8[xx]16

  // Three 8th triplet notes.
  [xxx]8/3
  // Expands to
  x8/3 x8/3 x8/3

  // Nested groups are also supported:
  [[fcg] [fcg]#]8
  // Expands to
  f8 c8 g8 f#8 c#8 g#8
  ```

- ### Bars

  Bars are used to specify multiple tracks playing at once.
  Only `timesig`, `velocity` and `channel` are scoped to the bar.
  The bar is executed with the `play` command.

  ```
  // Define a bar.
  :bar "Rock beat"
  // Setting timesig makes the interpreter validate the bar length.
  // Incomplete bars are filled with silence.
  :timesig 4 4
      [xx xx xx xx]8
      // Using braces for nice alignment.
      [k  s  k  s]
  :end

  // You can also write the same bar as:
  :bar "The same beat"
      [xxxxxxxx]8
      ksks
  :end

  // Play the bar.
  :play "Rock beat"
  ```

## Examples

### The Bonham Half Time Shuffle

The file is included in the `examples` directory. To play into the default port, run

```sh
balafon play examples/bonham
```

```
// A simplified Bonham half time shuffle

// Percussion channel.
:channel 10

// Kick drum.
:assign k 36
// Acoustic snare drum.
:assign s 38
// Hi-Hat closed.
:assign x 42
// Hi-Hat open.
:assign o 46
// Hi-Hat foot.
:assign X 44
// Crash cymbal.
:assign c 49
// Low tom.
:assign q 45
// Floor tom 2.
:assign g 41

:tempo 132
:timesig 4 4
:velocity 100

// Start the first :bar with a crash cymbal and let it ring.
:bar "bonham 1"
	[[c*-o]   [x^-x]    [x^-x] [x^-x]]8/3
	-         [-s)-]8/3 s^     [-s)-]8/3
	[k-k]^8/3 [--k]8/3  -      [--k]8/3
	-         X         -2
:end

:bar "bonham 2"
	[[x^-o]   [x^-x]    [x^-x] [x^-x]]8/3
	-         [-s)-]8/3 s^     [-s)-]8/3
	[k-k]^8/3 [--k]8/3  -      [--k]8/3
	-         X         -2
:end

:bar "fill"
	[[x^-s] [sss] [ssq] [qgg]]8/3
	[[k-k]^ [--k]]8/3   -2
	-       X     X     X
:end

// Count in.
xxxo

// Play 8 :bars of the Bonham groove.
:play "bonham 1"
:play "bonham 2"
:play "bonham 2"
:play "fill"
:play "bonham 1"
:play "bonham 2"
:play "bonham 2"
:play "fill"
```

### J.S. Bach - Musikalisches Opfer - 6. Canon A 2 Per Tonos

The file is included in the `examples` directory. To play into the default port, run

```sh
balafon play examples/bach
```

It is possible to write melodies using balafon in a limited way. Here's 2 bars of Bach:

```
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

:bar "bar 1"
	:timesig 4 4
	c.            d8 [e$ e f f#]8
	[-CE$G]16 c2          [B$A]8
:end

// 16th rests instead of ties (unimplemented).
:bar "bar 2"
	:timesig 4 4
	g2                  a$      [-fd$c]16
	[-GB$d]16  g2               [f e]8
	B$        [-EDE]16 [FCFG]16  A$
:end

:tempo 73

:play "bar 1"
:play "bar 2"
```

### Multichannel

The file is included in the `examples` directory.

```
:channel 1
:assign x 42
:channel 2
:assign k 36
:tempo 60
:timesig 4 4
:bar "test"
	:channel 1
	xxxx
	:channel 2
	kkkk
:end
:play "test"
```

## Possible features in the future

- Tie (a curved line connecting the heads of two notes of the same pitch) - no idea about the syntax. Can be partially emulated by using dotted notes if the rhythm is simple enough.
- WebAssembly support with Web MIDI for running in browsers.
- Accelerando/Ritardando.
