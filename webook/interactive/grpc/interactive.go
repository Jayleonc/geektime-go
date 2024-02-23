package grpc

import (
	"context"
	"github.com/jayleonc/geektime-go/webook/api/proto/gen/intr/v1"
	"github.com/jayleonc/geektime-go/webook/interactive/domain"
	"github.com/jayleonc/geektime-go/webook/interactive/service"
	"google.golang.org/grpc"
)

type InteractiveServiceServer struct {
	intrv1.UnimplementedInteractiveServiceServer
	svc service.InteractiveService
}

func NewInteractiveServiceServer(svc service.InteractiveService) *InteractiveServiceServer {
	return &InteractiveServiceServer{svc: svc}
}

func (i *InteractiveServiceServer) Register(s *grpc.Server) {
	intrv1.RegisterInteractiveServiceServer(s, i)
}

func (i *InteractiveServiceServer) IncrReadCnt(ctx context.Context, request *intrv1.IncrReadCntRequest) (*intrv1.IncrReadCntResponse, error) {
	err := i.svc.IncrReadCnt(ctx, request.GetBiz(), request.GetBizId())
	return &intrv1.IncrReadCntResponse{}, err
}

func (i *InteractiveServiceServer) Like(ctx context.Context, request *intrv1.LikeRequest) (*intrv1.LikeResponse, error) {
	err := i.svc.Like(ctx, request.GetBiz(), request.GetId(), request.GetUid())
	return &intrv1.LikeResponse{}, err
}

func (i *InteractiveServiceServer) CancelLike(ctx context.Context, request *intrv1.CancelLikeRequest) (*intrv1.CancelLikeResponse, error) {
	err := i.svc.CancelLike(ctx, request.GetBiz(), request.GetId(), request.GetUid())
	return &intrv1.CancelLikeResponse{}, err
}

func (i *InteractiveServiceServer) Collect(ctx context.Context, request *intrv1.CollectRequest) (*intrv1.CollectResponse, error) {
	err := i.svc.Collect(ctx, request.GetBiz(), request.GetBizId(), request.GetCid(), request.GetUid())
	return &intrv1.CollectResponse{}, err
}

func (i *InteractiveServiceServer) Get(ctx context.Context, request *intrv1.GetRequest) (*intrv1.GetResponse, error) {
	intr, err := i.svc.Get(ctx, request.GetBiz(), request.GetId(), request.GetUid())
	return &intrv1.GetResponse{
		Intr: i.toDTO(intr),
	}, err
}

func (i *InteractiveServiceServer) GetByIds(ctx context.Context, request *intrv1.GetByIdsRequest) (*intrv1.GetByIdsResponse, error) {
	res, err := i.svc.GetByIds(ctx, request.GetBiz(), request.GetIds())
	if err != nil {
		return nil, err
	}

	var intrs = make(map[int64]*intrv1.Interactive, len(res))
	for k, v := range res {
		intrs[k] = i.toDTO(v)
	}

	return &intrv1.GetByIdsResponse{
		Intrs: intrs,
	}, nil
}

func (i *InteractiveServiceServer) GetTopNLikedArticles(ctx context.Context, request *intrv1.GetTopNLikedArticlesRequest) (*intrv1.GetTopNLikedArticlesResponse, error) {
	res, err := i.svc.GetTopNLikedArticles(ctx, request.GetBiz(), int(request.GetN()))
	if err != nil {
		return nil, err
	}

	var topns = make([]*intrv1.ArticleLike, len(res))
	for _, v := range res {
		topns = append(topns, i.toTopNDTO(v))
	}

	return &intrv1.GetTopNLikedArticlesResponse{
		ArticleLike: topns,
	}, nil
}

func (i *InteractiveServiceServer) toDTO(intr domain.Interactive) *intrv1.Interactive {
	return &intrv1.Interactive{
		BizId:      intr.BizId,
		ReadCnt:    intr.ReadCnt,
		LikeCnt:    intr.LikeCnt,
		CollectCnt: intr.CollectCnt,
		Liked:      intr.Liked,
		Collected:  intr.Collected,
	}
}

func (i *InteractiveServiceServer) toTopNDTO(articleLike domain.ArticleLike) *intrv1.ArticleLike {
	return &intrv1.ArticleLike{
		ArticleId: articleLike.ArticleId,
		LikeCnt:   articleLike.LikeCnt,
	}
}
