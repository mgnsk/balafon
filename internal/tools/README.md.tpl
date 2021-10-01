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
  $ gong list-ports
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
{{.HelpSection | indent 2 | trim_trailing_newlines}}
  ```

## Syntax

- #### Comments
  ```
  // This is a line comment.
  ```
- #### Commands
  ```
  tempo 120
  channel 10
  velocity 127
  program 0
  control 1 127
  ```
- #### Note assignment
  Assign a MIDI note number to a note letter.
  ```
  // Kick drum (on the drum channel).
  k=36
  // Middle C (on other channels).
  c=60
  ```
- #### Notes
  Notes are written as a letter symbol (must be assigned first) plus properties.
  The available properties are
  - sharp (`#`)
  - flat (`$`)
  - numeric note value (`1`, `2`, `4`, `8` and so on)
  - dot (`.`)
  - tuplet (`/3`)
    The number in the tuplet specifies the divison, for example a quintuplet `/5`.
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
  ```
  // Ti-Tiri.
  x8 x16 x16
  // Can be written as:
  x8xx16
  // Three 8th triplet notes.
  xxx8/3
  ```
- #### Bars

  Bars are used to to specify multiple tracks playing at once.

  ```
  // Define a bar.
  bar "Rock beat"
  xx8 xx8 xx8 xx8
  k   s   k   s
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
{{.BonhamExample | trim_trailing_newlines}}
```

### J.S. Bach - Musikalisches Opfer - 6. Canon A 2 Per Tonos

The file is included in the `examples` directory. To play into the default port, run

```sh
$ gong play examples/bach
```

It is possible to write melodies using gong in a limited way. Here's 2 bars of Bach:

```
{{.BachExample | trim_trailing_newlines}}
```

## Possible features in the future

- Tie (a curved line connecting the heads of two notes of the same pitch) - no idea about the syntax. Can be partially emulated by using dotted notes if the rhythm is simple enough.
- Ghost note property - gonna have to think about the syntax. Probably `x)`.
- Accentuated note property - probably `x^`.
- WebAssembly support with Web MIDI for running in browsers.
- Generating an SMF midi file.
- Accelerando/Ritardando.
