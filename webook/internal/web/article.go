package web

import (
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	domain2 "github.com/jayleonc/geektime-go/webook/interactive/domain"
	service2 "github.com/jayleonc/geektime-go/webook/interactive/service"
	"github.com/jayleonc/geektime-go/webook/internal/domain"
	"github.com/jayleonc/geektime-go/webook/internal/service"
	ijwt "github.com/jayleonc/geektime-go/webook/internal/web/jwt"
	"github.com/jayleonc/geektime-go/webook/internal/web/vo"
	"github.com/jayleonc/geektime-go/webook/pkg/ginx"
	"github.com/jayleonc/geektime-go/webook/pkg/logger"
	"golang.org/x/sync/errgroup"
	"net/http"
	"strconv"
	"time"
)

type ArticleHandler struct {
	svc     service.ArticleService
	intrSvc service2.InteractiveService
	l       logger.Logger
	biz     string
}

func NewArticleHandler(l logger.Logger, svc service.ArticleService, intrSvc service2.InteractiveService) *ArticleHandler {
	return &ArticleHandler{
		l:       l,
		svc:     svc,
		intrSvc: intrSvc,
		biz:     "article",
	}
}

func (h *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/articles")

	g.POST("/edit", ginx.WrapBodyAndClaims(h.Edit))
	g.POST("/publish", ginx.WrapBodyAndClaims(h.Publish))
	g.GET("/detail/:id", h.Detail)
	g.POST("/list", h.List)

	pub := g.Group("/pub")
	pub.GET("/:id", h.PubDetail)

	pub.POST("/like", ginx.WrapBodyAndClaims(h.Like))
	pub.POST("/collect", ginx.WrapBodyAndClaims(h.Collect))
	pub.GET("/top/:n", h.TopNArticle)
}

// Edit 接收一个 Article 输入，返回文章 ID
// 创建一个 Article
func (h *ArticleHandler) Edit(ctx *gin.Context, req vo.ArticleEditReq, uc ijwt.UserClaims) (ginx.Response, error) {

	id, err := h.svc.Save(ctx, h.biz, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author:  domain.Author{Id: uc.Uid},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Response{
			Msg: "系统错误",
		})
		h.l.Error("保存文章失败", logger.Int64("uid", uc.Uid), logger.Error(err))
	}
	return ginx.Response{Data: id}, nil
}

// Publish 发布文章，也可能是修订文章
func (h *ArticleHandler) Publish(ctx *gin.Context, req vo.ArticlePublishReq, uc ijwt.UserClaims) (ginx.Response, error) {

	id, err := h.svc.Publish(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author:  domain.Author{Id: uc.Uid},
	})
	if err != nil {
		return ginx.Response{Code: 5, Msg: "系统错误"}, err
	}
	return ginx.Response{Data: id}, nil
}

func (h *ArticleHandler) Detail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ginx.Error(ctx, http.StatusOK, "id 参数错误")
		return
	}
	art, err := h.svc.GetById(ctx, id)
	if err != nil {
		ginx.Error(ctx, 5, "系统错误")
		return
	}
	uc := ctx.MustGet("user").(ijwt.UserClaims)
	if uc.Uid != art.Author.Id {
		ginx.Error(ctx, 5, "系统错误")
		return
	}
	v := vo.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
		Ctime:    art.Ctime.Format(time.DateTime),
		Utime:    art.Utime.Format(time.DateTime),
	}

	ginx.OK(ctx, ginx.Response{Data: v})
}

func (h *ArticleHandler) List(ctx *gin.Context) {
	type Res struct {
		Title string
		Page
	}
	var res Res
	if err := ctx.Bind(&res); err != nil {
		return
	}

	uc := ctx.MustGet("user").(ijwt.UserClaims)
	articles, count, err := h.svc.GetByAuthor(ctx, uc.Uid, res.PageIndex, res.PageSize, res.Title)
	if err != nil {
		ginx.Error(ctx, 5, err.Error())
		return
	}
	result := slice.Map(articles, func(idx int, src domain.Article) vo.Article {
		return vo.Article{
			Id:       src.Id,
			Title:    src.Title,
			Abstract: src.Abstract(),
			Content:  src.Content,
			AuthorId: src.Author.Id,
			Status:   src.Status.ToUint8(),
		}
	})
	g := ginx.Page{
		List:      result,
		Count:     count,
		PageIndex: res.PageIndex,
		PageSize:  res.PageSize,
	}
	ginx.PageOK(ctx, g, "")
}

func (h *ArticleHandler) PubDetail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ginx.Error(ctx, http.StatusOK, "id 参数错误")
		return
	}

	var (
		eg   errgroup.Group
		art  domain.Article
		intr domain2.Interactive
	)

	uc := ctx.MustGet("user").(ijwt.UserClaims)
	eg.Go(func() error {
		var er error
		art, er = h.svc.GetPubById(ctx, id, uc.Uid)
		return er
	})

	// todo 这里有 bug
	eg.Go(func() error {
		var er error
		intr, er = h.intrSvc.Get(ctx, h.biz, id, uc.Uid)
		return er
	})
	err = eg.Wait()
	if err != nil {
		ginx.Error(ctx, 5, "系统错误")
	}

	v := vo.Article{
		Id:         art.Id,
		Title:      art.Title,
		Content:    art.Content,
		AuthorId:   art.Author.Id,
		AuthorName: art.Author.Name,
		Status:     art.Status.ToUint8(),
		Ctime:      art.Ctime.Format(time.DateTime),
		Utime:      art.Utime.Format(time.DateTime),

		ReadCnt:    intr.ReadCnt,
		LikeCnt:    intr.LikeCnt,
		CollectCnt: intr.CollectCnt,
		Liked:      intr.Liked,
		Collected:  intr.Collected,
	}
	ginx.OK(ctx, ginx.Response{Data: v})
}

func (h *ArticleHandler) Like(ctx *gin.Context, req vo.ArticleLikeReq, uc ijwt.UserClaims) (ginx.Response, error) {
	var err error
	if req.Like {
		err = h.intrSvc.Like(ctx, h.biz, req.Id, uc.Uid)
	} else {
		err = h.intrSvc.CancelLike(ctx, h.biz, req.Id, uc.Uid)
	}
	if err != nil {
		ginx.Error(ctx, 5, "系统错误")
		return ginx.Response{Code: 5, Msg: "系统错误"}, err
	}
	return ginx.Response{Msg: "OK"}, nil
}

func (h *ArticleHandler) Collect(ctx *gin.Context, req vo.ArticleCollectReq, uc ijwt.UserClaims) (ginx.Response, error) {
	type Req struct {
		Id  int64 `json:"id,omitempty"`
		Cid int64 `json:"cid,omitempty"`
	}

	err := h.intrSvc.Collect(ctx, h.biz, req.Id, req.Cid, uc.Uid)

	if err != nil {
		ginx.Error(ctx, 5, "系统错误")
		return ginx.Response{Code: 5, Msg: "系统错误"}, err
	}
	return ginx.Response{Msg: "OK"}, nil
}

// TopNArticle 处理获取点赞数前N的文章的请求
// todo
// 批量查询：从 Interactive 获取到 ArticleLike 数据集后
// 可以一次性地从 Article 的缓存中查询所有相关的文章。
// 这样做可以减少缓存访问次数，提高效率。
// 缓存穿透：对于缓存中不存在的数据，查询数据库后应立即更新缓存，避免后续相同的查询再次穿透到数据库。
func (h *ArticleHandler) TopNArticle(c *gin.Context) {
	// 从URL参数中获取N的值
	nStr := c.Param("n")
	n, err := strconv.Atoi(nStr)
	if err != nil {
		// 如果N不是一个有效的整数，则返回错误
		h.l.Error("Invalid parameter", logger.Error(err))
		ginx.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 使用 InteractiveService 的 GetTopNLikedArticles 方法获取数据
	// 这里获取的是 文章ID 和 点赞数
	topArticles, err := h.intrSvc.GetTopNLikedArticles(c.Request.Context(), h.biz, n)
	if err != nil {
		// 如果查询过程中出现错误，则返回错误
		h.l.Error("Error getting top N liked articles", logger.Error(err))
		ginx.Error(c, 5, err.Error())
		return
	}

	// 构建文章ID的切片
	ids := make([]int64, len(topArticles))
	for i, al := range topArticles {
		ids[i] = al.ArticleId
	}

	// 使用 ArticleService 的 GetByIds 方法获取数据
	// 获取的是 TopN 的文章ID、标题和摘要
	articles, err := h.svc.GetByIds(c.Request.Context(), ids)
	if err != nil {
		ginx.Error(c, 5, err.Error())
		return
	}

	articlesMap := make(map[int64]domain.Article)
	for _, article := range articles {
		articlesMap[article.Id] = article
	}

	var sortedArticles []domain.Article
	for _, id := range ids {
		if article, exists := articlesMap[id]; exists {
			sortedArticles = append(sortedArticles, article)
		}
	}

	ginx.OK(c, ginx.Response{
		Data: sortedArticles,
	})
}
