## Introduction

gong is a small domain-specific language for controlling MIDI devices.
It includes a live interpreter and can play back standalone text files.

## Install

To install gong from source, `go` and `rtmidi` are required.
Not tested on platforms other than Linux.

```sh
go install github.com/mgnsk/gong@latest
```

## Running

- List the available MIDI ports. The default port is the first port in the list.
  ```sh
  $ gong list
  0: Midi Through:Midi Through Port-0 14:0
  1: VMPK Input:in 128:0
  2: Hydrogen:Hydrogen Midi-In 135:0
  ```
- Play a file through a specific port. The port name must contain the passed in flag value:
  ```sh
  $ gong play --port "VM" examples/piano
  ```
- Port can also be specified by its number:
  ```sh
  $ gong play --port 2 examples/bonham
  ```
- Enter the live shell on the default port:
  ```sh
  $ gong
  Welcome to the gong shell on MIDI port '0: Midi Through:Midi Through Port-0 14:0'!
  >
  ```
- Enter a live shell on a specific port:
  ```sh
  $ gong --port "VM"
  Welcome to the gong shell on MIDI port '1: VMPK Input:in 128:0'!
  >
  ```
- Help.

  ```sh
  gong is a MIDI control language and interpreter.

  Usage:
     [flags]
     [command]

  Available Commands:
    help        Help about any command
    list        List available MIDI output ports
    play        Play a file

  Flags:
    -h, --help          help for this command
        --port string   MIDI output port (default "0")

  Use " [command] --help" for more information about a command.
  ```

## Syntax

- #### Comments
  ```
  // This is a line comment.
  ```
- #### Commands
  ```
  // Assign a note.
  assign c 60
  // Set the tempo.
  tempo 120
  // Set the MIDI channel.
  channel 10
  // Set velocity.
  velocity 127
  // Program change message.
  program 0
  // Control change message.
  control 1 127
  // Start message. Useful for controlling a DAW which records MIDI input.
  start
  // Stop message.
  stop
  ```
- #### Note assignment
  Assign a MIDI note number to a note letter.
  ```
  // Kick drum (on the drum channel).
  assign k 36
  // Middle C (on other channels).
  assign c 60
  ```
- #### Notes
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
- #### Note values
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
- #### Dotted notes and tuplets
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
- #### Flat and sharp notes.
  ```
  // A note.
  c
  // A sharp note (MIDI note number + 1).
  c#
  // A flat note (MIDI note number - 1).
  c$
  ```
- #### Note grouping
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
  [[cde] [cde]#]8
  // Expands to
  c8 d8 e8 c#8 d#8 e#8
  ```
- #### Bars

  Bars are used to to specify multiple tracks playing at once.

  ```
  // Define a bar.
  bar "Rock beat"
  [xx xx xx xx]8
  // Using braces for nice alignment.
  [k  s  k  s]
  end

  // You can also write the same bar as:
  bar "The same beat"
  [xxxxxxxx]8
  ksks
  end

  // Play the bar.
  play "Rock beat"
  ```

## Examples

### The Bonham Half Time Shuffle

The file is included in the `examples` directory. To play into the default port, run

```sh
$ gong play examples/bonham
```

```
// A simplified Bonham half time shuffle

tempo 132
velocity 100

// Percussion channel.
channel 10

// Kick drum.
assign k 36
// Acoustic snare drum.
assign s 38
// Hi-Hat closed.
assign x 42
// Hi-Hat open.
assign o 46
// Hi-Hat foot.
assign X 44
// Crash cymbal.
assign c 49
// Low tom.
assign q 45
// Floor tom 2.
assign g 41

// Start the first bar with a crash cymbal and let it ring.
bar "bonham 1"
[[c*-o]    [x^-x]    [x^-x] [x^-x]]8/3
-          [-s)-]8/3 s      [-s)-]8/3
[k^-k]8/3  [--k]8/3  -      [--k]8/3
-          X
end

bar "bonham 2"
[[x-o]    [x^-x]    [x^-x] [x^-x]]8/3
-         [-s)-]8/3 s      [-s)-]8/3
[k^-k]8/3 [--k]8/3  -      [--k]8/3
-         X
end

bar "fill"
[[--s] [sss] [ssq] [qgg]]8/3
[[k-k] [--k]]8/3
x      X     X     X
end

// Count in.
xxxo

// Play 8 bars of the Bonham groove.
play "bonham 1"
play "bonham 2"
play "bonham 2"
play "fill"
play "bonham 1"
play "bonham 2"
play "bonham 2"
play "fill"
```

### J.S. Bach - Musikalisches Opfer - 6. Canon A 2 Per Tonos

The file is included in the `examples` directory. To play into the default port, run

```sh
$ gong play examples/bach
```

It is possible to write melodies using gong in a limited way. Here's 2 bars of Bach:

```
// J.S. Bach - Musikalisches Opfer - 6. Canon A 2 Per Tonos

// C3
assign C 48
assign D 50
assign E 52
assign F 53
assign G 55
assign A 57
assign B 59

// C4 (middle C)
assign c 60
assign d 62
assign e 64
assign f 65
assign g 67
assign a 69
assign b 71

tempo 73
velocity 100

bar "bar 1"
c.            d8 [e$ e f f#]8
[-CE$G]16 c2          [B$A]8
end

// 16th rests instead of ties (unimplemented).
bar "bar 2"
g2                  a$      [- f d$ c]16
[-GB$d]16  g2               [f   e]8
B$        [-EDE]16 [FCFG]16  A$
end

play "bar 1"
play "bar 2"
```

## Possible features in the future

- Tie (a curved line connecting the heads of two notes of the same pitch) - no idea about the syntax. Can be partially emulated by using dotted notes if the rhythm is simple enough.
- WebAssembly support with Web MIDI for running in browsers.
- Generating an SMF midi file.
- Accelerando/Ritardando.
