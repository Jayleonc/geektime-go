

### 查找 Top N 数据

**核心接口**与**涉及的代码文件**

```go
func (h *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	// 其他业务接口
	pub.POST("/top/:n", h.TopNArticle)
}
```

![image-20240221164624264](/Users/jayleonc/Developer/Environment/go/src/github.com/Jayleonc/geektime-go/webook/internal/web/readme/image-20240221164624264.png)

**主要实现的功能包括：**

- **点赞数的增加与减少**：实现了对文章点赞数的增加和减少的处理，这包括在数据库和Redis缓存中同步更新点赞数。
- **Top N文章的查询**：设计了缓存策略来查询点赞数最多的Top N篇文章，包括如何在缓存命中和未命中的情况下处理数据。
- **文章内容的更新**：当更新文章内容时，检查并更新缓存在Top N列表中的文章，以保持数据一致性。

哈哈，可爱的助教大大，由于时间紧迫，我没有办法绘制 UML 序列图。
下面，我将描述"缓存命中"和"缓存未命中"的场景

1. **客户端请求**：客户端发起请求，查询Top N文章。

   ``` go
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
   ```

2. **数据处理**：服务器接收到请求，查询Redis缓存。

    1. 查询 Interactive 服务的 cache，得到 TopN 区间的文章ID 与 点赞数的有序集合。

        1. “缓存命中”，得到文章ID与点赞数的有序集合，返回。

           ```go
               // 尝试从缓存获取数据
               inters, err := c.cache.GetTopNLikedInteractive(ctx, biz, n)
               if err == nil && len(inters) >= n { // 如果缓存命中且数据量充足
                   return inters, nil
               }
           ```

        2. “缓存未命中” 或 数据不足 N，从数据库加载点赞数前 N 的数据，并异步存储到 cache 中。

           ```go
               // 如果缓存未命中或数据不足，从 DB 获取点赞数前 N 的 Interactive 数据
               interactives, err := c.dao.GetTopNLikedInteractive(ctx, biz, n)
               if err != nil {
                   return nil, err
               }
           
               var articleLikes []domain.ArticleLike
               // 转换 interactives 切片到 ArticleLike 结构
               for _, interactive := range interactives {
                   articleLike := domain.ArticleLike{
                       ArticleId: interactive.BizId,
                       LikeCnt:   interactive.LikeCnt,
                   }
                   articleLikes = append(articleLikes, articleLike)
               }
           
               // 异步更新缓存以反映最新的Top N数据
               // 下次进来的时候，就能从缓存中直接获取
               go func() {
                   if er := c.cache.SetTopNLikedInteractive(ctx, biz, articleLikes); er != nil {
                       // 在实际应用中，这里应该记录日志或进行其他错误处理
                   }
               }()
           ```

    2. 通过文章 ID 列表，查询 Article 服务的 cache，得到文章信息的哈希数组（单条数据包括文章ID、文章标题和文章摘要）

        1. 这里我把缓存命中和未命中的情况考虑到一起，如果从 cache 得到的数据量不足（即有遗漏的文章 ID 不能从 redis 的哈希中查询出来），就去数据库查找遗漏 ID 的文章信息，然后更新到 cache，并与已经从 cache 找到的文章，一起返回

           ```go
           func (c *CachedArticleRepository) GetByIds(ctx context.Context, ids []int64) ([]domain.Article, error) {
               // 尝试从缓存中批量获取文章数据
               articles, missingIds, err := c.cache.GetArticlesByIds(ctx, ids)
               if err != nil {
                  return nil, err
               }
           
               var missingArticles []domain.Article
           
               // 如果有缓存未命中的ID，从数据库获取这些文章数据
               if len(missingIds) > 0 {
                  missingArticlesDao, err := c.dao.GetByIds(ctx, missingIds)
                  if err != nil {
                     return nil, err
                  }
           
                  // 转换 dao.PublishedArticle 到 domain.Article
                  for _, daoArt := range missingArticlesDao {
                     domainArt := c.toDomain(dao.Article(daoArt))
                     missingArticles = append(missingArticles, domainArt)
                  }
           
                  // 将从数据库中获取的文章添加到最终结果列表中
                  articles = append(articles, missingArticles...)
           
                  // 异步更新缓存
                  go func() {
                     if er := c.cache.SetArticles(ctx, missingArticles); er != nil {
                        fmt.Println(er)
                     }
                  }()
               }
           
               return articles, nil
           }
           ```

3. **更新与维护**：考虑到文章更新与点赞的增加和取消。

    1. 当点赞数的增加与减少：实现了对文章点赞数的增加和减少的处理，这包括在数据库和Redis缓存中同步更新点赞数。
    2. 当更新文章内容时，检查该文章是否存在于缓存，存在就更新缓存在Top N列表中的文章，以保持数据一致性。

![image-20240221161518199](/Users/jayleonc/Developer/Environment/go/src/github.com/Jayleonc/geektime-go/webook/internal/web/readme/image-20240221161518199.png)

### 性能测试

**测试命令**：

```bash
wrk -t2 -c10 -d30s -s topn.lua http://localhost:8080
```

**机器参数**：

- CPU: 8核 Apple M1 Pro
- 内存: 16GB RAM

**性能测试结果**：

- ![image-20240221161802874](/Users/jayleonc/Developer/Environment/go/src/github.com/Jayleonc/geektime-go/webook/internal/web/readme/image-20240221161802874.png)

**其他说明**：

1. 当前实现的代码，仍有一些缺陷，比如考虑不周的：当文章删除，或者被作者下线之后，如何退出 TopN 列表等等，时间问题，以后再实现。
2. 存在的优化点：可以考虑与本地缓存和消息队列结合，我的想法是：（后续可改进的点）
    1. 通过本地缓存存储频繁 TopN，减少对 Redis 和数据库的访问次数。可以显著降低延迟，提高数据读取速度。
    2. 通过消息队列处理点赞事件，并且通过消息队列处理数据的更新，可以在高并发的场景下，提高性能。
3. 设置了 const MaxAllowedN = 100 // 假设100是业务上可接受的最大值
4. 当前我的数据库中文章数量并不多，所以当前性能测试的意义并不大，有空我一定会去做好这一块。作业还有很多，我要去卷了。