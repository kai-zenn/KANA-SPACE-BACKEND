package workers

import (
    "log"
    "os"

    "github.com/robfig/cron/v3"
)

func NewScheduler(autoCancel *AutoCancelWorker) *cron.Cron {
  logger := cron.VerbosePrintfLogger(log.New(os.Stdout, "cron: ", log.LstdFlags))

  c := cron.New(cron.WithChain(
    cron.SkipIfStillRunning(logger),
    cron.Recover(logger),
  ))

  if _, err := c.AddFunc("@every 10m", autoCancel.Run); err != nil {
    log.Fatalf("gagal register auto-cancel worker: %v", err)
  }
  return c
}
