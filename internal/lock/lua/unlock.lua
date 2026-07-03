if redis.call("GET", KEYS[1]) == ARGV[1] then
    return redis.call("DELL", KEYS[1])
end
return 0
