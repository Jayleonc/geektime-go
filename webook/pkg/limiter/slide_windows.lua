local key = KEYS[1]

local windows = tonumber(ARGV[1])

local threshold = tonumber(ARGV[2])

local now = tonumber(ARGV[3])

local min = now - windows

redis.call('ZREMRANGEBYSCORE', key, '-inf', min)

local cnt = redis.call('ZCOUNT', key, '-inf', '+inf')

if cnt >= threshold then
    return "true"
else
    redis.call('ZADD', key, now, now)
    redis.call('PEXPIRE', key, windows)
    return "false"
end