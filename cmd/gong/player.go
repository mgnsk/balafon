package main

import (
	"context"
	"fmt"
	"io"
	"runtime"

	"github.com/mgnsk/gong/internal/interpreter"
	"github.com/mgnsk/gong/internal/player"
	"github.com/mgnsk/gong/internal/util"
	"github.com/spf13/cobra"
	"gitlab.com/gomidi/midi/v2/drivers"
)

func playFile(c *cobra.Command, args []string) error {
	f, err := util.Open(args[0])
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

	if err := playAll(context.TODO(), out, messages); err != nil {
		return err
	}

	return nil
}

func playAll(ctx context.Context, out drivers.Out, messages []interpreter.Message) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	p := player.New(out)

	for _, msg := range messages {
		if err := p.Play(ctx, msg); err != nil {
			return err
		}
	}

	return nil
}

func runPlayer(ctx context.Context, out drivers.Out, resultC <-chan result, tempo uint16) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	p := player.New(out)

	if tempo > 0 {
		p.SetTempo(tempo)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case res, ok := <-resultC:
			if !ok {
				return io.ErrClosedPipe
			}
			if res.input != "" {
				fmt.Println(res.input)
			}
			for _, msg := range res.messages {
				if err := p.Play(ctx, msg); err != nil {
					return err
				}
			}
		}
	}
}
