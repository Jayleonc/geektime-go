package web

import (
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
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
	intrSvc service.InteractiveService
	l       logger.Logger
	biz     string
}

func NewArticleHandler(l logger.Logger, svc service.ArticleService, intrSvc service.InteractiveService) *ArticleHandler {
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
}

// Edit 接收一个 Article 输入，返回文章 ID
func (h *ArticleHandler) Edit(ctx *gin.Context, req vo.ArticleEditReq, uc ijwt.UserClaims) (ginx.Response, error) {

	id, err := h.svc.Save(ctx, domain.Article{
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
		intr domain.Interactive
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
