local assert = assert
local error = error
local setmetatable = setmetatable
local string_format = string.format
local type = type

local new = function(cls, ...)
    local instance = {}
    setmetatable(instance, { __index = cls })
    instance.Class = cls
    instance:Constructor(...)
    return instance
end

function gbc.Class(className, superClass)
    assert(type(className) == "string", string_format("gbc.class() - invalid class name \"%s\"", tostring(className)))

    -- create class
    local cls
    cls = { __cname = className, New = function(...)
        return new(cls, ...)
    end }

    -- set super class
    local superType = type(superClass)
    if superType == "table" then
        assert(type(superClass.__cname) == "string", string_format("gbc.class() - create class \"%s\" used super class isn't declared by gbc.class()", className))
        cls.super = superClass
        setmetatable(cls, { __index = cls.super })
    elseif superType ~= "nil" then
        error(string_format("gbc.class() - create class \"%s\" with invalid super type \"%s\"", className, superType))
    end

    if not cls.Constructor then
        cls.Constructor = function() end -- add default constructor
    end

    return cls
end

function gbc.Handler(target, method)
    return function(...)
        return method(target, ...)
    end
end
