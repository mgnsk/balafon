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
  1: Hydrogen:Hydrogen Midi-In 128:0
  ```
- Play a file.
  ```sh
  $ gong play --port 1 mysong
  ```
- Enter the live shell.
  ```sh
  $ gong
  Welcome to the gong shell!
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
  # This is a line comment.
  ```
- #### Tempo change
  ```
  tempo 120
  ```
- #### MIDI channel
  ```
  channel 10
  ```
- #### Global velocity
  ```
  velocity 127
  ```
- #### Program change
  ```
  program 0
  ```
- #### Control change
  ```
  control 1 127
  ```
- #### Note assignment
  ```
  # Kick drum (on the drum channel).
  k=36
  # Middle C (on other channels).
  c=60
  ```
- #### Note values
  ```
  # Whole note.
  x1
  # Half note.
  x2
  # Quarter note (same as x4).
  x
  # 8th note.
  x8
  # 16th note.
  x16
  # 32th note.
  x32
  # And so on...
  ```
- #### Dotted notes and triplets
  ```
  # Dotted quarter note.
  x.
  # Dotted 8th note.
  x8.
  # Quarter triplet note.
  x/3
  # Dotted 8th quintuplet note.
  x8./5
  ```
- #### Note grouping
  ```
  # Ti-Tiri.
  x8 x16 x16
  # Can be written as:
  x8xx16
  # Three 8th triplet notes.
  xxx8/3
  ```
- #### Bars

  ```
  # Define a bar.
  bar "Rock beat"
  xx8 xx8 xx8 xx8
  k   s   k   s
  end

  # Play the bar.
  play "Rock beat"
  ```

## Example of The Bonham Half Time Shuffle

The file is included in the `examples` directory. To play into the default port, run

```sh
$ gong play examples/bonham
```

```
# The Bonham half time shuffle

tempo 132
channel 10
velocity 100

# Kick drum.
k=36
# Acoustic snare drum.
s=38
# Hi-Hat closed.
x=42
# Hi-Hat open.
o=46
# Hi-Hat foot.
X=44
# Crash cymbal.
c=49
# Low tom.
q=45
# Floor tom 2.
g=41

# Start the first bar with a crash cymbal.
bar "bonham 1"
# A crash whole note in the first bar.
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

# Count in.
xxxo

# Play 8 bars of the Bonham groove.
play "bonham 1"
play "bonham 2"
play "bonham 2"
play "fill"
play "bonham 1"
play "bonham 2"
play "bonham 2"
play "fill"
```

## Possible features in the future

- WebAssembly support with Web MIDI for running in browsers
- Velocity for individual notes
  this makes the syntax complicated. For example a normal 8th triplet `xxx8/3`. To set a velocity 50 for the second note (to emulate a ghost note), one would have to expand the syntax manually and write `x8/3 x8/3(50) x8/3`.
- Generating an SMF midi file
- Accelerando/Ritardando
