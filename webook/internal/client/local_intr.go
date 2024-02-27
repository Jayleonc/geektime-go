package client

import (
	"context"
	intrv1 "github.com/jayleonc/geektime-go/webook/api/proto/gen/intr/v1"
	"github.com/jayleonc/geektime-go/webook/interactive/domain"
	"github.com/jayleonc/geektime-go/webook/interactive/service"
	"google.golang.org/grpc"
)

type LocalInteractiveServiceAdapter struct {
	svc service.InteractiveService
}

func NewLocalInteractiveServiceAdapter(svc service.InteractiveService) *LocalInteractiveServiceAdapter {
	return &LocalInteractiveServiceAdapter{svc: svc}
}

func (l *LocalInteractiveServiceAdapter) IncrReadCnt(ctx context.Context, in *intrv1.IncrReadCntRequest, opts ...grpc.CallOption) (*intrv1.IncrReadCntResponse, error) {
	err := l.svc.IncrReadCnt(ctx, in.GetBiz(), in.GetBizId())
	return &intrv1.IncrReadCntResponse{}, err
}

func (l *LocalInteractiveServiceAdapter) Like(ctx context.Context, in *intrv1.LikeRequest, opts ...grpc.CallOption) (*intrv1.LikeResponse, error) {
	err := l.svc.Like(ctx, in.GetBiz(), in.GetId(), in.GetUid())
	return &intrv1.LikeResponse{}, err
}

func (l *LocalInteractiveServiceAdapter) CancelLike(ctx context.Context, in *intrv1.CancelLikeRequest, opts ...grpc.CallOption) (*intrv1.CancelLikeResponse, error) {
	err := l.svc.CancelLike(ctx, in.GetBiz(), in.GetId(), in.GetUid())
	return &intrv1.CancelLikeResponse{}, err
}

func (l *LocalInteractiveServiceAdapter) Collect(ctx context.Context, in *intrv1.CollectRequest, opts ...grpc.CallOption) (*intrv1.CollectResponse, error) {
	err := l.svc.Collect(ctx, in.GetBiz(), in.GetBizId(), in.GetCid(), in.GetUid())
	return &intrv1.CollectResponse{}, err
}

func (l *LocalInteractiveServiceAdapter) Get(ctx context.Context, in *intrv1.GetRequest, opts ...grpc.CallOption) (*intrv1.GetResponse, error) {
	intr, err := l.svc.Get(ctx, in.GetBiz(), in.GetId(), in.GetUid())
	return &intrv1.GetResponse{
		Intr: l.toDTO(intr),
	}, err
}

func (l *LocalInteractiveServiceAdapter) GetByIds(ctx context.Context, in *intrv1.GetByIdsRequest, opts ...grpc.CallOption) (*intrv1.GetByIdsResponse, error) {
	res, err := l.svc.GetByIds(ctx, in.GetBiz(), in.GetIds())
	if err != nil {
		return nil, err
	}

	var intrs = make(map[int64]*intrv1.Interactive, len(res))
	for k, v := range res {
		intrs[k] = l.toDTO(v)
	}

	return &intrv1.GetByIdsResponse{
		Intrs: intrs,
	}, nil
}

func (l *LocalInteractiveServiceAdapter) GetTopNLikedArticles(ctx context.Context, in *intrv1.GetTopNLikedArticlesRequest, opts ...grpc.CallOption) (*intrv1.GetTopNLikedArticlesResponse, error) {
	res, err := l.svc.GetTopNLikedArticles(ctx, in.GetBiz(), int(in.GetN()))
	if err != nil {
		return nil, err
	}

	var topns []*intrv1.ArticleLike
	for _, v := range res {
		topns = append(topns, l.toTopNDTO(v))
	}

	return &intrv1.GetTopNLikedArticlesResponse{
		ArticleLike: topns,
	}, nil
}

func (l *LocalInteractiveServiceAdapter) toDTO(intr domain.Interactive) *intrv1.Interactive {
	return &intrv1.Interactive{
		BizId:      intr.BizId,
		ReadCnt:    intr.ReadCnt,
		LikeCnt:    intr.LikeCnt,
		CollectCnt: intr.CollectCnt,
		Liked:      intr.Liked,
		Collected:  intr.Collected,
	}
}

func (l *LocalInteractiveServiceAdapter) toTopNDTO(articleLike domain.ArticleLike) *intrv1.ArticleLike {
	return &intrv1.ArticleLike{
		ArticleId: articleLike.ArticleId,
		LikeCnt:   articleLike.LikeCnt,
	}
}
