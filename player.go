package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"runtime"

	"github.com/mgnsk/gong/internal/interpreter"
	"github.com/mgnsk/gong/internal/player"
	"github.com/spf13/cobra"
	"gitlab.com/gomidi/midi/v2"
)

func playFile(c *cobra.Command, args []string) error {
	f, err := stdinOrFile(args)
	if err != nil {
		return err
	}
	defer f.Close()

	it := interpreter.New()
	messages, err := it.EvalAll(f)
	if err != nil {
		return err
	}

	out, err := getPort(c.Flag("port").Value.String())
	if err != nil {
		return err
	}

	if err := out.Open(); err != nil {
		return err
	}

	playAll(context.Background(), out, messages)

	return nil
}

func playAll(ctx context.Context, out midi.Sender, messages []interpreter.Message) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	p := player.New(out)
	for _, msg := range messages {
		if err := p.Play(ctx, msg); err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			log.Fatal(err)
		}
	}
}

func startPlayer(ctx context.Context, out midi.Sender, resultC <-chan result, tempo uint16) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	p := player.New(out)
	if tempo > 0 {
		p.SetTempo(tempo)
	}
	for {
		select {
		case <-ctx.Done():
			return
		case res, ok := <-resultC:
			if !ok {
				return
			}
			if res.input != "" {
				fmt.Println(res.input)
			}
			for _, msg := range res.messages {
				if err := p.Play(ctx, msg); err != nil {
					if errors.Is(err, context.Canceled) {
						return
					}
					log.Fatal(err)
				}
			}
		}
	}
}
