local debug_traceback = debug.traceback
local error = error
local print = print
local string_format = string.format
local string_rep = string.rep
local string_upper = string.upper
local tostring = tostring

function gbc.RaiseError(fmt, ...)
    local msg
    if #{ ... } == 0 then
        msg = fmt
    else
        msg = string_format(fmt, ...)
    end
    if gbc.DEBUG > gbc.DEBUG_WARN then
        error(msg, 2)
    else
        error(msg, 0)
    end
end

local function dumpValue(v)
    if type(v) == "string" then
        v = "\"" .. v .. "\""
    end
    return tostring(v)
end

function gbc.Dump(value, desciption, nesting, _print)
    if type(nesting) ~= "number" then nesting = 3 end
    _print = _print or print

    local lookup = {}
    local result = {}
    local traceback = string.split(debug_traceback("", 2), "\n")
    _print("dump from: " .. string.trim(traceback[2]))

    local function _dump(value, desciption, indent, nest, keylen)
        desciption = desciption or "<var>"
        local spc = ""
        if type(keylen) == "number" then
            spc = string_rep(" ", keylen - string.len(dumpValue(desciption)))
        end
        if type(value) ~= "table" then
            result[#result + 1] = string_format("%s%s%s = %s", indent, dumpValue(desciption), spc, dumpValue(value))
        elseif lookup[tostring(value)] then
            result[#result + 1] = string_format("%s%s%s = *REF*", indent, dumpValue(desciption), spc)
        else
            lookup[tostring(value)] = true
            if nest > nesting then
                result[#result + 1] = string_format("%s%s = *MAX NESTING*", indent, dumpValue(desciption))
            else
                result[#result + 1] = string_format("%s%s = {", indent, dumpValue(desciption))
                local indent2 = indent .. "    "
                local keys = {}
                local keylen = 0
                local values = {}
                for k, v in pairs(value) do
                    keys[#keys + 1] = k
                    local vk = dumpValue(k)
                    local vkl = string.len(vk)
                    if vkl > keylen then keylen = vkl end
                    values[k] = v
                end
                table.sort(keys, function(a, b)
                    if type(a) == "number" and type(b) == "number" then
                        return a < b
                    else
                        return tostring(a) < tostring(b)
                    end
                end)
                for i, k in ipairs(keys) do
                    _dump(values[k], k, indent2, nest + 1, keylen)
                end
                result[#result + 1] = string_format("%s}", indent)
            end
        end
    end
    _dump(value, desciption, "- ", 1)

    for i, line in ipairs(result) do
        _print(line)
    end
end

function gbc.Printf(fmt, ...)
    print(string_format(tostring(fmt), ...))
end

function gbc.PrintLog(tag, fmt, ...)
    fmt = tostring(fmt)
    local t = {
        "[",
        string_upper(tostring(tag)),
        "] ",
        string_format(fmt, ...)
    }
    if tag == "ERR" then
        table_insert(t, debug_traceback("", 2))
    end
    print(table.concat(t))
end

local printLog = gbc.PrintLog

function gbc.PrintError(fmt, ...)
    printLog("ERR", fmt, ...)
end

function gbc.PrintDebug(fmt, ...)
    if gbc.DEBUG >= gbc.DEBUG_VERBOSE then
        printLog("DEBUG", fmt, ...)
    end
end

function gbc.PrintInfo(fmt, ...)
    if gbc.DEBUG >= gbc.DEBUG_INFO then
        printLog("INFO", fmt, ...)
    end
end

function gbc.PrintWarn(fmt, ...)
    if gbc.DEBUG >= gbc.DEBUG_WARN then
        printLog("WARN", fmt, ...)
    end
end
