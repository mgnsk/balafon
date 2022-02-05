## Introduction

gong is a multitrack MIDI control language. It consists of a shell with live mode,
an SMF compiler, and a playback engine.

There exists a strict YAML specification that compiles down to gong script.

## Install

To install the `gong` command from source, `go` and `rtmidi` are required.
Not tested on platforms other than Linux.

```sh
go install github.com/mgnsk/gong/cmd/gong@latest # Requires rtmidi development package.
go install github.com/mgnsk/gong/cmd/gong2smf@latest
go install github.com/mgnsk/gong/cmd/gonglint@latest
go install github.com/mgnsk/gong/cmd/yaml2gong@latest
```

## Running

- List the available MIDI ports. The default port is the first port in the list.
  ```sh
  $ gong list
  0: Midi Through:Midi Through Port-0 14:0
  1: Hydrogen:Hydrogen Midi-In 135:0
  2: VMPK Input:in 128:0
  ```
- Play a file through a specific port. The port name must contain the passed in flag value:
  ```sh
  $ gong play --port "VMPK" examples/bach
  ```
  To use piped input, pass `-` as the argument:
  ```sh
  $ cat examples/bach | gong play --port "VMPK" -
  ```
- Port can also be specified by its number:
  ```sh
  $ gong play --port 2 examples/bonham
  ```
- Enter a shell on the default port:
  ```sh
  $ gong
  Welcome to the gong shell on MIDI port '0: Midi Through:Midi Through Port-0 14:0'!
  >
  ```
  A shell is a line-based shell for the gong language.
- Enter a shell on a specific port:
  ```sh
  $ gong --port "Hydrogen"
  Welcome to the gong shell on MIDI port '1: Hydrogen:Hydrogen Midi-In 128:0'!
  >
  ```
- Load a file and enter a shell:
  ```sh
  $ gong --port "Hydrogen" load examples/bonham
  Welcome to the gong shell on MIDI port '1: Hydrogen:Hydrogen Midi-In 128:0'!
  >
  ```
- Enter live mode:
  ```sh
  $ gong --port "Hydrogen" load examples/bonham
  Welcome to the gong shell on MIDI port '1: Hydrogen:Hydrogen Midi-In 128:0'!
  > live
  Entered live mode. Press Ctrl+D to exit.
  ```
  Live mode is an unbuffered input mode in the shell. Whenever an assigned key is pressed,
  the corresponding MIDI note on event is immediately sent to the port. In this mode, all notes
  are left ringing and a note off event is sent only when a key is pressed more than once.
- Lint a file:
  ```sh
  $ gonglint examples/bonham
  ```
- Compile to SMF:
  ```sh
  $ gong2smf -o examples/bonham.mid examples/bonham
  ```
  Piping is also supported:
  ```sh
  $ cat examples/bach | gong2smf -o examples/bach.mid -
  ```
- Compile a YAML file to gong script and play it:
  ```sh
  $ yaml2gong examples/example.yml | gong play -
  ```
- Compile a YAML file to SMF:
  ```sh
  $ yaml2gong examples/example.yml | gong2smf -o example.mid -
  ```

- Help.

  ```sh
{{.HelpSection | indent 2 | trim_trailing_newlines}}
  ```

## Syntax

The language consists of commands and note lists. It is possible to group commands and notes in bars.

- #### Comments
  ```
  // This is a line comment.
  ```
- #### Commands
  ```
  // Assign a note.
  assign c 60

  // Start message. Useful for controlling a DAW which records MIDI input.
  start

  // Stop message.
  stop

  // Set the time signature.
  // Optional and applicable only as the first command in a bar.
  timesig 4 4

  // Set the current global tempo.
  tempo 120

  // Set the current global MIDI channel.
  channel 10

  // Set current global velocity.
  velocity 127

  // Program change message on the current channel.
  program 0

  // Control change message on the current channel.
  control 1 127
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
  [[fcg] [fcg]#]8
  // Expands to
  f8 c8 g8 f#8 c#8 g#8
  ```
- #### Bars

  Bars are used to specify multiple tracks playing at once.
  Commands used inside bars are not scoped and have global state.
  For example setting a channel, it becomes the default for all following messages.
  In multi-channel files, each bar must specify the its channel.
  See a multi-channel example at the end of this document.

  ```
  // Define a bar.
  bar "Rock beat"
  // Setting timesig makes the interpreter validate the bar length.
  timesig 4 4
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

### Multichannel

The file is included in the `examples` directory.

```
{{.MultiExample | trim_trailing_newlines}}
```

### YAML example

The gong language has a strict YAML wrapper that compiles to valid gong script.

The file is included in the `examples` directory.

```yaml
{{.YAMLExample | trim_trailing_newlines}}
```

## Possible features in the future

- Tie (a curved line connecting the heads of two notes of the same pitch) - no idea about the syntax. Can be partially emulated by using dotted notes if the rhythm is simple enough.
- WebAssembly support with Web MIDI for running in browsers.
- Accelerando/Ritardando.
