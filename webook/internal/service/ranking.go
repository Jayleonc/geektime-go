package service

import (
	"context"
	"github.com/ecodeclub/ekit/queue"
	"github.com/ecodeclub/ekit/slice"
	intrv1 "github.com/jayleonc/geektime-go/webook/api/proto/gen/intr/v1"
	"github.com/jayleonc/geektime-go/webook/internal/domain"
	"github.com/jayleonc/geektime-go/webook/internal/repository"
	"math"
	"time"
)

type RankingService interface {
	TopN(ctx context.Context) error
	GetTopN(ctx context.Context) ([]domain.Article, error)
}

type BatchRankingService struct {
	interSvc  intrv1.InteractiveServiceClient
	artSvc    ArticleService
	batchSize int
	n         int
	scoreFunc func(likeCnt int64, utime time.Time) float64

	repo repository.RankingRepository
}

func (b *BatchRankingService) GetTopN(ctx context.Context) ([]domain.Article, error) {
	return b.repo.GetTopN(ctx)
}

func NewBatchRankingService(interSvc intrv1.InteractiveServiceClient, artSvc ArticleService, repo repository.RankingRepository) RankingService {
	return &BatchRankingService{
		interSvc:  interSvc,
		artSvc:    artSvc,
		batchSize: 100,
		n:         100,
		scoreFunc: func(likeCnt int64, utime time.Time) float64 {
			duration := time.Since(utime).Seconds()
			return float64(likeCnt-1) / math.Pow(duration*2, 1.5)
		},
		repo: repo,
	}
}

func (b *BatchRankingService) TopN(ctx context.Context) error {
	articles, err := b.topN(ctx)
	if err != nil {
		return err
	}
	return b.repo.ReplaceTopN(ctx, articles)
}

func (b *BatchRankingService) topN(ctx context.Context) ([]domain.Article, error) {
	offset := 0
	start := time.Now()
	ddl := start.Add(-7 * 24 * time.Hour)

	type Score struct {
		score float64
		art   domain.Article
	}
	topN := queue.NewPriorityQueue[Score](b.n,
		func(src Score, dst Score) int {
			if src.score > dst.score {
				return 1
			} else if src.score == dst.score {
				return 0
			} else {
				return -1
			}
		})

	for {
		// 取数据
		arts, err := b.artSvc.ListPub(ctx, start, offset, b.batchSize)
		if err != nil {
			return nil, err
		}
		//if len(arts) == 0 {
		//	break
		//}
		ids := slice.Map(arts, func(idx int, art domain.Article) int64 {
			return art.Id
		})
		// 取点赞数
		intrResp, err := b.interSvc.GetByIds(ctx, &intrv1.GetByIdsRequest{
			Biz: "article", Ids: ids,
		})
		intrMap := intrResp.Intrs
		if err != nil {
			return nil, err
		}
		for _, art := range arts {
			intr, ok := intrMap[art.Id]
			if !ok || intr == nil {
				continue
			}
			score := b.scoreFunc(intr.LikeCnt, art.Utime)
			ele := Score{
				score: score,
				art:   art,
			}
			err = topN.Enqueue(ele)
			if err == queue.ErrOutOfCapacity {
				// 这个也是满了
				// 拿出最小的元素
				minEle, _ := topN.Dequeue()
				if minEle.score < score {
					_ = topN.Enqueue(ele)
				} else {
					_ = topN.Enqueue(ele)
				}
			}
		}
		offset = offset + len(arts)
		// 没有取够一批，我们就直接中断执行
		// 没有下一批了
		if len(arts) < b.batchSize ||
			// 这个是一个优化
			arts[len(arts)-1].Utime.Before(ddl) {
			break
		}
	}

	// 这边 topN 里面就是最终结果
	res := make([]domain.Article, topN.Len())
	for i := topN.Len() - 1; i >= 0; i-- {
		ele, _ := topN.Dequeue()
		res[i] = ele.art
	}
	return res, nil
}
