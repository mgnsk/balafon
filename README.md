## Introduction

gong is a small domain-specific language for controlling live MIDI devices.
It includes a live interpreter with autocompletion and can play back standalone text files.

## Install

To install gong from source, `go` and `rtmidi` are required.
Not tested on platforms other than Linux.

```sh
go install github.com/mgnsk/gong@latest
```

## Running

- List the available MIDI ports. The default port is the first port in the list.
  ```sh
  > gong list-ports
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
  $ gong --help
  gong is a MIDI control language interpreter.

  Usage:
     [flags]
     [command]

  Available Commands:
    help        Help about any command
    list-ports  List available MIDI output ports
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
- #### Dotted notes and triplets
  ```
  // Dotted quarter note.
  x.
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
// The Bonham half time shuffle

tempo 132
velocity 100

// Percussion channel.
channel 10

// Kick drum.
k=36
// Acoustic snare drum.
s=38
// Hi-Hat closed.
x=42
// Hi-Hat open.
o=46
// Hi-Hat foot.
X=44
// Crash cymbal.
c=49
// Low tom.
q=45
// Floor tom 2.
g=41

// Start the first bar with a crash cymbal.
bar "bonham 1"
c1
--o8/3 x-x8/3 x-x8/3 x-x8/3
k-k8/3 -sk8/3 s      -sk8/3
-      X
end

bar "bonham 2"
x-o8/3 x-x8/3 x-x8/3 x-x8/3
k-k8/3 -sk8/3 s      -sk8/3
-      X
end

bar "fill"
--s8/3 sss8/3 ssq8/3 qgg8/3
k-k8/3 --k8/3
x      X      X      X
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
C=48
D=50
E=52
F=53
G=55
A=57
B=59

// C4 (middle C)
c=60
d=62
e=64
f=65
g=67
a=69
b=71

tempo 73
velocity 100

bar "bar 1"
c.                d8 e$8 e8 f8 f#8
-C16 E$16 G16 c2            B$8A8
end

// 16th rests instead of ties (unimplemented).
bar "bar 2"
g2                 a$     -f16d$16c16
-G16B$16d16 g2            fe8
B$          -EDE16 FCFG16 A$
end

play "bar 1"
play "bar 2"
```

## Possible features in the future

- Tie (a curved line connecting the heads of two notes of the same pitch) - no idea about the syntax. Can be partially emulated by using note values if the rhythm is simple enough.
- Double dotted notes.
- Ghost note property - gonna have to think about the syntax. Probably `x)`.
- Accentuated note property - probably `x^`.
- WebAssembly support with Web MIDI for running in browsers.
- Generating an SMF midi file.
- Accelerando/Ritardando.
