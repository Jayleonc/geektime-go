package domain

// ArticleLike 点赞数前 N 的文章
type ArticleLike struct {
	ArticleId int64
	LikeCnt   int64
}
