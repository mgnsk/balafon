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
- Compile a YAML file to SMF:
  ```sh
  $ gong compile examples/example.yml | gong smf -o example.mid -
  ```

- Help.

  ```sh
  gong is a MIDI control language and interpreter.

  Usage:
     [flags]
     [command]

  Available Commands:
    compile     Compile a YAML file to gong script
    help        Help about any command
    lint        Lint a file
    list        List available MIDI output ports
    load        Load a file and continue in a gong shell
    play        Play a file
    smf         Compile a gong file to SMF

  Flags:
    -h, --help          help for this command
        --port string   MIDI output port (default "0")

  Use " [command] --help" for more information about a command.
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
// A simplified Bonham half time shuffle

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

velocity 100

// Start the first bar with a crash cymbal and let it ring.
bar "bonham 1"
timesig 4 4
[[c*-o]   [x^-x]    [x^-x] [x^-x]]8/3
-         [-s)-]8/3 s      [-s)-]8/3
[k^-k]8/3 [--k]8/3  -      [--k]8/3
-         X         -2
end

bar "bonham 2"
timesig 4 4
[[x^-o]   [x^-x]    [x^-x] [x^-x]]8/3
-         [-s)-]8/3 s      [-s)-]8/3
[k^-k]8/3 [--k]8/3  -      [--k]8/3
-         X         -2
end

bar "fill"
timesig 4 4
[[x^-s] [sss] [ssq] [qgg]]8/3
[[k-k]  [--k]]8/3   -2
-       X     X     X
end

tempo 132

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

velocity 100

bar "bar 1"
timesig 4 4
c.            d8 [e$ e f f#]8
[-CE$G]16 c2          [B$A]8
end

// 16th rests instead of ties (unimplemented).
bar "bar 2"
timesig 4 4
g2                  a$      [-fd$c]16
[-GB$d]16  g2               [f e]8
B$        [-EDE]16 [FCFG]16  A$
end

tempo 73

play "bar 1"
play "bar 2"
```

### Multichannel

The file is included in the `examples` directory.

```
channel 10
program 1
// Kick drum.
assign k 36
// Acoustic snare drum.
assign s 38
// Hi-Hat closed.
assign x 42

channel 2
program 2
assign C 48
assign c 60
assign e 64
assign g 67

bar "bar 1"
	timesig 4 4

	channel 1
	control 1 1
	xxxx
	ksks

	channel 2
	control 2 2
	cegc
	C1
end

play "setup tracks"
play "bar 1"
```

### YAML example

The file is included in the `examples` directory.

```yaml
---
instruments:
  lead:
    channel: 1
    assign:
      c: 60
      d: 62

  bass:
    channel: 2
    assign:
      c: 48
      d: 50

  drums:
    channel: 10
    assign:
      k: 36
      s: 38

bars:
  - name: sound A
    params:
      lead:
        program: 1
      bass:
        program: 1
      drums:
        program: 127

  - name: lead reverb on
    params:
      lead:
        control: 100
        parameter: 100

  - name: lead reverb off
    params:
      lead:
        control: 100
        parameter: 0

  - name: tempo 2
    tempo: 200

  - name: Verse
    time: 4
    sig: 4
    tracks:
      bass:
        - "[cd]2"
      lead:
        - ccdd
        - "[cd]2"
      drums:
        - ksks

  - name: Fill
    time: 3
    sig: 8
    tracks:
      drums:
        - "[ksk]8"

play:
  - sound A
  - lead reverb on
  - Verse
  - lead reverb off
  - tempo 2
  - Fill
  - Verse
```

## Possible features in the future

- Tie (a curved line connecting the heads of two notes of the same pitch) - no idea about the syntax. Can be partially emulated by using dotted notes if the rhythm is simple enough.
- WebAssembly support with Web MIDI for running in browsers.
- Accelerando/Ritardando.
