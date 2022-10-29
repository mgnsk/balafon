package main

import (
	"context"
	"runtime"

	"github.com/spf13/cobra"
	"gitlab.com/gomidi/midi/v2/drivers"
	"gitlab.com/gomidi/midi/v2/sequencer"
)

func playFile(c *cobra.Command, args []string) error {
	panic("TODO: implement")
	// f, err := util.Open(args[0])
	// if err != nil {
	// 	return err
	// }
	// defer f.Close()

	// it := interpreter.New()

	// song, err := it.EvalAll(f)
	// if err != nil {
	// 	return err
	// }

	// out, err := getPort(c.Flag("port").Value.String())
	// if err != nil {
	// 	return err
	// }

	// if err := out.Open(); err != nil {
	// 	return err
	// }

	// buf := &bytes.Buffer{}

	// s := song.ToSMF1()
	// if _, err := s.WriteTo(buf); err != nil {
	// 	return err
	// }

	// r := smf.ReadTracksFrom(buf)

	// return r.Play(out)
}

func playAll(ctx context.Context, out drivers.Out, events sequencer.Events) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// p := player.New(out)

	// for _, msg := range events {
	// 	if err := p.Play(ctx, msg); err != nil {
	// 		return err
	// 	}
	// }

	return nil
}

// func runPlayer(ctx context.Context, out drivers.Out, resultC <-chan result, tempo uint16) error {
// 	runtime.LockOSThread()
// 	defer runtime.UnlockOSThread()

// 	p := player.New(out)

// 	if tempo > 0 {
// 		p.SetTempo(tempo)
// 	}

// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return ctx.Err()
// 		case res, ok := <-resultC:
// 			if !ok {
// 				return io.ErrClosedPipe
// 			}
// 			if res.input != "" {
// 				fmt.Println(res.input)
// 			}
// 			for _, msg := range res.messages {
// 				if err := p.Play(ctx, msg); err != nil {
// 					return err
// 				}
// 			}
// 		}
// 	}
// }
