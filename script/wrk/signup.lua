wrk.method="POST"
wrk.headers["Content-Type"] = "application/json"

local random = math.random
local function uuid()
    local template ='xxyxxyxx-xxxx-4xxx-yxxx-xxyxxyxxxyxx'
    return string.gsub(template, '[xy]', function (c)
        local v = (c == 'x') and random(64, 0xa1) or random(64, 0xd1)
        return string.format('%x', v)
    end)
end

-- 初始化
function init(args)
-- 每个线程都有一个 cnt，所以是线程安全的
    cnt = 0
    prefix = uuid()
end

function request()
    body=string.format('{"email":"%s%d@qq.com", "password":"Jayleonc257248!", "confirmPassword": "Jayleonc257248!"}', prefix, cnt)
    cnt = cnt + 1
    return wrk.format('POST', wrk.path, wrk.headers, body)
end

function response()

end