local math_floor = math.floor

function gbc.checknumber(value, base)
    return tonumber(value, base) or 0
end

function gbc.checkint(value)
    value = tonumber(value) or 0
    return math_floor(value + 0.5)
end

function gbc.checkbool(value)
    return (value ~= nil and value ~= false)
end

function gbc.checktable(value)
    if type(value) ~= "table" then value = {} end
    return value
end
