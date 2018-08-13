local debug_getlocal = debug.getlocal
local string_byte = string.byte
local string_find = string.find
local string_format = string.format
local string_sub = string.sub

gbc = gbc or {}

socket = {} -- avoid require("socket") warning

-- export global variable
local _g = _G
gbc.exports = {}
setmetatable(gbc.exports, {
    __newindex = function(_, name, value)
        rawset(_g, name, value)
    end,

    __index = function(_, name)
        return rawget(_g, name)
    end
})

-- disable create unexpected global variable
setmetatable(_g, {
    __newindex = function(_, name, value)
        local msg = string_format("USE \"gbc.exports.%s = <value>\" INSTEAD OF SET GLOBAL VARIABLE", name)
        print(debug.traceback(msg, 2))
        if not ngx then print("") end
    end
})

--

gbc.DEBUG_ERROR = 0
gbc.DEBUG_WARN = 1
gbc.DEBUG_INFO = 2
gbc.DEBUG_VERBOSE = 3
gbc.DEBUG = gbc.DEBUG_DEBUG

local _loaded = {}
-- loader
function gbc.import(name, current)
    local _name = name
    local first = string_byte(name)
    if first ~= 46 and _loaded[name] then
        return _loaded[name]
    end

    if first == 35 --[[ "#" ]] then
        name = string_sub(name, 2)
        name = string_format("packages.%s.%s", name, name)
    end

    if first ~= 46 --[[ "." ]] then
        _loaded[_name] = require(name)
        return _loaded[_name]
    end

    if not current then
        local _, v = debug_getlocal(3, 1)
        current = v
    end

    _name = current .. name
    if not _loaded[_name] then
        local pos = string_find(current, "%.[^%.]*$")
        if pos then
            current = string_sub(current, 1, pos - 1)
        end

        _loaded[_name] = require(current .. name)
    end
    return _loaded[_name]
end

-- load basics modules
require("framework.class")
require("framework.table")
require("framework.string")
require("framework.debug")
require("framework.math")
require("framework.ctype")
require("framework.os")
require("framework.io")
