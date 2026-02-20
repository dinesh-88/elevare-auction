package httptransport

import (
	"context"
	"log"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/usecase"
)

func StartScheduler(ctx context.Context, auctionUC *usecase.AuctionUseCase, lotUC *usecase.LotUseCase, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case t := <-ticker.C:
				if err := auctionUC.SchedulerTick(ctx, t.UTC()); err != nil {
					log.Printf("scheduler tick failed: %v", err)
				}
				if err := lotUC.SchedulerTick(ctx, t.UTC()); err != nil {
					log.Printf("lot scheduler tick failed: %v", err)
				}
			}
		}
	}()
}
