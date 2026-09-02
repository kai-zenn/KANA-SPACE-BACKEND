package workers

import (
  "context"
  "log"
  "time"

  "KANA-SPACE-BACKEND/internal/modules/lapak"
)

type AutoCancelWorker struct {
  tnansactionUsecase lapak.ITransactionUseCase
}

func NewAutoCancelWorker(uc lapak.ITransactionUseCase) *AutoCancelWorker {
  return &AutoCancelWorker{tnansactionUsecase: uc}
}

func (w *AutoCancelWorker) Run() {
  ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
  defer cancel()

  count, err := w.tnansactionUsecase.ExpireStaleTransactions(ctx)
  if err != nil {
    log.Printf("[AutoCancelWorker] gagal: %v", err)
    return
  }
  if count > 0 {
    log.Printf("[AutoCancelWorker] meng-expire %d transaksi LOCKED yang kadaluarsa", count)
  }
}
