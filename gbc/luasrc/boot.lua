if type(WORKER) == "table" then
    package.path = string.format("%s/?.lua;./?.lua", WORKER.BOOT_DIR)
else
    error("ERR: NOT FOUND GLOBAL VAR 'WORKER'")
end

require("stdlib.ext")
require("gbc.init")

local function main()
    --- @type gbc.MessageHandler
    local handler = gbc.MessageHandler.New(WORKER.ID, WORKER.WORKER_CHAN, WORKER.INPUT_CHAN, WORKER.OUTPUT_CHAN)
    handler:StartLoop()
end

--

xpcall(main, function(err)
    print(err)
    print(debug.traceback())
end)
