package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/pillash/mp4util"
)

var (
	address = flag.String("address", "http://localhost:8080", "Address of the timelapse queue instance")

	// TODO: test doesn't work under headless for some reason...

	// For development, run this test with `--wait --headless=false`
	headless = flag.Bool("headless", false, "Whether to run chrome in headless mode")
	wait     = flag.Bool("wait", false, "Whether to wait for SIGTERM")
	cleanup  = flag.Bool("cleanup", true, "Whether to delete generated files once completed")
)

func WaitForTerm() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}

func TestIntegration(t *testing.T) {
	opts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
	}
	if *headless {
		opts = append(opts, chromedp.Headless)
	}

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	first_folder := `document.querySelector("tq-app").shadowRoot.querySelector("tq-browse").shadowRoot.querySelector("div.files.dirs > div:nth-child(1) > paper-button")`

	new_job := `document.querySelector("tq-app").shadowRoot.querySelector("tq-browse").shadowRoot.querySelector("div.files.timelapses > div.timelapse > paper-button")`

	cropper := `document.querySelector("tq-app").shadowRoot.querySelector("tq-setup").shadowRoot.querySelector(".cropper-crop-box")`

	output_filename := `document.querySelector("tq-app").shadowRoot.querySelector("tq-setup").shadowRoot.querySelector("#output-filename")`

	add_button := `document.querySelector("tq-app").shadowRoot.querySelector("tq-setup").shadowRoot.querySelector("#add-button")`

	queue_done := `document.querySelector("tq-app").shadowRoot.querySelector("tq-queue").shadowRoot.querySelector("div.queue-item.queue-done")`

	remove_job := `document.querySelector("tq-app").shadowRoot.querySelector("tq-queue").shadowRoot.querySelector("div.queue-item.queue-done .remove-button")`

	// TODO: chromedp.Click doesn't seem to work, uncaught exceptions?

	shortSleep := func(ctx context.Context) error {
		// TODO: ideally would replace this with a wait on the polymer dom
		// flush
		time.Sleep(time.Second)
		return nil
	}

	var dummy bool

	// Clean up our output files once we're finished testing.
	defer func() {
		if *cleanup {
			os.Remove("testdata/output.mp4")
			os.Remove("testdata/output.mp4.log")
		}
	}()

	// Start the test timelapse conversion.
	runCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	if *wait {
		runCtx = ctx
	}
	if err := chromedp.Run(runCtx,
		chromedp.Navigate(*address),
		// Click first folder.
		chromedp.WaitVisible(first_folder, chromedp.ByJSPath),
		chromedp.ActionFunc(shortSleep),
		chromedp.Evaluate(first_folder+".click(); true;", &dummy),
		// Click first timelapse.
		chromedp.WaitVisible(new_job, chromedp.ByJSPath),
		chromedp.ActionFunc(shortSleep),
		chromedp.Evaluate(new_job+".click(); true;", &dummy),
		// Wait for cropper to finish loading.
		chromedp.WaitVisible(cropper, chromedp.ByJSPath),
		// Input filename.
		chromedp.WaitVisible(output_filename, chromedp.ByJSPath),
		chromedp.ActionFunc(shortSleep),
		chromedp.SendKeys(output_filename, "output", chromedp.ByJSPath),
		// Add timelapse to the queue.
		chromedp.ActionFunc(shortSleep),
		chromedp.Evaluate(add_button+".click(); true;", &dummy),
		// Wait for timelapse conversion to complete.
		chromedp.ActionFunc(shortSleep),
		chromedp.WaitVisible(queue_done, chromedp.ByJSPath),
		// Remove the finished job.
		chromedp.WaitVisible(remove_job, chromedp.ByJSPath),
		chromedp.ActionFunc(shortSleep),
		chromedp.Evaluate(remove_job+".click(); true;", &dummy),
	); err != nil {
		if outLog, err := ioutil.ReadFile("testdata/output.mp4.log"); err != nil {
			log.Printf("FFmpeg output:\n%s", string(outLog))
		}
		t.Fatalf("chromedp run: %v", err)
	}

	log.Printf("Timelapse generation successful")

	if *wait {
		t.Logf("Waiting for Ctrl+C before continuing")
		WaitForTerm()
	}

	// Read and echo the FFmpeg output log
	outLog, err := ioutil.ReadFile("testdata/output.mp4.log")
	if err != nil {
		t.Fatalf("read FFmpeg output log: %v", err)
	}
	log.Printf("FFmpeg output:\n%s", string(outLog))

	// Sanity check that our video file was generated correctly
	if _, err := mp4util.Duration("testdata/output.mp4"); err != nil {
		t.Fatalf("get output mp4 duration: %v", err)
	}
}
