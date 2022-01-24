## Introduction

gong is a small low-level domain-specific language for controlling MIDI devices.
It includes a live interpreter and can play back standalone text files.

There also exists a high-level YAML specification that compiles down to gong script.

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
  Piped input is accepted:
  ```sh
  $ cat examples/piano | gong play --port "VM" -
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
- Load a file and enter a live shell:
  ```sh
  $ gong load examples/bonham
  >
  ```
- Lint a file:
  ```sh
  $ gong lint examples/bonham
  ```
- Compile to SMF:
  ```sh
  $ gong smf -o examples/bonham.mid examples/bonham
  $ cat examples/bach | gong smf -o examples/bach.mid -
  ```
- Compile a YAML file to gong script and play it:
  ```sh
  $ gong compile examples/example.yml | gong play -
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
  // Assign a note.
  assign c 60
  //
  // Start message. Useful for controlling a DAW which records MIDI input.
  start
  //
  // Stop message.
  stop
  //
  // Set the time signature.
  // Optional and applicable only as the first command in a bar.
  timesig 4 4
  //
  // The following commands, when used inside a bar,
  // apply to the beginning of the bar regardless of position.
  //
  // Set the current global tempo.
  tempo 120
  //
  // Set the current global MIDI channel.
  channel 10
  //
  // Set current global velocity.
  velocity 127
  //
  // Program change message on the current channel.
  program 0
  //
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

The file is included in the `examples` directory.

```
{{.YAMLExample | trim_trailing_newlines}}
```

## Possible features in the future

- Tie (a curved line connecting the heads of two notes of the same pitch) - no idea about the syntax. Can be partially emulated by using dotted notes if the rhythm is simple enough.
- WebAssembly support with Web MIDI for running in browsers.
- Accelerando/Ritardando.
