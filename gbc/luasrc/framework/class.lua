local assert = assert
local error = error
local setmetatable = setmetatable
local string_format = string.format
local type = type

local _new = function(cls, ...)
    local instance = {}
    setmetatable(instance, { __index = cls })
    instance.class = cls
    instance:ctor(...)
    return instance
end

function gbc.class(classname, super)
    assert(type(classname) == "string", string_format("gbc.class() - invalid class name \"%s\"", tostring(classname)))

    -- create class
    local cls
    cls = { __cname = classname, new = function(...)
        return _new(cls, ...)
    end}

    -- set super class
    local superType = type(super)
    if superType == "table" then
        assert(type(super.__cname) == "string", string_format("gbc.class() - create class \"%s\" used super class isn't declared by gbc.class()", classname))
        cls.super = super
        setmetatable(cls, { __index = cls.super })
    elseif superType ~= "nil" then
        error(string_format("gbc.class() - create class \"%s\" with invalid super type \"%s\"", classname, superType))
    end

    if not cls.ctor then
        cls.ctor = function() end -- add default constructor
    end

    return cls
end

function gbc.handler(target, method)
    return function(...)
        return method(target, ...)
    end
end
