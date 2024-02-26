package client

import (
	"context"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	intrv1 "github.com/jayleonc/geektime-go/webook/api/proto/gen/intr/v1"
	"google.golang.org/grpc"
	"math/rand"
)

type InteractiveClient struct {
	remote intrv1.InteractiveServiceClient
	local  intrv1.InteractiveServiceClient

	threshold *atomicx.Value[int32]
}

func (i *InteractiveClient) IncrReadCnt(ctx context.Context, in *intrv1.IncrReadCntRequest, opts ...grpc.CallOption) (*intrv1.IncrReadCntResponse, error) {
	return i.selectClient().IncrReadCnt(ctx, in, opts...)
}

func (i *InteractiveClient) Like(ctx context.Context, in *intrv1.LikeRequest, opts ...grpc.CallOption) (*intrv1.LikeResponse, error) {
	return i.selectClient().Like(ctx, in, opts...)
}

func (i *InteractiveClient) CancelLike(ctx context.Context, in *intrv1.CancelLikeRequest, opts ...grpc.CallOption) (*intrv1.CancelLikeResponse, error) {
	return i.selectClient().CancelLike(ctx, in, opts...)
}

func (i *InteractiveClient) Collect(ctx context.Context, in *intrv1.CollectRequest, opts ...grpc.CallOption) (*intrv1.CollectResponse, error) {
	return i.selectClient().Collect(ctx, in, opts...)
}

func (i *InteractiveClient) Get(ctx context.Context, in *intrv1.GetRequest, opts ...grpc.CallOption) (*intrv1.GetResponse, error) {
	return i.selectClient().Get(ctx, in, opts...)
}

func (i *InteractiveClient) GetByIds(ctx context.Context, in *intrv1.GetByIdsRequest, opts ...grpc.CallOption) (*intrv1.GetByIdsResponse, error) {
	return i.selectClient().GetByIds(ctx, in, opts...)
}

func (i *InteractiveClient) GetTopNLikedArticles(ctx context.Context, in *intrv1.GetTopNLikedArticlesRequest, opts ...grpc.CallOption) (*intrv1.GetTopNLikedArticlesResponse, error) {
	return i.selectClient().GetTopNLikedArticles(ctx, in, opts...)
}

func NewInteractiveClient(remote intrv1.InteractiveServiceClient, local intrv1.InteractiveServiceClient) *InteractiveClient {
	return &InteractiveClient{
		remote:    remote,
		local:     local,
		threshold: atomicx.NewValue[int32](),
	}
}

func (i *InteractiveClient) selectClient() intrv1.InteractiveServiceClient {
	num := rand.Int31n(100)
	if num < i.threshold.Load() {
		return i.remote
	}
	return i.local
}

func (i *InteractiveClient) UpdateThreshold(val int32) {
	i.threshold.Store(val)
}