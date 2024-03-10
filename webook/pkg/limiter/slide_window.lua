-- 限流对象
local key = KEYS[1]
-- 窗口大小
local window = tonumber(ARGV[1])
-- 阈值
local threshold = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
-- 窗口的起始时间
local min = now - window

-- 通过这行代码，所有分数（时间戳）小于“当前时间减去窗口大小”（即 min）的成员都将从有序集合中移除。
-- 这里的 min 代表滑动窗口的开始时间，'-inf' 表示最小可能分数。
-- 移除了窗口时间范围之前的所有请求记录，使集合中仅保留了当前窗口内的请求。
-- 全称 Remove Range By Score 按分数移除范围
redis.call('ZREMRANGEBYSCORE', key, '-inf', min)
-- 这行代码计算当前有序集合中元素的数量，即当前窗口内的请求总数。
-- '-inf' 和 '+inf' 表示计数范围从最小分数到最大分数，也就是集合中的所有元素。
-- 得到了当前时间窗口内的总请求次数，为接下来的限流判断提供了依据。
local cnt = redis.call('ZCOUNT', key, '-inf', '+inf')

-- 判断当前窗口内的请求次数是否已经超过了设定的阈值 threshold。
if cnt >= threshold then
    -- 表示需要执行限流措施。
    return "true"
else
    -- 把 score(分数) 和 member(成员) 都设置成 now
    -- 这行代码将当前时间戳 now 作为分数和成员值添加到有序集合 key 中。
    -- 每个新请求都以其到达的时间戳作为唯一标识，确保了请求的顺序性和唯一性。
    -- 确保了所有请求都按照到达的时间顺序被记录。
    redis.call('ZADD', key, now, now)
    -- 通过 PEXPIRE 设置了有序集合的过期时间，确保了集合中只保留当前窗口内的数据。
    redis.call('PEXPIRE', key, window)
    return "false"
end