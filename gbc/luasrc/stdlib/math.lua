local math_ceil = math.ceil
local math_floor = math.floor

function math.round(value)
    value = tonumber(value) or 0
    return math_floor(value + 0.5)
end

function math.trunc(x)
    if x <= 0 then
        return math_ceil(x)
    end
    if math_ceil(x) == x then
        x = math_ceil(x)
    else
        x = math_ceil(x) - 1
    end
    return x
end
