// Command life runs the terminal ecosystem simulation.
//
// Modes are chosen automatically:
//   - interactive: stdout is a terminal and --headless is not set
//   - plain: stdout is piped (same frame, no escape sequences)
//   - headless: no frames at all, only the closing summary
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/doogev/terminal-life/internal/render"
	"github.com/doogev/terminal-life/internal/sim"
	"github.com/doogev/terminal-life/internal/world"
)

type mode int

const (
	modeInteractive mode = iota
	modePlain
	modeHeadless
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "life:", err)
		os.Exit(1)
	}
}

func run() error {
	def := world.DefaultParams()
	var (
		width      = flag.Int("w", def.W, "grid width in cells")
		height     = flag.Int("h", def.H, "grid height in cells")
		seed       = flag.Uint64("seed", 1, "RNG seed; same seed and tick count reproduce a run")
		fps        = flag.Int("fps", 20, "frames per second in interactive mode (0 = uncapped)")
		ticks      = flag.Int("ticks", 0, "stop after N ticks (0 = run until interrupted)")
		headless   = flag.Bool("headless", false, "simulate without rendering")
		sun        = flag.Float64("sun", float64(def.SunRate), "sunlight energy per cell per tick")
		seedPlants = flag.Int("plants", 60, "number of initial plant patches")
	)
	flag.Parse()

	p := def
	p.W, p.H = *width, *height
	p.SunRate = float32(*sun)
	if p.W <= 0 || p.H <= 0 {
		return fmt.Errorf("invalid grid %dx%d", p.W, p.H)
	}
	if *fps < 0 || *ticks < 0 || *seedPlants < 0 {
		return fmt.Errorf("fps, ticks and plants must be >= 0")
	}

	wld := world.New(p, *seed)
	world.SeedPlants(wld, *seedPlants)

	m := modePlain
	switch {
	case *headless:
		m = modeHeadless
	case render.IsTTY(os.Stdout):
		m = modeInteractive
	}

	if m == modeInteractive {
		render.EnterAltScreen(os.Stdout)
		defer render.LeaveAltScreen(os.Stdout)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	var ticker *time.Ticker
	if m == modeInteractive && *fps > 0 {
		ticker = time.NewTicker(time.Second / time.Duration(*fps))
		defer ticker.Stop()
	}

	frame := render.NewFrame(p.W, p.H)
	start := time.Now()

	for {
		if *ticks > 0 && wld.Tick >= *ticks {
			break
		}
		if ticker != nil {
			select {
			case <-sigCh:
				return summarize(wld, m, time.Since(start))
			case <-ticker.C:
			}
		} else {
			select {
			case <-sigCh:
				return summarize(wld, m, time.Since(start))
			default:
			}
		}

		sim.Step(wld)

		if m != modeHeadless {
			render.Draw(frame, wld)
			if m == modeInteractive {
				if err := frame.Flush(os.Stdout); err != nil {
					return err
				}
			} else if err := frame.FlushPlain(os.Stdout); err != nil {
				return err
			}
		}
	}

	return summarize(wld, m, time.Since(start))
}

// summarize writes a one-line closing report to stderr so that piped frame
// output on stdout stays byte-clean.
func summarize(w *world.World, m mode, elapsed time.Duration) error {
	modeName := [...]string{"interactive", "plain", "headless"}[m]
	ticksPerSec := 0.0
	if elapsed > 0 {
		ticksPerSec = float64(w.Tick) / elapsed.Seconds()
	}
	fmt.Fprintf(os.Stderr,
		"ticks=%d plants=%.2f agents=%d mode=%s %.0f ticks/s\n",
		w.Tick, w.TotalPlantEnergy(), len(w.Agents), modeName, ticksPerSec)
	return nil
}
