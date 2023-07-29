# Balafon

> The balafon is a gourd-resonated xylophone, a type of struck idiophone. It is closely associated with the neighbouring Mandé, Senoufo and Gur peoples of West Africa, particularly the Guinean branch of the Mandinka ethnic group, but is now found across West Africa from Guinea to Mali. Its common name, balafon, is likely a European coinage combining its Mandinka name ߓߟߊ bala with the word ߝߐ߲ fôn 'to speak' or the Greek root phono.
>
> [Balafon](https://en.wikipedia.org/wiki/Balafon) - From Wikipedia, the free encyclopedia

## Introduction

balafon is a multitrack MIDI sequencer language and interpreter. It consists of a live shell, player and a linter.

## Install

To install the `balafon` command from source, `go` and `rtmidi` are required.
Not tested on platforms other than Linux.

```sh
go install github.com/mgnsk/balafon/cmd/balafon@latest
```

## Running

- The default command lists the available MIDI ports. The default port is the 0 port.

```sh
balafon
0: Midi Through:Midi Through Port-0 14:0
1: Hydrogen:Hydrogen Midi-In 135:0
2: VMPK Input:in 128:0
```

- Play a file through a specific port. The port name must contain the passed in flag value:

```sh
balafon play --port "VMPK" examples/bach.bal
```

- Port can also be specified by its number:

```sh
balafon play --port 2 examples/bonham.bal
```

- Enter live mode:

```sh
balafon live --port hydro examples/live_drumset.bal
```

Live mode is an unbuffered input mode in the shell. Whenever an assigned key is pressed,
a note on message is sent to the port.

- Lint a file:

```sh
balafon lint examples/bonham.bal
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
  convert     Convert a file to SMF2
  fmt         Format a file
  help        Help about any command
  lint        Lint a file
  live        Load a file and continue in a live shell
  play        Play a file

Flags:
  -h, --help   help for this command

Use " [command] --help" for more information about a command.
```

## Syntax

The language consists of commands and note lists.

### Comments

Only block comments are supported for now.

```
/* This is a block comment. */

/*
This is a multi line block comment.
*/
```

### Commands

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

### Note assignment

Assign a MIDI note number to a note letter.

```
// Kick drum (on the drum channel).
:assign k 36
// Middle C (on other channels).
:assign c 60
```

### Notes

Notes are written as a letter symbol (must be assigned first) plus properties.
The available properties are

- sharp (`#`)
- flat (`$`)
- staccato (`` ` ``) shorten note by 50%
- accent (`>`) +5 velocity
- marcato (`^`) +10 velocity
- ghost (`)`) -5 velocity
- dot (`.`)
- numeric note value (`1`, `2`, `4`, `8` and so on)
- tuplet (`/3`) (only triplet `/3` or quintuplet `/5`)
- let ring (`*`)

### Note values

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

### Rests

```
// A quarter rest.
-
// An 8th rest.
-8
```

### Dotted notes and tuplets

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

### Flat and sharp notes

```
// A note.
c
// A sharp note (MIDI note number + 1).
c#
// A flat note (MIDI note number - 1).
c$
```

### Note grouping

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

### Additive properties

When used on note groups, these properties are added to the notes' already existing properties:

- sharp (`#`)
- flat (`$`)
- staccato (`` ` ``)
- accent (`>`)
- marcato (`^`)
- ghost (`)`)
- dot (`.`)

### Unique properties

When used on note groups, these properties override the notes' existing properties of the same type:

- numeric note value (`1`, `2`, `4`, `8` and so on)
- tuplet (`/3`) (only triplet `/3` or quintuplet `/5`)
- let ring (`*`)

### Bars

Bars are used to specify multiple tracks playing at once.
Only `timesig`, `velocity` and `channel` commands are scoped to the bar.
Other commands, when used inside a bar, have global effect when the bar is played back.
The bar is executed with the `play` command.

```
// Define a bar.
:bar RockBeat
// Setting timesig makes the interpreter validate the bar length.
// Incomplete bars are filled with silence.
:timesig 4 4
  [xx xx xx xx]8
  // Using braces for nice alignment.
  [k  s  k  s]
:end

// You can also write the same bar as:
:bar SameBeat
  [xxxxxxxx]8
  ksks
:end

// Play the bar.
:play RockBeat
```

## Neovim configuration

### Filetype detection:

```lua
vim.api.nvim_create_autocmd({ "BufNewFile", "BufRead" }, {
    pattern = "*.bal",
    callback = function()
        vim.bo.filetype = "balafon"
    end,
})
```

### Neomake

```lua
vim.g.neomake_balafon_lint_maker = {
    exe = "balafon",
    args = "lint",
    errorformat = "%f:%l:%c: error: %m",
}

vim.g.neomake_balafon_enabled_makers = { "lint" }
```

### Treesitter

Example configuration for lazy.nvim:

```lua
{
    "mgnsk/tree-sitter-balafon",
    dependencies = {
        "nvim-treesitter/nvim-treesitter",
    },
    build = function()
        local parser_config = require("nvim-treesitter.parsers").get_parser_configs()
        parser_config.balafon = {
            install_info = {
                url = "https://github.com/mgnsk/tree-sitter-balafon",
                files = { "src/parser.c" },
                branch = "main",
                generate_requires_npm = true,
                requires_generate_from_grammar = false,
            },
            filetype = "bal",
        }
        vim.cmd("TSUpdate balafon")
	end,
},
```

## Examples

### The Bonham Half Time Shuffle

[examples/bonham.bal](examples/bonham.bal)

To play into the default port, run

```sh
balafon play examples/bonham.bal
```

### J.S. Bach - Musikalisches Opfer - 6. Canon A 2 Per Tonos

[examples/bach.bal](examples/bach.bal)

To play into the default port, run

```sh
balafon play examples/bach.bal
```

### Multichannel

[examples/multichannel.bal](examples/multichannel.bal)

To play into the default port, run

```sh
balafon play examples/bach.bal
```

## Possible features in the future

- Tie (a curved line connecting the heads of two notes of the same pitch) - no idea about the syntax. Can be partially emulated by using dotted notes if the rhythm is simple enough.
- WebAssembly support with Web MIDI for running in browsers.
- Accelerando/Ritardando.
